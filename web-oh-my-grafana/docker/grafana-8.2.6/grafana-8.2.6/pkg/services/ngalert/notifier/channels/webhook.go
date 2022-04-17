package channels

import (
	"context"
	"encoding/json"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"

	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
)

// WebhookNotifier is responsible for sending
// alert notifications as webhooks.
type WebhookNotifier struct {
	*Base
	URL        string
	User       string
	Password   string
	HTTPMethod string
	MaxAlerts  int
	log        log.Logger
	tmpl       *template.Template
}

// NewWebHookNotifier is the constructor for
// the WebHook notifier.
func NewWebHookNotifier(model *NotificationChannelConfig, t *template.Template) (*WebhookNotifier, error) {
	if model.Settings == nil {
		return nil, receiverInitError{Cfg: *model, Reason: "could not find settings property"}
	}
	url := model.Settings.Get("url").MustString()
	if url == "" {
		return nil, receiverInitError{Cfg: *model, Reason: "could not find url property in settings"}
	}
	return &WebhookNotifier{
		Base: NewBase(&models.AlertNotification{
			Uid:                   model.UID,
			Name:                  model.Name,
			Type:                  model.Type,
			DisableResolveMessage: model.DisableResolveMessage,
			Settings:              model.Settings,
		}),
		URL:        url,
		User:       model.Settings.Get("username").MustString(),
		Password:   model.DecryptedValue("password", model.Settings.Get("password").MustString()),
		HTTPMethod: model.Settings.Get("httpMethod").MustString("POST"),
		MaxAlerts:  model.Settings.Get("maxAlerts").MustInt(0),
		log:        log.New("alerting.notifier.webhook"),
		tmpl:       t,
	}, nil
}

// webhookMessage defines the JSON object send to webhook endpoints.
type webhookMessage struct {
	*ExtendedData

	// The protocol version.
	Version         string `json:"version"`
	GroupKey        string `json:"groupKey"`
	TruncatedAlerts int    `json:"truncatedAlerts"`

	// Deprecated, to be removed in 8.1.
	// These are present to make migration a little less disruptive.
	Title   string `json:"title"`
	State   string `json:"state"`
	Message string `json:"message"`
}

// Notify implements the Notifier interface.
func (wn *WebhookNotifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	groupKey, err := notify.ExtractGroupKey(ctx)
	if err != nil {
		return false, err
	}

	as, numTruncated := truncateAlerts(wn.MaxAlerts, as)
	var tmplErr error
	tmpl, data := TmplText(ctx, wn.tmpl, as, wn.log, &tmplErr)
	msg := &webhookMessage{
		Version:         "1",
		ExtendedData:    data,
		GroupKey:        groupKey.String(),
		TruncatedAlerts: numTruncated,
		Title:           tmpl(`{{ template "default.title" . }}`),
		Message:         tmpl(`{{ template "default.message" . }}`),
	}

	if types.Alerts(as...).Status() == model.AlertFiring {
		msg.State = string(models.AlertStateAlerting)
	} else {
		msg.State = string(models.AlertStateOK)
	}

	if tmplErr != nil {
		wn.log.Debug("failed to template webhook message", "err", tmplErr.Error())
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return false, err
	}

	cmd := &models.SendWebhookSync{
		Url:        wn.URL,
		User:       wn.User,
		Password:   wn.Password,
		Body:       string(body),
		HttpMethod: wn.HTTPMethod,
	}

	if err := bus.DispatchCtx(ctx, cmd); err != nil {
		return false, err
	}

	return true, nil
}

func truncateAlerts(maxAlerts int, alerts []*types.Alert) ([]*types.Alert, int) {
	if maxAlerts > 0 && len(alerts) > maxAlerts {
		return alerts[:maxAlerts], len(alerts) - maxAlerts
	}

	return alerts, 0
}

func (wn *WebhookNotifier) SendResolved() bool {
	return !wn.GetDisableResolveMessage()
}
