package usagestats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grafana/grafana/pkg/infra/metrics"
	"github.com/grafana/grafana/pkg/models"
)

var usageStatsURL = "https://stats.grafana.org/grafana-usage-report"

type UsageReport struct {
	Version         string                 `json:"version"`
	Metrics         map[string]interface{} `json:"metrics"`
	Os              string                 `json:"os"`
	Arch            string                 `json:"arch"`
	Edition         string                 `json:"edition"`
	HasValidLicense bool                   `json:"hasValidLicense"`
	Packaging       string                 `json:"packaging"`
	UsageStatsId    string                 `json:"usageStatsId"`
}

func (uss *UsageStatsService) GetUsageReport(ctx context.Context) (UsageReport, error) {
	version := strings.ReplaceAll(uss.Cfg.BuildVersion, ".", "_")

	metrics := map[string]interface{}{}

	edition := "oss"
	if uss.Cfg.IsEnterprise {
		edition = "enterprise"
	}
	report := UsageReport{
		Version:      version,
		Metrics:      metrics,
		Os:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		Edition:      edition,
		Packaging:    uss.Cfg.Packaging,
		UsageStatsId: uss.GetUsageStatsId(ctx),
	}

	statsQuery := models.GetSystemStatsQuery{}
	if err := uss.Bus.Dispatch(&statsQuery); err != nil {
		uss.log.Error("Failed to get system stats", "error", err)
		return report, err
	}

	metrics["stats.dashboards.count"] = statsQuery.Result.Dashboards
	metrics["stats.users.count"] = statsQuery.Result.Users
	metrics["stats.admins.count"] = statsQuery.Result.Admins
	metrics["stats.editors.count"] = statsQuery.Result.Editors
	metrics["stats.viewers.count"] = statsQuery.Result.Viewers
	metrics["stats.orgs.count"] = statsQuery.Result.Orgs
	metrics["stats.playlist.count"] = statsQuery.Result.Playlists
	metrics["stats.plugins.apps.count"] = uss.PluginManager.AppCount()
	metrics["stats.plugins.panels.count"] = uss.PluginManager.PanelCount()
	metrics["stats.plugins.datasources.count"] = uss.PluginManager.DataSourceCount()
	metrics["stats.alerts.count"] = statsQuery.Result.Alerts
	metrics["stats.active_users.count"] = statsQuery.Result.ActiveUsers
	metrics["stats.active_admins.count"] = statsQuery.Result.ActiveAdmins
	metrics["stats.active_editors.count"] = statsQuery.Result.ActiveEditors
	metrics["stats.active_viewers.count"] = statsQuery.Result.ActiveViewers
	metrics["stats.active_sessions.count"] = statsQuery.Result.ActiveSessions
	metrics["stats.daily_active_users.count"] = statsQuery.Result.DailyActiveUsers
	metrics["stats.daily_active_admins.count"] = statsQuery.Result.DailyActiveAdmins
	metrics["stats.daily_active_editors.count"] = statsQuery.Result.DailyActiveEditors
	metrics["stats.daily_active_viewers.count"] = statsQuery.Result.DailyActiveViewers
	metrics["stats.daily_active_sessions.count"] = statsQuery.Result.DailyActiveSessions
	metrics["stats.datasources.count"] = statsQuery.Result.Datasources
	metrics["stats.stars.count"] = statsQuery.Result.Stars
	metrics["stats.folders.count"] = statsQuery.Result.Folders
	metrics["stats.dashboard_permissions.count"] = statsQuery.Result.DashboardPermissions
	metrics["stats.folder_permissions.count"] = statsQuery.Result.FolderPermissions
	metrics["stats.provisioned_dashboards.count"] = statsQuery.Result.ProvisionedDashboards
	metrics["stats.snapshots.count"] = statsQuery.Result.Snapshots
	metrics["stats.teams.count"] = statsQuery.Result.Teams
	metrics["stats.total_auth_token.count"] = statsQuery.Result.AuthTokens
	metrics["stats.dashboard_versions.count"] = statsQuery.Result.DashboardVersions
	metrics["stats.annotations.count"] = statsQuery.Result.Annotations
	metrics["stats.alert_rules.count"] = statsQuery.Result.AlertRules
	metrics["stats.library_panels.count"] = statsQuery.Result.LibraryPanels
	metrics["stats.library_variables.count"] = statsQuery.Result.LibraryVariables
	metrics["stats.dashboards_viewers_can_edit.count"] = statsQuery.Result.DashboardsViewersCanEdit
	metrics["stats.dashboards_viewers_can_admin.count"] = statsQuery.Result.DashboardsViewersCanAdmin
	metrics["stats.folders_viewers_can_edit.count"] = statsQuery.Result.FoldersViewersCanEdit
	metrics["stats.folders_viewers_can_admin.count"] = statsQuery.Result.FoldersViewersCanAdmin

	liveUsersAvg := 0
	liveClientsAvg := 0
	if uss.liveStats.sampleCount > 0 {
		liveUsersAvg = uss.liveStats.numUsersSum / uss.liveStats.sampleCount
		liveClientsAvg = uss.liveStats.numClientsSum / uss.liveStats.sampleCount
	}
	metrics["stats.live_samples.count"] = uss.liveStats.sampleCount
	metrics["stats.live_users_max.count"] = uss.liveStats.numUsersMax
	metrics["stats.live_users_min.count"] = uss.liveStats.numUsersMin
	metrics["stats.live_users_avg.count"] = liveUsersAvg
	metrics["stats.live_clients_max.count"] = uss.liveStats.numClientsMax
	metrics["stats.live_clients_min.count"] = uss.liveStats.numClientsMin
	metrics["stats.live_clients_avg.count"] = liveClientsAvg

	ossEditionCount := 1
	enterpriseEditionCount := 0
	if uss.Cfg.IsEnterprise {
		enterpriseEditionCount = 1
		ossEditionCount = 0
	}
	metrics["stats.edition.oss.count"] = ossEditionCount
	metrics["stats.edition.enterprise.count"] = enterpriseEditionCount

	uss.registerExternalMetrics(metrics)

	// must run after registration of external metrics
	if v, ok := metrics["stats.valid_license.count"]; ok {
		report.HasValidLicense = v == 1
	} else {
		metrics["stats.valid_license.count"] = 0
	}

	userCount := statsQuery.Result.Users
	avgAuthTokensPerUser := statsQuery.Result.AuthTokens
	if userCount != 0 {
		avgAuthTokensPerUser /= userCount
	}

	metrics["stats.avg_auth_token_per_user.count"] = avgAuthTokensPerUser

	dsStats := models.GetDataSourceStatsQuery{}
	if err := uss.Bus.Dispatch(&dsStats); err != nil {
		uss.log.Error("Failed to get datasource stats", "error", err)
		return report, err
	}

	// send counters for each data source
	// but ignore any custom data sources
	// as sending that name could be sensitive information
	dsOtherCount := 0
	for _, dsStat := range dsStats.Result {
		if uss.ShouldBeReported(dsStat.Type) {
			metrics["stats.ds."+dsStat.Type+".count"] = dsStat.Count
		} else {
			dsOtherCount += dsStat.Count
		}
	}
	metrics["stats.ds.other.count"] = dsOtherCount

	esDataSourcesQuery := models.GetDataSourcesByTypeQuery{Type: models.DS_ES}
	if err := uss.Bus.Dispatch(&esDataSourcesQuery); err != nil {
		uss.log.Error("Failed to get elasticsearch json data", "error", err)
		return report, err
	}

	for _, data := range esDataSourcesQuery.Result {
		esVersion, err := data.JsonData.Get("esVersion").Int()
		if err != nil {
			continue
		}

		statName := fmt.Sprintf("stats.ds.elasticsearch.v%d.count", esVersion)

		count, _ := metrics[statName].(int64)

		metrics[statName] = count + 1
	}

	metrics["stats.packaging."+uss.Cfg.Packaging+".count"] = 1
	metrics["stats.distributor."+uss.Cfg.ReportingDistributor+".count"] = 1

	// Alerting stats
	alertingUsageStats, err := uss.AlertingUsageStats.QueryUsageStats()
	if err != nil {
		uss.log.Error("Failed to get alerting usage stats", "error", err)
		return report, err
	}

	var addAlertingUsageStats = func(dsType string, usageCount int) {
		metrics[fmt.Sprintf("stats.alerting.ds.%s.count", dsType)] = usageCount
	}

	alertingOtherCount := 0
	for dsType, usageCount := range alertingUsageStats.DatasourceUsage {
		if uss.ShouldBeReported(dsType) {
			addAlertingUsageStats(dsType, usageCount)
		} else {
			alertingOtherCount += usageCount
		}
	}

	addAlertingUsageStats("other", alertingOtherCount)

	// fetch datasource access stats
	dsAccessStats := models.GetDataSourceAccessStatsQuery{}
	if err := uss.Bus.Dispatch(&dsAccessStats); err != nil {
		uss.log.Error("Failed to get datasource access stats", "error", err)
		return report, err
	}

	// send access counters for each data source
	// but ignore any custom data sources
	// as sending that name could be sensitive information
	dsAccessOtherCount := make(map[string]int64)
	for _, dsAccessStat := range dsAccessStats.Result {
		if dsAccessStat.Access == "" {
			continue
		}

		access := strings.ToLower(dsAccessStat.Access)

		if uss.ShouldBeReported(dsAccessStat.Type) {
			metrics["stats.ds_access."+dsAccessStat.Type+"."+access+".count"] = dsAccessStat.Count
		} else {
			old := dsAccessOtherCount[access]
			dsAccessOtherCount[access] = old + dsAccessStat.Count
		}
	}

	for access, count := range dsAccessOtherCount {
		metrics["stats.ds_access.other."+access+".count"] = count
	}

	// get stats about alert notifier usage
	anStats := models.GetAlertNotifierUsageStatsQuery{}
	if err := uss.Bus.DispatchCtx(ctx, &anStats); err != nil {
		uss.log.Error("Failed to get alert notification stats", "error", err)
		return report, err
	}

	for _, stats := range anStats.Result {
		metrics["stats.alert_notifiers."+stats.Type+".count"] = stats.Count
	}

	// Add stats about auth configuration
	authTypes := map[string]bool{}
	authTypes["anonymous"] = uss.Cfg.AnonymousEnabled
	authTypes["basic_auth"] = uss.Cfg.BasicAuthEnabled
	authTypes["ldap"] = uss.Cfg.LDAPEnabled
	authTypes["auth_proxy"] = uss.Cfg.AuthProxyEnabled

	for provider, enabled := range uss.oauthProviders {
		authTypes["oauth_"+provider] = enabled
	}

	for authType, enabled := range authTypes {
		enabledValue := 0
		if enabled {
			enabledValue = 1
		}
		metrics["stats.auth_enabled."+authType+".count"] = enabledValue
	}

	// Get concurrent users stats as histogram
	concurrentUsersStats, err := uss.GetConcurrentUsersStats(ctx)
	if err != nil {
		uss.log.Error("Failed to get concurrent users stats", "error", err)
		return report, err
	}

	// Histogram is cumulative and metric name has a postfix of le_"<upper inclusive bound>"
	metrics["stats.auth_token_per_user_le_3"] = concurrentUsersStats.BucketLE3
	metrics["stats.auth_token_per_user_le_6"] = concurrentUsersStats.BucketLE6
	metrics["stats.auth_token_per_user_le_9"] = concurrentUsersStats.BucketLE9
	metrics["stats.auth_token_per_user_le_12"] = concurrentUsersStats.BucketLE12
	metrics["stats.auth_token_per_user_le_15"] = concurrentUsersStats.BucketLE15
	metrics["stats.auth_token_per_user_le_inf"] = concurrentUsersStats.BucketLEInf

	metrics["stats.uptime"] = int64(time.Since(uss.startTime).Seconds())

	return report, nil
}

