package dtos

import (
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/components/null"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
)

func formatShort(interval time.Duration) string {
	var result string

	hours := interval / time.Hour
	if hours > 0 {
		result += fmt.Sprintf("%dh", hours)
	}

	remaining := interval - (hours * time.Hour)
	mins := remaining / time.Minute
	if mins > 0 {
		result += fmt.Sprintf("%dm", mins)
	}

	remaining -= (mins * time.Minute)
	seconds := remaining / time.Second
	if seconds > 0 {
		result += fmt.Sprintf("%ds", seconds)
	}

	return result
}

func NewAlertNotification(notification *models.AlertNotification) *AlertNotification {
	dto := &AlertNotification{
		Id:                    notification.Id,
		Uid:                   notification.Uid,
		Name:                  notification.Name,
		Type:                  notification.Type,
		IsDefault:             notification.IsDefault,
		Created:               notification.Created,
		Updated:               notification.Updated,
		Frequency:             formatShort(notification.Frequency),
		SendReminder:          notification.SendReminder,
		DisableResolveMessage: notification.DisableResolveMessage,
		Settings:              notification.Settings,
		SecureFields:          map[string]bool{},
	}

	if notification.SecureSettings != nil {
		for k := range notification.SecureSettings {
			dto.SecureFields[k] = true
		}
	}

	return dto
}

type AlertNotification struct {
	Id                    int64            `json:"id"`
	Uid                   string           `json:"uid"`
	Name                  string           `json:"name"`
	Type                  string           `json:"type"`
	IsDefault             bool             `json:"isDefault"`
	SendReminder          bool             `json:"sendReminder"`
	DisableResolveMessage bool             `json:"disableResolveMessage"`
	Frequency             string           `json:"frequency"`
	Created               time.Time        `json:"created"`
	Updated               time.Time        `json:"updated"`
	Settings              *simplejson.Json `json:"settings"`
	SecureFields          map[string]bool  `json:"secureFields"`
}

func NewAlertNotificationLookup(notification *models.AlertNotification) *AlertNotificationLookup {
	return &AlertNotificationLookup{
		Id:        notification.Id,
		Uid:       notification.Uid,
		Name:      notification.Name,
		Type:      notification.Type,
		IsDefault: notification.IsDefault,
	}
}

type AlertNotificationLookup struct {
	Id        int64  `json:"id"`
	Uid       string `json:"uid"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsDefault bool   `json:"isDefault"`
}

type AlertTestCommand struct {
	Dashboard *simplejson.Json `json:"dashboard" binding:"Required"`
	PanelId   int64            `json:"panelId" binding:"Required"`
}

type AlertTestResult struct {
	Firing         bool                  `json:"firing"`
	State          models.AlertStateType `json:"state"`
	ConditionEvals string                `json:"conditionEvals"`
	TimeMs         string                `json:"timeMs"`
	Error          string                `json:"error,omitempty"`
	EvalMatches    []*EvalMatch          `json:"matches,omitempty"`
	Logs           []*AlertTestResultLog `json:"logs,omitempty"`
}

type AlertTestResultLog struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type EvalMatch struct {
	Tags   map[string]string `json:"tags,omitempty"`
	Metric string            `json:"metric"`
	Value  null.Float        `json:"value"`
}

type NotificationTestCommand struct {
	ID                    int64             `json:"id,omitempty"`
	Name                  string            `json:"name"`
	Type                  string            `json:"type"`
	SendReminder          bool              `json:"sendReminder"`
	DisableResolveMessage bool              `json:"disableResolveMessage"`
	Frequency             string            `json:"frequency"`
	Settings              *simplejson.Json  `json:"settings"`
	SecureSettings        map[string]string `json:"secureSettings"`
}

type PauseAlertCommand struct {
	AlertId int64 `json:"alertId"`
	Paused  bool  `json:"paused"`
}

type PauseAllAlertsCommand struct {
	Paused bool `json:"paused"`
}
