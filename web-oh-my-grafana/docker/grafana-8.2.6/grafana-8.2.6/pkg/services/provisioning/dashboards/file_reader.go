package dashboards

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	dboards "github.com/grafana/grafana/pkg/dashboards"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/dashboards"
	"github.com/grafana/grafana/pkg/util"
)

var (
	// ErrFolderNameMissing is returned when folder name is missing.
	ErrFolderNameMissing = errors.New("folder name missing")
)

// FileReader is responsible for reading dashboards from disk and
// insert/update dashboards to the Grafana database using
// `dashboards.DashboardProvisioningService`.
type FileReader struct {
	Cfg                          *config
	Path                         string
	log                          log.Logger
	dashboardProvisioningService dashboards.DashboardProvisioningService
	FoldersFromFilesStructure    bool

	mux                     sync.RWMutex
	usageTracker            *usageTracker
	dbWriteAccessRestricted bool
}

// NewDashboardFileReader returns a new filereader based on `config`
func NewDashboardFileReader(cfg *config, log log.Logger, store dboards.Store) (*FileReader, error) {
	var path string
	path, ok := cfg.Options["path"].(string)
	if !ok {
		path, ok = cfg.Options["folder"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to load dashboards, path param is not a string")
		}

		log.Warn("[Deprecated] The folder property is deprecated. Please use path instead.")
	}

	foldersFromFilesStructure, _ := cfg.Options["foldersFromFilesStructure"].(bool)
	if foldersFromFilesStructure && cfg.Folder != "" && cfg.FolderUID != "" {
		return nil, fmt.Errorf("'folder' and 'folderUID' should be empty using 'foldersFromFilesStructure' option")
	}

	return &FileReader{
		Cfg:                          cfg,
		Path:                         path,
		log:                          log,
		dashboardProvisioningService: dashboards.NewProvisioningService(store),
		FoldersFromFilesStructure:    foldersFromFilesStructure,
		usageTracker:                 newUsageTracker(),
	}, nil
}