func (uss *UsageStatsService) registerExternalMetrics(metrics map[string]interface{}) {
	for _, fn := range uss.externalMetrics {
		fnMetrics, err := fn()
		if err != nil {
			uss.log.Error("Failed to fetch external metrics", "error", err)
			continue
		}

		for name, value := range fnMetrics {
			metrics[name] = value
		}
	}
}

func (uss *UsageStatsService) RegisterMetricsFunc(fn MetricsFunc) {
	uss.externalMetrics = append(uss.externalMetrics, fn)
}

func (uss *UsageStatsService) sendUsageStats(ctx context.Context) error {
	if !uss.Cfg.ReportingEnabled {
		return nil
	}

	uss.log.Debug(fmt.Sprintf("Sending anonymous usage stats to %s", usageStatsURL))

	report, err := uss.GetUsageReport(ctx)
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(report, "", " ")
	if err != nil {
		return err
	}

	data := bytes.NewBuffer(out)
	sendUsageStats(uss, data)
	return nil
}

// sendUsageStats sends usage statistics.
//
// Stubbable by tests.
var sendUsageStats = func(uss *UsageStatsService, data *bytes.Buffer) {
	go func() {
		client := http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(usageStatsURL, "application/json", data)
		if err != nil {
			uss.log.Error("Failed to send usage stats", "err", err)
			return
		}
		if err := resp.Body.Close(); err != nil {
			uss.log.Warn("Failed to close response body", "err", err)
		}
	}()
}

