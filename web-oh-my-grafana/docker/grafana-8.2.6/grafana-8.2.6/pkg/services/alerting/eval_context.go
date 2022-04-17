package alerting

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
)

// EvalContext is the context object for an alert evaluation.
type EvalContext struct {
	Firing         bool
	IsTestRun      bool
	IsDebug        bool
	EvalMatches    []*EvalMatch
	Logs           []*ResultLogEntry
	Error          error
	ConditionEvals string
	StartTime      time.Time
	EndTime        time.Time
	Rule           *Rule
	log            log.Logger

	dashboardRef *models.DashboardRef

	ImagePublicURL  string
	ImageOnDiskPath string
	NoDataFound     bool
	PrevAlertState  models.AlertStateType

	RequestValidator models.PluginRequestValidator

	Ctx context.Context
}

// NewEvalContext is the EvalContext constructor.
func NewEvalContext(alertCtx context.Context, rule *Rule, requestValidator models.PluginRequestValidator) *EvalContext {
	return &EvalContext{
		Ctx:              alertCtx,
		StartTime:        time.Now(),
		Rule:             rule,
		Logs:             make([]*ResultLogEntry, 0),
		EvalMatches:      make([]*EvalMatch, 0),
		log:              log.New("alerting.evalContext"),
		PrevAlertState:   rule.State,
		RequestValidator: requestValidator,
	}
}

// StateDescription contains visual information about the alert state.
type StateDescription struct {
	Color string
	Text  string
	Data  string
}

// GetStateModel returns the `StateDescription` based on current state.
func (c *EvalContext) GetStateModel() *StateDescription {
	switch c.Rule.State {
	case models.AlertStateOK:
		return &StateDescription{
			Color: "#36a64f",
			Text:  "OK",
		}
	case models.AlertStateNoData:
		return &StateDescription{
			Color: "#888888",
			Text:  "No Data",
		}
	case models.AlertStateAlerting:
		return &StateDescription{
			Color: "#D63232",
			Text:  "Alerting",
		}
	case models.AlertStateUnknown:
		return &StateDescription{
			Color: "#888888",
			Text:  "Unknown",
		}
	default:
		panic("Unknown rule state for alert " + c.Rule.State)
	}
}

func (c *EvalContext) shouldUpdateAlertState() bool {
	return c.Rule.State != c.PrevAlertState
}

// GetDurationMs returns the duration of the alert evaluation.
func (c *EvalContext) GetDurationMs() float64 {
	return float64(c.EndTime.Nanosecond()-c.StartTime.Nanosecond()) / float64(1000000)
}

// GetNotificationTitle returns the title of the alert rule including alert state.
func (c *EvalContext) GetNotificationTitle() string {
	return "[" + c.GetStateModel().Text + "] " + c.Rule.Name
}

// GetDashboardUID returns the dashboard uid for the alert rule.
func (c *EvalContext) GetDashboardUID() (*models.DashboardRef, error) {
	if c.dashboardRef != nil {
		return c.dashboardRef, nil
	}

	uidQuery := &models.GetDashboardRefByIdQuery{Id: c.Rule.DashboardID}
	if err := bus.DispatchCtx(c.Ctx, uidQuery); err != nil {
		return nil, err
	}

	c.dashboardRef = uidQuery.Result
	return c.dashboardRef, nil
}

const urlFormat = "%s?tab=alert&viewPanel=%d&orgId=%d"

// GetRuleURL returns the url to the dashboard containing the alert.
func (c *EvalContext) GetRuleURL() (string, error) {
	if c.IsTestRun {
		return setting.AppUrl, nil
	}

	ref, err := c.GetDashboardUID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(urlFormat, models.GetFullDashboardUrl(ref.Uid, ref.Slug), c.Rule.PanelID, c.Rule.OrgID), nil
}

// GetNewState returns the new state from the alert rule evaluation.
func (c *EvalContext) GetNewState() models.AlertStateType {
	ns := getNewStateInternal(c)
	if ns != models.AlertStateAlerting || c.Rule.For == 0 {
		return ns
	}

	since := time.Since(c.Rule.LastStateChange)
	if c.PrevAlertState == models.AlertStatePending && since > c.Rule.For {
		return models.AlertStateAlerting
	}

	if c.PrevAlertState == models.AlertStateAlerting {
		return models.AlertStateAlerting
	}

	return models.AlertStatePending
}

func getNewStateInternal(c *EvalContext) models.AlertStateType {
	if c.Error != nil {
		c.log.Error("Alert Rule Result Error",
			"ruleId", c.Rule.ID,
			"name", c.Rule.Name,
			"error", c.Error,
			"changing state to", c.Rule.ExecutionErrorState.ToAlertState())

		if c.Rule.ExecutionErrorState == models.ExecutionErrorKeepState {
			return c.PrevAlertState
		}
		return c.Rule.ExecutionErrorState.ToAlertState()
	}

	if c.Firing {
		return models.AlertStateAlerting
	}

	if c.NoDataFound {
		c.log.Info("Alert Rule returned no data",
			"ruleId", c.Rule.ID,
			"name", c.Rule.Name,
			"changing state to", c.Rule.NoDataState.ToAlertState())

		if c.Rule.NoDataState == models.NoDataKeepState {
			return c.PrevAlertState
		}
		return c.Rule.NoDataState.ToAlertState()
	}

	return models.AlertStateOK
}

// evaluateNotificationTemplateFields will treat the alert evaluation rule's name and message fields as
// templates, and evaluate the templates using data from the alert evaluation's tags
func (c *EvalContext) evaluateNotificationTemplateFields() error {
	if len(c.EvalMatches) < 1 {
		return nil
	}

	templateDataMap, err := buildTemplateDataMap(c.EvalMatches)
	if err != nil {
		return err
	}

	ruleMsg, err := evaluateTemplate(c.Rule.Message, templateDataMap)
	if err != nil {
		return err
	}
	c.Rule.Message = ruleMsg

	ruleName, err := evaluateTemplate(c.Rule.Name, templateDataMap)
	if err != nil {
		return err
	}
	c.Rule.Name = ruleName

	return nil
}

func evaluateTemplate(s string, m map[string]string) (string, error) {
	for k, v := range m {
		re, err := regexp.Compile(fmt.Sprintf(`\${%s}`, regexp.QuoteMeta(k)))
		if err != nil {
			return "", err
		}
		s = re.ReplaceAllString(s, v)
	}

	return s, nil
}

// buildTemplateDataMap builds a map of alert evaluation tag names to a set of associated values (comma separated)
func buildTemplateDataMap(evalMatches []*EvalMatch) (map[string]string, error) {
	var result = map[string]string{}
	for _, match := range evalMatches {
		for tagName, tagValue := range match.Tags {
			// skip duplicate values
			rVal, err := regexp.Compile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(tagValue)))
			if err != nil {
				return nil, err
			}
			rMatch := rVal.FindString(result[tagName])
			if len(rMatch) > 0 {
				continue
			}
			if _, exists := result[tagName]; exists {
				result[tagName] = fmt.Sprintf("%s, %s", result[tagName], tagValue)
			} else {
				result[tagName] = tagValue
			}
		}
	}
	return result, nil
}
