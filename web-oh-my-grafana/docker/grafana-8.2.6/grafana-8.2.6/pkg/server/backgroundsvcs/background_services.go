package backgroundsvcs

import (
	"github.com/grafana/grafana/pkg/api"
	"github.com/grafana/grafana/pkg/infra/metrics"
	"github.com/grafana/grafana/pkg/infra/remotecache"
	"github.com/grafana/grafana/pkg/infra/tracing"
	"github.com/grafana/grafana/pkg/infra/usagestats"
	"github.com/grafana/grafana/pkg/models"
	backendmanager "github.com/grafana/grafana/pkg/plugins/backendplugin/manager"
	"github.com/grafana/grafana/pkg/plugins/manager"
	"github.com/grafana/grafana/pkg/plugins/plugindashboards"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/services/alerting"
	"github.com/grafana/grafana/pkg/services/cleanup"
	"github.com/grafana/grafana/pkg/services/dashboardsnapshots"
	"github.com/grafana/grafana/pkg/services/live"
	"github.com/grafana/grafana/pkg/services/live/pushhttp"
	"github.com/grafana/grafana/pkg/services/ngalert"
	"github.com/grafana/grafana/pkg/services/notifications"
	"github.com/grafana/grafana/pkg/services/provisioning"
	"github.com/grafana/grafana/pkg/services/rendering"
	"github.com/grafana/grafana/pkg/tsdb/azuremonitor"
	"github.com/grafana/grafana/pkg/tsdb/cloudwatch"
	"github.com/grafana/grafana/pkg/tsdb/elasticsearch"
	"github.com/grafana/grafana/pkg/tsdb/grafanads"
	"github.com/grafana/grafana/pkg/tsdb/graphite"
	"github.com/grafana/grafana/pkg/tsdb/influxdb"
	"github.com/grafana/grafana/pkg/tsdb/loki"
	"github.com/grafana/grafana/pkg/tsdb/mssql"
	"github.com/grafana/grafana/pkg/tsdb/mysql"
	"github.com/grafana/grafana/pkg/tsdb/opentsdb"
	"github.com/grafana/grafana/pkg/tsdb/postgres"
	"github.com/grafana/grafana/pkg/tsdb/prometheus"
	"github.com/grafana/grafana/pkg/tsdb/tempo"
	"github.com/grafana/grafana/pkg/tsdb/testdatasource"
)

func ProvideBackgroundServiceRegistry(
	httpServer *api.HTTPServer, ng *ngalert.AlertNG, cleanup *cleanup.CleanUpService,
	live *live.GrafanaLive, pushGateway *pushhttp.Gateway, notifications *notifications.NotificationService,
	rendering *rendering.RenderingService, tokenService models.UserTokenBackgroundService,
	provisioning *provisioning.ProvisioningServiceImpl, alerting *alerting.AlertEngine, pm *manager.PluginManager,
	backendPM *backendmanager.Manager, metrics *metrics.InternalMetricsService,
	usageStats *usagestats.UsageStatsService, tracing *tracing.TracingService, remoteCache *remotecache.RemoteCache,
	// Need to make sure these are initialized, is there a better place to put them?
	_ *azuremonitor.Service, _ *cloudwatch.CloudWatchService, _ *elasticsearch.Service, _ *graphite.Service,
	_ *influxdb.Service, _ *loki.Service, _ *opentsdb.Service, _ *prometheus.Service, _ *tempo.Service,
	_ *testdatasource.TestDataPlugin, _ *plugindashboards.Service, _ *dashboardsnapshots.Service,
	_ *postgres.Service, _ *mysql.Service, _ *mssql.Service, _ *grafanads.Service,

) *BackgroundServiceRegistry {
	return NewBackgroundServiceRegistry(
		httpServer,
		ng,
		cleanup,
		live,
		pushGateway,
		notifications,
		rendering,
		tokenService,
		provisioning,
		alerting,
		pm,
		backendPM,
		metrics,
		usageStats,
		tracing,
		remoteCache)
}

// BackgroundServiceRegistry provides background services.
type BackgroundServiceRegistry struct {
	Services []registry.BackgroundService
}

func NewBackgroundServiceRegistry(services ...registry.BackgroundService) *BackgroundServiceRegistry {
	return &BackgroundServiceRegistry{services}
}

func (r *BackgroundServiceRegistry) GetServices() []registry.BackgroundService {
	return r.Services
}