func (uss *UsageStatsService) sampleLiveStats() {
	current := uss.grafanaLive.UsageStats()

	uss.liveStats.sampleCount++
	uss.liveStats.numClientsSum += current.NumClients
	uss.liveStats.numUsersSum += current.NumUsers

	if current.NumClients > uss.liveStats.numClientsMax {
		uss.liveStats.numClientsMax = current.NumClients
	}

	if current.NumClients < uss.liveStats.numClientsMin {
		uss.liveStats.numClientsMin = current.NumClients
	}

	if current.NumUsers > uss.liveStats.numUsersMax {
		uss.liveStats.numUsersMax = current.NumUsers
	}

	if current.NumUsers < uss.liveStats.numUsersMin {
		uss.liveStats.numUsersMin = current.NumUsers
	}
}

func (uss *UsageStatsService) resetLiveStats() {
	uss.liveStats = liveUsageStats{}
}

func (uss *UsageStatsService) updateTotalStats() {
	if !uss.Cfg.MetricsEndpointEnabled || uss.Cfg.MetricsEndpointDisableTotalStats {
		return
	}

	statsQuery := models.GetSystemStatsQuery{}
	if err := uss.Bus.Dispatch(&statsQuery); err != nil {
		uss.log.Error("Failed to get system stats", "error", err)
		return
	}

	metrics.MStatTotalDashboards.Set(float64(statsQuery.Result.Dashboards))
	metrics.MStatTotalFolders.Set(float64(statsQuery.Result.Folders))
	metrics.MStatTotalUsers.Set(float64(statsQuery.Result.Users))
	metrics.MStatActiveUsers.Set(float64(statsQuery.Result.ActiveUsers))
	metrics.MStatTotalPlaylists.Set(float64(statsQuery.Result.Playlists))
	metrics.MStatTotalOrgs.Set(float64(statsQuery.Result.Orgs))
	metrics.StatsTotalViewers.Set(float64(statsQuery.Result.Viewers))
	metrics.StatsTotalActiveViewers.Set(float64(statsQuery.Result.ActiveViewers))
	metrics.StatsTotalEditors.Set(float64(statsQuery.Result.Editors))
	metrics.StatsTotalActiveEditors.Set(float64(statsQuery.Result.ActiveEditors))
	metrics.StatsTotalAdmins.Set(float64(statsQuery.Result.Admins))
	metrics.StatsTotalActiveAdmins.Set(float64(statsQuery.Result.ActiveAdmins))
	metrics.StatsTotalDashboardVersions.Set(float64(statsQuery.Result.DashboardVersions))
	metrics.StatsTotalAnnotations.Set(float64(statsQuery.Result.Annotations))
	metrics.StatsTotalAlertRules.Set(float64(statsQuery.Result.AlertRules))
	metrics.StatsTotalLibraryPanels.Set(float64(statsQuery.Result.LibraryPanels))
	metrics.StatsTotalLibraryVariables.Set(float64(statsQuery.Result.LibraryVariables))

	dsStats := models.GetDataSourceStatsQuery{}
	if err := uss.Bus.Dispatch(&dsStats); err != nil {
		uss.log.Error("Failed to get datasource stats", "error", err)
		return
	}

	for _, dsStat := range dsStats.Result {
		metrics.StatsTotalDataSources.WithLabelValues(dsStat.Type).Set(float64(dsStat.Count))
	}
}

func (uss *UsageStatsService) ShouldBeReported(dsType string) bool {
	ds := uss.PluginManager.GetDataSource(dsType)
	if ds == nil {
		return false
	}

	return ds.Signature.IsValid() || ds.Signature.IsInternal()
}

func (uss *UsageStatsService) GetUsageStatsId(ctx context.Context) string {
	anonId, ok, err := uss.kvStore.Get(ctx, "anonymous_id")
	if err != nil {
		uss.log.Error("Failed to get usage stats id", "error", err)
		return ""
	}

	if ok {
		return anonId
	}

	newId, err := uuid.NewRandom()
	if err != nil {
		uss.log.Error("Failed to generate usage stats id", "error", err)
		return ""
	}

	anonId = newId.String()

	err = uss.kvStore.Set(ctx, "anonymous_id", anonId)
	if err != nil {
		uss.log.Error("Failed to store usage stats id", "error", err)
		return ""
	}

	return anonId
}
