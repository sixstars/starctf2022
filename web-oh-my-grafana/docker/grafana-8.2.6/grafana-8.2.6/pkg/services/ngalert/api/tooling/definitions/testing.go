package definitions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"

	"github.com/grafana/grafana/pkg/services/ngalert/models"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/prometheus/promql"
)

// swagger:route Post /api/v1/rule/test/{Recipient} testing RouteTestRuleConfig
//
// Test rule
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Responses:
//       200: TestRuleResponse

// swagger:route Post /api/v1/eval testing RouteEvalQueries
//
// Test rule
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Responses:
//       200: EvalQueriesResponse

// swagger:parameters RouteTestReceiverConfig
type TestReceiverRequest struct {
	// in:body
	Body ExtendedReceiver
}

// swagger:parameters RouteTestRuleConfig
type TestRuleRequest struct {
	// in:body
	Body TestRulePayload
}

// swagger:model
type TestRulePayload struct {
	// Example: (node_filesystem_avail_bytes{fstype!="",job="integrations/node_exporter"} node_filesystem_size_bytes{fstype!="",job="integrations/node_exporter"} * 100 < 5 and node_filesystem_readonly{fstype!="",job="integrations/node_exporter"} == 0)
	Expr string `json:"expr,omitempty"`
	// GrafanaManagedCondition for grafana alerts
	GrafanaManagedCondition *models.EvalAlertConditionCommand `json:"grafana_condition,omitempty"`
}

// swagger:parameters RouteEvalQueries
type EvalQueriesRequest struct {
	// in:body
	Body EvalQueriesPayload
}

// swagger:model
type EvalQueriesPayload struct {
	Data []models.AlertQuery `json:"data"`
	Now  time.Time           `json:"now"`
}

func (p *TestRulePayload) UnmarshalJSON(b []byte) error {
	type plain TestRulePayload
	if err := json.Unmarshal(b, (*plain)(p)); err != nil {
		return err
	}

	return p.validate()
}

func (p *TestRulePayload) validate() error {
	if p.Expr != "" && p.GrafanaManagedCondition != nil {
		return fmt.Errorf("cannot mix Grafana & Prometheus style expressions")
	}

	if p.Expr == "" && p.GrafanaManagedCondition == nil {
		return fmt.Errorf("missing either Grafana or Prometheus style expressions")
	}

	return nil
}

func (p *TestRulePayload) Type() (backend Backend) {
	if p.Expr != "" {
		return LoTexRulerBackend
	}

	if p.GrafanaManagedCondition != nil {
		return GrafanaBackend
	}

	return
}

// swagger:model
type TestRuleResponse struct {
	Alerts                promql.Vector          `json:"alerts"`
	GrafanaAlertInstances AlertInstancesResponse `json:"grafana_alert_instances"`
}

// swagger:model
type EvalQueriesResponse = backend.QueryDataResponse

// swagger:model
type AlertInstancesResponse struct {
	// Instances is an array of arrow encoded dataframes
	// each frame has a single row, and a column for each instance (alert identified by unique labels) with a boolean value (firing/not firing)
	Instances [][]byte `json:"instances"`
}

// swagger:model
type ExtendedReceiver struct {
	EmailConfigs     config.EmailConfig      `yaml:"email_configs,omitempty" json:"email_configs,omitempty"`
	PagerdutyConfigs config.PagerdutyConfig  `yaml:"pagerduty_configs,omitempty" json:"pagerduty_configs,omitempty"`
	SlackConfigs     config.SlackConfig      `yaml:"slack_configs,omitempty" json:"slack_configs,omitempty"`
	WebhookConfigs   config.WebhookConfig    `yaml:"webhook_configs,omitempty" json:"webhook_configs,omitempty"`
	OpsGenieConfigs  config.OpsGenieConfig   `yaml:"opsgenie_configs,omitempty" json:"opsgenie_configs,omitempty"`
	WechatConfigs    config.WechatConfig     `yaml:"wechat_configs,omitempty" json:"wechat_configs,omitempty"`
	PushoverConfigs  config.PushoverConfig   `yaml:"pushover_configs,omitempty" json:"pushover_configs,omitempty"`
	VictorOpsConfigs config.VictorOpsConfig  `yaml:"victorops_configs,omitempty" json:"victorops_configs,omitempty"`
	GrafanaReceiver  PostableGrafanaReceiver `yaml:"grafana_managed_receiver,omitempty" json:"grafana_managed_receiver,omitempty"`
}

// swagger:model
type Success ResponseDetails

// swagger:model
type SmtpNotEnabled ResponseDetails

// swagger:model
type Failure ResponseDetails

// swagger:model
type ResponseDetails struct {
	Msg string `json:"msg"`
}
