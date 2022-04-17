package ngalert

import (
	"context"
	"net/url"
	"time"

	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/infra/kvstore"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/datasourceproxy"
	"github.com/grafana/grafana/pkg/services/datasources"
	"github.com/grafana/grafana/pkg/services/ngalert/api"
	"github.com/grafana/grafana/pkg/services/ngalert/eval"
	"github.com/grafana/grafana/pkg/services/ngalert/metrics"
	"github.com/grafana/grafana/pkg/services/ngalert/notifier"
	"github.com/grafana/grafana/pkg/services/ngalert/schedule"
	"github.com/grafana/grafana/pkg/services/ngalert/state"
	"github.com/grafana/grafana/pkg/services/ngalert/store"
	"github.com/grafana/grafana/pkg/services/quota"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/tsdb"

	"github.com/benbjohnson/clock"
	"golang.org/x/sync/errgroup"
)

const (
	// scheduler interval
	// changing this value is discouraged
	// because this could cause existing alert definition
	// with intervals that are not exactly divided by this number
	// not to be evaluated
	defaultBaseIntervalSeconds = 10
	// default alert definition interval
	defaultIntervalSeconds int64 = 6 * defaultBaseIntervalSeconds
)

func ProvideService(cfg *setting.Cfg, dataSourceCache datasources.CacheService, routeRegister routing.RouteRegister,
	sqlStore *sqlstore.SQLStore, kvStore kvstore.KVStore, dataService *tsdb.Service, dataProxy *datasourceproxy.DataSourceProxyService,
	quotaService *quota.QuotaService, m *metrics.NGAlert) (*AlertNG, error) {
	ng := &AlertNG{
		Cfg:             cfg,
		DataSourceCache: dataSourceCache,
		RouteRegister:   routeRegister,
		SQLStore:        sqlStore,
		KVStore:         kvStore,
		DataService:     dataService,
		DataProxy:       dataProxy,
		QuotaService:    quotaService,
		Metrics:         m,
		Log:             log.New("ngalert"),
	}

	if ng.IsDisabled() {
		return ng, nil
	}

	if err := ng.init(); err != nil {
		return nil, err
	}

	return ng, nil
}

// AlertNG is the service for evaluating the condition of an alert definition.
type AlertNG struct {
	Cfg             *setting.Cfg
	DataSourceCache datasources.CacheService
	RouteRegister   routing.RouteRegister
	SQLStore        *sqlstore.SQLStore
	KVStore         kvstore.KVStore
	DataService     *tsdb.Service
	DataProxy       *datasourceproxy.DataSourceProxyService
	QuotaService    *quota.QuotaService
	Metrics         *metrics.NGAlert
	Log             log.Logger
	schedule        schedule.ScheduleService
	stateManager    *state.Manager

	// Alerting notification services
	MultiOrgAlertmanager *notifier.MultiOrgAlertmanager
}

