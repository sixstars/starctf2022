package frontendlogging

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sourcemap "github.com/go-sourcemap/sourcemap"

	"github.com/getsentry/sentry-go"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/setting"
)

type sourceMapLocation struct {
	dir      string
	path     string
	pluginID string
}

type sourceMap struct {
	consumer *sourcemap.Consumer
	pluginID string
}

type ReadSourceMapFn func(dir string, path string) ([]byte, error)

func ReadSourceMapFromFS(dir string, path string) ([]byte, error) {
	file, err := http.Dir(dir).Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Error("Failed to close source map file", "err", err)
		}
	}()
	return ioutil.ReadAll(file)
}

type SourceMapStore struct {
	sync.Mutex
	cache         map[string]*sourceMap
	cfg           *setting.Cfg
	readSourceMap ReadSourceMapFn
	pluginManager plugins.Manager
}

func NewSourceMapStore(cfg *setting.Cfg, pluginManager plugins.Manager, readSourceMap ReadSourceMapFn) *SourceMapStore {
	return &SourceMapStore{
		cache:         make(map[string]*sourceMap),
		cfg:           cfg,
		pluginManager: pluginManager,
		readSourceMap: readSourceMap,
	}
}

/* guessSourceMapLocation will attempt to guess location of a source map on fs.
 * it does not read the source file or make any web requests,
 * just assumes that a [source filename].map file might exist in the same dir as the source file
 * and only considers sources coming from grafana core or plugins`
 */
func (store *SourceMapStore) guessSourceMapLocation(sourceURL string) (*sourceMapLocation, error) {
	u, err := url.Parse(sourceURL)
	if err != nil {
		return nil, err
	}

	// determine if source comes from grafana core, locally or CDN, look in public build dir on fs
	if strings.HasPrefix(u.Path, "/public/build/") || (store.cfg.CDNRootURL != nil &&
		strings.HasPrefix(sourceURL, store.cfg.CDNRootURL.String()) && strings.Contains(u.Path, "/public/build/")) {
		pathParts := strings.SplitN(u.Path, "/public/build/", 2)
		if len(pathParts) == 2 {
			return &sourceMapLocation{
				dir:      store.cfg.StaticRootPath,
				path:     filepath.Join("build", pathParts[1]+".map"),
				pluginID: "",
			}, nil
		}
		// if source comes from a plugin, look in plugin dir
	} else if strings.HasPrefix(u.Path, "/public/plugins/") {
		for _, route := range store.pluginManager.StaticRoutes() {
			pluginPrefix := filepath.Join("/public/plugins/", route.PluginId)
			if strings.HasPrefix(u.Path, pluginPrefix) {
				return &sourceMapLocation{
					dir:      route.Directory,
					path:     u.Path[len(pluginPrefix):] + ".map",
					pluginID: route.PluginId,
				}, nil
			}
		}
	}
	return nil, nil
}

func (store *SourceMapStore) getSourceMap(sourceURL string) (*sourceMap, error) {
	store.Lock()
	defer store.Unlock()

	if smap, ok := store.cache[sourceURL]; ok {
		return smap, nil
	}
	sourceMapLocation, err := store.guessSourceMapLocation(sourceURL)
	if err != nil {
		return nil, err
	}
	if sourceMapLocation == nil {
		// Cache nil value for sourceURL, since we want to flag that we couldn't guess the map location and not try again
		store.cache[sourceURL] = nil
		return nil, nil
	}
	path := strings.ReplaceAll(sourceMapLocation.path, "../", "") // just in case
	b, err := store.readSourceMap(sourceMapLocation.dir, path)
	if err != nil {
		if os.IsNotExist(err) {
			// Cache nil value for sourceURL, since we want to flag that it wasn't found in the filesystem and not try again
			store.cache[sourceURL] = nil
			return nil, nil
		}
		return nil, err
	}

	consumer, err := sourcemap.Parse(sourceURL+".map", b)
	if err != nil {
		return nil, err
	}
	smap := &sourceMap{
		consumer: consumer,
		pluginID: sourceMapLocation.pluginID,
	}
	store.cache[sourceURL] = smap
	return smap, nil
}

func (store *SourceMapStore) resolveSourceLocation(frame sentry.Frame) (*sentry.Frame, error) {
	smap, err := store.getSourceMap(frame.Filename)
	if err != nil {
		return nil, err
	}
	if smap == nil {
		return nil, nil
	}
	file, function, line, col, ok := smap.consumer.Source(frame.Lineno, frame.Colno)
	if !ok {
		return nil, nil
	}
	// unfortunately in many cases go-sourcemap fails to determine the original function name.
	// not a big issue as long as file, line and column are correct
	if len(function) == 0 {
		function = "?"
	}
	module := "core"
	if len(smap.pluginID) > 0 {
		module = smap.pluginID
	}
	return &sentry.Frame{
		Filename: file,
		Lineno:   line,
		Colno:    col,
		Function: function,
		Module:   module,
	}, nil
}
