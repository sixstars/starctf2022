package channels

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
)

const (
	// victoropsAlertStateCritical - Victorops uses "CRITICAL" string to indicate "Alerting" state
	victoropsAlertStateCritical = "CRITICAL"

	// victoropsAlertStateRecovery - VictorOps "RECOVERY" message type
	victoropsAlertStateRecovery = "RECOVERY"
)

// NewVictoropsNotifier creates an instance of VictoropsNotifier that
// handles posting notifications to Victorops REST API
func NewVictoropsNotifier(model *NotificationChannelConfig, t *template.Template) (*VictoropsNotifier, error) {
	url := model.Settings.Get("url").MustString()
	if url == "" {
		return nil, receiverInitError{Cfg: *model, Reason: "could not find victorops url property in settings"}
	}

	return &VictoropsNotifier{
		Base: NewBase(&models.AlertNotification{
			Uid:                   model.UID,
			Name:                  model.Name,
			Type:                  model.Type,
			DisableResolveMessage: model.DisableResolveMessage,
			Settings:              model.Settings,
		}),
		URL:         url,
		MessageType: model.Settings.Get("messageType").MustString(),
		log:         log.New("alerting.notifier.victorops"),
		tmpl:        t,
	}, nil
}

// VictoropsNotifier defines URL property for Victorops REST API
// and handles notification process by formatting POST body according to
// Victorops specifications (http://victorops.force.com/knowledgebase/articles/Integration/Alert-Ingestion-API-Documentation/)
type VictoropsNotifier struct {
	*Base
	URL         string
	MessageType string
	log         log.Logger
	tmpl        *template.Template
}

// Notify sends notification to Victorops via POST to URL endpoint
func (vn *VictoropsNotifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	vn.log.Debug("Executing victorops notification", "notification", vn.Name)

	var tmplErr error
	tmpl, _ := TmplText(ctx, vn.tmpl, as, vn.log, &tmplErr)

	messageType := strings.ToUpper(tmpl(vn.MessageType))
	if messageType == "" {
		messageType = victoropsAlertStateCritical
	}
	alerts := types.Alerts(as...)
	if alerts.Status() == model.AlertResolved {
		messageType = victoropsAlertStateRecovery
	}

	groupKey, err := notify.ExtractGroupKey(ctx)
	if err != nil {
		return false, err
	}

	bodyJSON := simplejson.New()
	bodyJSON.Set("message_type", messageType)
	bodyJSON.Set("entity_id", groupKey.Hash())
	bodyJSON.Set("entity_display_name", tmpl(`{{ template "default.title" . }}`))
	bodyJSON.Set("timestamp", time.Now().Unix())
	bodyJSON.Set("state_message", tmpl(`{{ template "default.message" . }}`))
	bodyJSON.Set("monitoring_tool", "Grafana v"+setting.BuildVersion)

	ruleURL := joinUrlPath(vn.tmpl.ExternalURL.String(), "/alerting/list", vn.log)
	bodyJSON.Set("alert_url", ruleURL)

	u := tmpl(vn.URL)
	if tmplErr != nil {
		vn.log.Debug("failed to template VictorOps message", "err", tmplErr.Error())
	}

	b, err := bodyJSON.MarshalJSON()
	if err != nil {
		return false, err
	}
	cmd := &models.SendWebhookSync{
		Url:  u,
		Body: string(b),
	}

	if err := bus.DispatchCtx(ctx, cmd); err != nil {
		vn.log.Error("Failed to send Victorops notification", "error", err, "webhook", vn.Name)
		return false, err
	}

	return true, nil
}

func (vn *VictoropsNotifier) SendResolved() bool {
	return !vn.GetDisableResolveMessage()
}