func (ng *AlertNG) init() error {
	var err error

	baseInterval := ng.Cfg.AlertingBaseInterval
	if baseInterval <= 0 {
		baseInterval = defaultBaseIntervalSeconds
	}
	baseInterval *= time.Second

	store := &store.DBstore{
		BaseInterval:    baseInterval,
		DefaultInterval: ng.getRuleDefaultInterval(),
		SQLStore:        ng.SQLStore,
		Logger:          ng.Log,
	}

	multiOrgMetrics := ng.Metrics.GetMultiOrgAlertmanagerMetrics()
	ng.MultiOrgAlertmanager, err = notifier.NewMultiOrgAlertmanager(ng.Cfg, store, store, ng.KVStore, multiOrgMetrics, log.New("ngalert.multiorg.alertmanager"))
	if err != nil {
		return err
	}

	// Let's make sure we're able to complete an initial sync of Alertmanagers before we start the alerting components.
	if err := ng.MultiOrgAlertmanager.LoadAndSyncAlertmanagersForOrgs(context.Background()); err != nil {
		return err
	}

	schedCfg := schedule.SchedulerCfg{
		C:                       clock.New(),
		BaseInterval:            baseInterval,
		Logger:                  ng.Log,
		MaxAttempts:             ng.Cfg.UnifiedAlerting.MaxAttempts,
		Evaluator:               eval.Evaluator{Cfg: ng.Cfg, Log: ng.Log},
		InstanceStore:           store,
		RuleStore:               store,
		AdminConfigStore:        store,
		OrgStore:                store,
		MultiOrgNotifier:        ng.MultiOrgAlertmanager,
		Metrics:                 ng.Metrics.GetSchedulerMetrics(),
		AdminConfigPollInterval: ng.Cfg.UnifiedAlerting.AdminConfigPollInterval,
		DisabledOrgs:            ng.Cfg.UnifiedAlerting.DisabledOrgs,
		MinRuleInterval:         ng.getRuleMinInterval(),
	}

	appUrl, err := url.Parse(ng.Cfg.AppURL)
	if err != nil {
		ng.Log.Error("Failed to parse application URL. Continue without it.", "error", err)
		appUrl = nil
	}
	stateManager := state.NewManager(ng.Log, ng.Metrics.GetStateMetrics(), store, store)
	scheduler := schedule.NewScheduler(schedCfg, ng.DataService, appUrl, stateManager)

	ng.stateManager = stateManager
	ng.schedule = scheduler

	api := api.API{
		Cfg:                  ng.Cfg,
		DatasourceCache:      ng.DataSourceCache,
		RouteRegister:        ng.RouteRegister,
		DataService:          ng.DataService,
		Schedule:             ng.schedule,
		DataProxy:            ng.DataProxy,
		QuotaService:         ng.QuotaService,
		InstanceStore:        store,
		RuleStore:            store,
		AlertingStore:        store,
		AdminConfigStore:     store,
		MultiOrgAlertmanager: ng.MultiOrgAlertmanager,
		StateManager:         ng.stateManager,
	}
	api.RegisterAPIEndpoints(ng.Metrics.GetAPIMetrics())

	return nil
}

// Run starts the scheduler and Alertmanager.
func (ng *AlertNG) Run(ctx context.Context) error {
	ng.Log.Debug("ngalert starting")
	ng.stateManager.Warm()

	children, subCtx := errgroup.WithContext(ctx)

	if ng.Cfg.UnifiedAlerting.ExecuteAlerts {
		children.Go(func() error {
			return ng.schedule.Run(subCtx)
		})
	}
	children.Go(func() error {
		return ng.MultiOrgAlertmanager.Run(subCtx)
	})
	return children.Wait()
}

// IsDisabled returns true if the alerting service is disable for this instance.
func (ng *AlertNG) IsDisabled() bool {
	if ng.Cfg == nil {
		return true
	}
	return !ng.Cfg.UnifiedAlerting.Enabled
}

// getRuleDefaultIntervalSeconds returns the default rule interval if the interval is not set.
// If this constant (1 minute) is lower than the configured minimum evaluation interval then
// this configuration is returned.
func (ng *AlertNG) getRuleDefaultInterval() time.Duration {
	ruleMinInterval := ng.getRuleMinInterval()
	if defaultIntervalSeconds < int64(ruleMinInterval.Seconds()) {
		return ruleMinInterval
	}
	return time.Duration(defaultIntervalSeconds) * time.Second
}

// getRuleMinIntervalSeconds returns the configured minimum rule interval.
// If this value is less or equal to zero or not divided exactly by the scheduler interval
// the scheduler interval (10 seconds) is returned.
func (ng *AlertNG) getRuleMinInterval() time.Duration {
	if ng.Cfg.UnifiedAlerting.MinInterval <= 0 {
		return defaultBaseIntervalSeconds // if it's not configured; apply default
	}

	if ng.Cfg.UnifiedAlerting.MinInterval%defaultBaseIntervalSeconds != 0 {
		ng.Log.Error("Configured minimum evaluation interval is not divided exactly by the scheduler interval and it will fallback to default", "alertingMinInterval", ng.Cfg.UnifiedAlerting.MinInterval, "baseIntervalSeconds", defaultBaseIntervalSeconds, "defaultIntervalSeconds", defaultIntervalSeconds)
		return defaultBaseIntervalSeconds // if it's invalid; apply default
	}

	return ng.Cfg.UnifiedAlerting.MinInterval
}
