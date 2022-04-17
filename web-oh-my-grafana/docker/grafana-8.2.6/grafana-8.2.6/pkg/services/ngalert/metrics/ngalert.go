package metrics

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/models"
	apimodels "github.com/grafana/grafana/pkg/services/ngalert/api/tooling/definitions"

	"github.com/prometheus/alertmanager/api/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gopkg.in/macaron.v1"
)

const (
	GrafanaBackend = "grafana"
	ProxyBackend   = "proxy"
	Namespace      = "grafana"
	Subsystem      = "alerting"
)

// ProvideService is a Metrics factory.
func ProvideService() *NGAlert {
	return NewNGAlert(prometheus.DefaultRegisterer)
}

// ProvideServiceForTest is a Metrics factory used for test.
func ProvideServiceForTest() *NGAlert {
	return NewNGAlert(prometheus.NewRegistry())
}

type NGAlert struct {
	// Registerer is for use by subcomponents which register their own metrics.
	Registerer                  prometheus.Registerer
	schedulerMetrics            *Scheduler
	stateMetrics                *State
	multiOrgAlertmanagerMetrics *MultiOrgAlertmanager
	apiMetrics                  *API
}

type Scheduler struct {
	Registerer   prometheus.Registerer
	EvalTotal    *prometheus.CounterVec
	EvalFailures *prometheus.CounterVec
	EvalDuration *prometheus.SummaryVec
}

type MultiOrgAlertmanager struct {
	Registerer               prometheus.Registerer
	ActiveConfigurations     prometheus.Gauge
	DiscoveredConfigurations prometheus.Gauge
	registries               *OrgRegistries
}

type API struct {
	RequestDuration *prometheus.HistogramVec
}

type Alertmanager struct {
	Registerer prometheus.Registerer
	*metrics.Alerts
}

type State struct {
	GroupRules *prometheus.GaugeVec
	AlertState *prometheus.GaugeVec
}

func (ng *NGAlert) GetSchedulerMetrics() *Scheduler {
	return ng.schedulerMetrics
}

func (ng *NGAlert) GetStateMetrics() *State {
	return ng.stateMetrics
}

func (ng *NGAlert) GetAPIMetrics() *API {
	return ng.apiMetrics
}

func (ng *NGAlert) GetMultiOrgAlertmanagerMetrics() *MultiOrgAlertmanager {
	return ng.multiOrgAlertmanagerMetrics
}

// NewNGAlert manages the metrics of all the alerting components.
func NewNGAlert(r prometheus.Registerer) *NGAlert {
	return &NGAlert{
		Registerer:                  r,
		schedulerMetrics:            newSchedulerMetrics(r),
		stateMetrics:                newStateMetrics(r),
		multiOrgAlertmanagerMetrics: newMultiOrgAlertmanagerMetrics(r),
		apiMetrics:                  newAPIMetrics(r),
	}
}

// NewAlertmanagerMetrics creates a set of metrics for the Alertmanager of each organization.
func NewAlertmanagerMetrics(r prometheus.Registerer) *Alertmanager {
	return &Alertmanager{
		Registerer: r,
		Alerts:     metrics.NewAlerts("grafana", prometheus.WrapRegistererWithPrefix(fmt.Sprintf("%s_%s_", Namespace, Subsystem), r)),
	}
}

// RemoveOrgRegistry removes the *prometheus.Registry for the specified org. It is safe to call concurrently.
func (moa *MultiOrgAlertmanager) RemoveOrgRegistry(id int64) {
	moa.registries.RemoveOrgRegistry(id)
}

// GetOrCreateOrgRegistry gets or creates a *prometheus.Registry for the specified org. It is safe to call concurrently.
func (moa *MultiOrgAlertmanager) GetOrCreateOrgRegistry(id int64) prometheus.Registerer {
	return moa.registries.GetOrCreateOrgRegistry(id)
}

func newSchedulerMetrics(r prometheus.Registerer) *Scheduler {
	return &Scheduler{
		Registerer: r,
		// TODO: once rule groups support multiple rules, consider partitioning
		// on rule group as well as tenant, similar to loki|cortex.
		EvalTotal: promauto.With(r).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: Subsystem,
				Name:      "rule_evaluations_total",
				Help:      "The total number of rule evaluations.",
			},
			[]string{"org"},
		),
		// TODO: once rule groups support multiple rules, consider partitioning
		// on rule group as well as tenant, similar to loki|cortex.
		EvalFailures: promauto.With(r).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: Subsystem,
				Name:      "rule_evaluation_failures_total",
				Help:      "The total number of rule evaluation failures.",
			},
			[]string{"org"},
		),
		EvalDuration: promauto.With(r).NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  Namespace,
				Subsystem:  Subsystem,
				Name:       "rule_evaluation_duration_seconds",
				Help:       "The duration for a rule to execute.",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"org"},
		),
	}
}