// pollChanges periodically runs walkDisk based on interval specified in the config.
func (fr *FileReader) pollChanges(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(int64(time.Second) * fr.Cfg.UpdateIntervalSeconds))
	for {
		select {
		case <-ticker.C:
			if err := fr.walkDisk(ctx); err != nil {
				fr.log.Error("failed to search for dashboards", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// walkDisk traverses the file system for the defined path, reading dashboard definition files,
// and applies any change to the database.
func (fr *FileReader) walkDisk(ctx context.Context) error {
	fr.log.Debug("Start walking disk", "path", fr.Path)
	resolvedPath := fr.resolvedPath()
	if _, err := os.Stat(resolvedPath); err != nil {
		return err
	}

	provisionedDashboardRefs, err := getProvisionedDashboardsByPath(fr.dashboardProvisioningService, fr.Cfg.Name)
	if err != nil {
		return err
	}

	// Find relevant files
	filesFoundOnDisk := map[string]os.FileInfo{}
	if err := filepath.Walk(resolvedPath, createWalkFn(filesFoundOnDisk)); err != nil {
		return err
	}

	fr.handleMissingDashboardFiles(provisionedDashboardRefs, filesFoundOnDisk)

	usageTracker := newUsageTracker()
	if fr.FoldersFromFilesStructure {
		err = fr.storeDashboardsInFoldersFromFileStructure(ctx, filesFoundOnDisk, provisionedDashboardRefs, resolvedPath, usageTracker)
	} else {
		err = fr.storeDashboardsInFolder(ctx, filesFoundOnDisk, provisionedDashboardRefs, usageTracker)
	}
	if err != nil {
		return err
	}

	fr.mux.Lock()
	defer fr.mux.Unlock()

	fr.usageTracker = usageTracker
	return nil
}

func (fr *FileReader) changeWritePermissions(restrict bool) {
	fr.mux.Lock()
	defer fr.mux.Unlock()

	fr.dbWriteAccessRestricted = restrict
}

func (fr *FileReader) isDatabaseAccessRestricted() bool {
	fr.mux.RLock()
	defer fr.mux.RUnlock()

	return fr.dbWriteAccessRestricted
}

// storeDashboardsInFolder saves dashboards from the filesystem on disk to the folder from config
func (fr *FileReader) storeDashboardsInFolder(ctx context.Context, filesFoundOnDisk map[string]os.FileInfo,
	dashboardRefs map[string]*models.DashboardProvisioning, usageTracker *usageTracker) error {
	folderID, err := getOrCreateFolderID(ctx, fr.Cfg, fr.dashboardProvisioningService, fr.Cfg.Folder)
	if err != nil && !errors.Is(err, ErrFolderNameMissing) {
		return err
	}

	// save dashboards based on json files
	for path, fileInfo := range filesFoundOnDisk {
		provisioningMetadata, err := fr.saveDashboard(path, folderID, fileInfo, dashboardRefs)
		if err != nil {
			fr.log.Error("failed to save dashboard", "error", err)
			continue
		}

		usageTracker.track(provisioningMetadata)
	}
	return nil
}

// storeDashboardsInFoldersFromFilesystemStructure saves dashboards from the filesystem on disk to the same folder
// in Grafana as they are in on the filesystem.
func (fr *FileReader) storeDashboardsInFoldersFromFileStructure(ctx context.Context, filesFoundOnDisk map[string]os.FileInfo,
	dashboardRefs map[string]*models.DashboardProvisioning, resolvedPath string, usageTracker *usageTracker) error {
	for path, fileInfo := range filesFoundOnDisk {
		folderName := ""

		dashboardsFolder := filepath.Dir(path)
		if dashboardsFolder != resolvedPath {
			folderName = filepath.Base(dashboardsFolder)
		}

		folderID, err := getOrCreateFolderID(ctx, fr.Cfg, fr.dashboardProvisioningService, folderName)
		if err != nil && !errors.Is(err, ErrFolderNameMissing) {
			return fmt.Errorf("can't provision folder %q from file system structure: %w", folderName, err)
		}

		provisioningMetadata, err := fr.saveDashboard(path, folderID, fileInfo, dashboardRefs)
		usageTracker.track(provisioningMetadata)
		if err != nil {
			fr.log.Error("failed to save dashboard", "error", err)
		}
	}
	return nil
}

// handleMissingDashboardFiles will unprovision or delete dashboards which are missing on disk.
func (fr *FileReader) handleMissingDashboardFiles(provisionedDashboardRefs map[string]*models.DashboardProvisioning,
	filesFoundOnDisk map[string]os.FileInfo) {
	// find dashboards to delete since json file is missing
	var dashboardsToDelete []int64
	for path, provisioningData := range provisionedDashboardRefs {
		_, existsOnDisk := filesFoundOnDisk[path]
		if !existsOnDisk {
			dashboardsToDelete = append(dashboardsToDelete, provisioningData.DashboardId)
		}
	}

	if fr.Cfg.DisableDeletion {
		// If deletion is disabled for the provisioner we just remove provisioning metadata about the dashboard
		// so afterwards the dashboard is considered unprovisioned.
		for _, dashboardID := range dashboardsToDelete {
			fr.log.Debug("unprovisioning provisioned dashboard. missing on disk", "id", dashboardID)
			err := fr.dashboardProvisioningService.UnprovisionDashboard(dashboardID)
			if err != nil {
				fr.log.Error("failed to unprovision dashboard", "dashboard_id", dashboardID, "error", err)
			}
		}
	} else {
		// delete dashboards missing JSON file
		for _, dashboardID := range dashboardsToDelete {
			fr.log.Debug("deleting provisioned dashboard, missing on disk", "id", dashboardID)
			err := fr.dashboardProvisioningService.DeleteProvisionedDashboard(dashboardID, fr.Cfg.OrgID)
			if err != nil {
				fr.log.Error("failed to delete dashboard", "id", dashboardID, "error", err)
			}
		}
	}
}

// saveDashboard saves or updates the dashboard provisioning file at path.
func (fr *FileReader) saveDashboard(path string, folderID int64, fileInfo os.FileInfo,
	provisionedDashboardRefs map[string]*models.DashboardProvisioning) (provisioningMetadata, error) {
	provisioningMetadata := provisioningMetadata{}
	resolvedFileInfo, err := resolveSymlink(fileInfo, path)
	if err != nil {
		return provisioningMetadata, err
	}

	provisionedData, alreadyProvisioned := provisionedDashboardRefs[path]

	jsonFile, err := fr.readDashboardFromFile(path, resolvedFileInfo.ModTime(), folderID)
	if err != nil {
		fr.log.Error("failed to load dashboard from ", "file", path, "error", err)
		return provisioningMetadata, nil
	}

	upToDate := alreadyProvisioned
	if provisionedData != nil {
		upToDate = jsonFile.checkSum == provisionedData.CheckSum
	}

	// keeps track of which UIDs and titles we have already provisioned
	dash := jsonFile.dashboard
	provisioningMetadata.uid = dash.Dashboard.Uid
	provisioningMetadata.identity = dashboardIdentity{title: dash.Dashboard.Title, folderID: dash.Dashboard.FolderId}

	if upToDate {
		return provisioningMetadata, nil
	}

	if dash.Dashboard.Id != 0 {
		dash.Dashboard.Data.Set("id", nil)
		dash.Dashboard.Id = 0
	}

	if alreadyProvisioned {
		dash.Dashboard.SetId(provisionedData.DashboardId)
	}

	if !fr.isDatabaseAccessRestricted() {
		fr.log.Debug("saving new dashboard", "provisioner", fr.Cfg.Name, "file", path, "folderId", dash.Dashboard.FolderId)
		dp := &models.DashboardProvisioning{
			ExternalId: path,
			Name:       fr.Cfg.Name,
			Updated:    resolvedFileInfo.ModTime().Unix(),
			CheckSum:   jsonFile.checkSum,
		}
		if _, err := fr.dashboardProvisioningService.SaveProvisionedDashboard(dash, dp); err != nil {
			return provisioningMetadata, err
		}
	} else {
		fr.log.Warn("Not saving new dashboard due to restricted database access", "provisioner", fr.Cfg.Name,
			"file", path, "folderId", dash.Dashboard.FolderId)
	}

	return provisioningMetadata, nil
}

func getProvisionedDashboardsByPath(service dashboards.DashboardProvisioningService, name string) (
	map[string]*models.DashboardProvisioning, error) {
	arr, err := service.GetProvisionedDashboardData(name)
	if err != nil {
		return nil, err
	}

	byPath := map[string]*models.DashboardProvisioning{}
	for _, pd := range arr {
		byPath[pd.ExternalId] = pd
	}

	return byPath, nil
}

func getOrCreateFolderID(ctx context.Context, cfg *config, service dashboards.DashboardProvisioningService, folderName string) (int64, error) {
	if folderName == "" {
		return 0, ErrFolderNameMissing
	}

	cmd := &models.GetDashboardQuery{Slug: models.SlugifyTitle(folderName), OrgId: cfg.OrgID}
	err := bus.DispatchCtx(ctx, cmd)

	if err != nil && !errors.Is(err, models.ErrDashboardNotFound) {
		return 0, err
	}

	// dashboard folder not found. create one.
	if errors.Is(err, models.ErrDashboardNotFound) {
		dash := &dashboards.SaveDashboardDTO{}
		dash.Dashboard = models.NewDashboardFolder(folderName)
		dash.Dashboard.IsFolder = true
		dash.Overwrite = true
		dash.OrgId = cfg.OrgID
		// set dashboard folderUid if given
		dash.Dashboard.SetUid(cfg.FolderUID)
		dbDash, err := service.SaveFolderForProvisionedDashboards(dash)
		if err != nil {
			return 0, err
		}

		return dbDash.Id, nil
	}

	if !cmd.Result.IsFolder {
		return 0, fmt.Errorf("got invalid response. expected folder, found dashboard")
	}

	return cmd.Result.Id, nil
}

func resolveSymlink(fileinfo os.FileInfo, path string) (os.FileInfo, error) {
	checkFilepath, err := filepath.EvalSymlinks(path)
	if path != checkFilepath {
		fi, err := os.Lstat(checkFilepath)
		if err != nil {
			return nil, err
		}

		return fi, nil
	}

	return fileinfo, err
}

func createWalkFn(filesOnDisk map[string]os.FileInfo) filepath.WalkFunc {
	return func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		isValid, err := validateWalkablePath(fileInfo)
		if !isValid {
			return err
		}

		filesOnDisk[path] = fileInfo
		return nil
	}
}

func validateWalkablePath(fileInfo os.FileInfo) (bool, error) {
	if fileInfo.IsDir() {
		if strings.HasPrefix(fileInfo.Name(), ".") {
			return false, filepath.SkipDir
		}
		return false, nil
	}

	if !strings.HasSuffix(fileInfo.Name(), ".json") {
		return false, nil
	}

	return true, nil
}

type dashboardJSONFile struct {
	dashboard    *dashboards.SaveDashboardDTO
	checkSum     string
	lastModified time.Time
}

func (fr *FileReader) readDashboardFromFile(path string, lastModified time.Time, folderID int64) (*dashboardJSONFile, error) {
	// nolint:gosec
	// We can ignore the gosec G304 warning on this one because `path` comes from the provisioning configuration file.
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			fr.log.Warn("Failed to close file", "path", path, "err", err)
		}
	}()

	all, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	checkSum, err := util.Md5SumString(string(all))
	if err != nil {
		return nil, err
	}

	data, err := simplejson.NewJson(all)
	if err != nil {
		return nil, err
	}

	dash, err := createDashboardJSON(data, lastModified, fr.Cfg, folderID)
	if err != nil {
		return nil, err
	}

	return &dashboardJSONFile{
		dashboard:    dash,
		checkSum:     checkSum,
		lastModified: lastModified,
	}, nil
}

func (fr *FileReader) resolvedPath() string {
	if _, err := os.Stat(fr.Path); os.IsNotExist(err) {
		fr.log.Error("Cannot read directory", "error", err)
	}

	path, err := filepath.Abs(fr.Path)
	if err != nil {
		fr.log.Error("Could not create absolute path", "path", fr.Path, "error", err)
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		fr.log.Error("Failed to read content of symlinked path", "path", fr.Path, "error", err)
	}

	if path == "" {
		path = fr.Path
		fr.log.Info("falling back to original path due to EvalSymlink/Abs failure")
	}
	return path
}

func (fr *FileReader) getUsageTracker() *usageTracker {
	fr.mux.RLock()
	defer fr.mux.RUnlock()

	return fr.usageTracker
}

type provisioningMetadata struct {
	uid      string
	identity dashboardIdentity
}

type dashboardIdentity struct {
	folderID int64
	title    string
}

func (d *dashboardIdentity) Exists() bool {
	return len(d.title) > 0
}

func newUsageTracker() *usageTracker {
	return &usageTracker{
		uidUsage:   map[string]uint8{},
		titleUsage: map[dashboardIdentity]uint8{},
	}
}

type usageTracker struct {
	uidUsage   map[string]uint8
	titleUsage map[dashboardIdentity]uint8
}

func (t *usageTracker) track(pm provisioningMetadata) {
	if len(pm.uid) > 0 {
		t.uidUsage[pm.uid]++
	}
	if pm.identity.Exists() {
		t.titleUsage[pm.identity]++
	}
}