func newStateMetrics(r prometheus.Registerer) *State {
	return &State{
		// TODO: once rule groups support multiple rules, consider partitioning
		// on rule group as well as tenant, similar to loki|cortex.
		GroupRules: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: Namespace,
				Subsystem: Subsystem,
				Name:      "rule_group_rules",
				Help:      "The number of rules.",
			},
			[]string{"org"},
		),
		AlertState: promauto.With(r).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      "alerts",
			Help:      "How many alerts by state.",
		}, []string{"state"}),
	}
}

func newMultiOrgAlertmanagerMetrics(r prometheus.Registerer) *MultiOrgAlertmanager {
	return &MultiOrgAlertmanager{
		Registerer: r,
		registries: NewOrgRegistries(),
		DiscoveredConfigurations: promauto.With(r).NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      "discovered_configurations",
			Help:      "The number of organizations we've discovered that require an Alertmanager configuration.",
		}),
		ActiveConfigurations: promauto.With(r).NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      "active_configurations",
			Help:      "The number of active Alertmanager configurations.",
		}),
	}
}

func newAPIMetrics(r prometheus.Registerer) *API {
	return &API{
		RequestDuration: promauto.With(r).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: Namespace,
				Subsystem: Subsystem,
				Name:      "request_duration_seconds",
				Help:      "Histogram of requests to the Alerting API",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "route", "status_code", "backend"},
		),
	}
}

// OrgRegistries represents a map of registries per org.
type OrgRegistries struct {
	regsMu sync.Mutex
	regs   map[int64]prometheus.Registerer
}

func NewOrgRegistries() *OrgRegistries {
	return &OrgRegistries{
		regs: make(map[int64]prometheus.Registerer),
	}
}

// GetOrCreateOrgRegistry gets or creates a *prometheus.Registry for the specified org. It is safe to call concurrently.
func (m *OrgRegistries) GetOrCreateOrgRegistry(orgID int64) prometheus.Registerer {
	m.regsMu.Lock()
	defer m.regsMu.Unlock()

	orgRegistry, ok := m.regs[orgID]
	if !ok {
		reg := prometheus.NewRegistry()
		m.regs[orgID] = reg
		return reg
	}
	return orgRegistry
}

// RemoveOrgRegistry removes the *prometheus.Registry for the specified org. It is safe to call concurrently.
func (m *OrgRegistries) RemoveOrgRegistry(org int64) {
	m.regsMu.Lock()
	defer m.regsMu.Unlock()
	delete(m.regs, org)
}

// Instrument wraps a middleware, instrumenting the request latencies.
func Instrument(
	method,
	path string,
	action interface{},
	metrics *API,
) macaron.Handler {
	normalizedPath := MakeLabelValue(path)

	return func(c *models.ReqContext) {
		start := time.Now()
		var res response.Response
		val, err := c.Invoke(action)
		if err == nil && val != nil && len(val) > 0 {
			res = val[0].Interface().(response.Response)
		} else {
			res = routing.ServerError(err)
		}

		// TODO: We could look up the datasource type via our datasource service
		var backend string
		recipient := macaron.Params(c.Req)[":Recipient"]
		if recipient == apimodels.GrafanaBackend.String() || recipient == "" {
			backend = GrafanaBackend
		} else {
			backend = ProxyBackend
		}

		ls := prometheus.Labels{
			"method":      method,
			"route":       normalizedPath,
			"status_code": fmt.Sprint(res.Status()),
			"backend":     backend,
		}
		res.WriteTo(c)
		metrics.RequestDuration.With(ls).Observe(time.Since(start).Seconds())
	}
}

var invalidChars = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// MakeLabelValue normalizes a path template
func MakeLabelValue(path string) string {
	// Convert non-alnums to underscores.
	result := invalidChars.ReplaceAllString(path, "_")

	// Trim leading and trailing underscores.
	result = strings.Trim(result, "_")

	// Make it all lowercase
	result = strings.ToLower(result)

	// Special case.
	if result == "" {
		result = "root"
	}
	return result
}
