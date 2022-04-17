package channels

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
)

func TestOpsgenieNotifier(t *testing.T) {
	tmpl := templateForTests(t)

	externalURL, err := url.Parse("http://localhost")
	require.NoError(t, err)
	tmpl.ExternalURL = externalURL

	cases := []struct {
		name         string
		settings     string
		alerts       []*types.Alert
		expMsg       string
		expInitError string
		expMsgError  error
	}{
		{
			name:     "Default config with one alert",
			settings: `{"apiKey": "abcdefgh0123456789"}`,
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1", "__dashboardUid__": "abcd", "__panelId__": "efgh"},
					},
				},
			},
			expMsg: `{
				"alias": "6e3538104c14b583da237e9693b76debbc17f0f8058ef20492e5853096cf8733",
				"description": "[FIRING:1]  (val1)\nhttp://localhost/alerting/list\n\n**Firing**\n\nLabels:\n - alertname = alert1\n - lbl1 = val1\nAnnotations:\n - ann1 = annv1\nSilence: http://localhost/alerting/silence/new?alertmanager=grafana&matchers=alertname%3Dalert1%2Clbl1%3Dval1\nDashboard: http://localhost/d/abcd\nPanel: http://localhost/d/abcd?viewPanel=efgh\n",
				"details": {
					"url": "http://localhost/alerting/list"
				},
				"message": "[FIRING:1]  (val1)",
				"source": "Grafana",
				"tags": ["alertname:alert1", "lbl1:val1"]
			}`,
		},
		{
			name: "Default config with one alert and send tags as tags",
			settings: `{
				"apiKey": "abcdefgh0123456789",
				"sendTagsAs": "tags"
			}`,
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1"},
					},
				},
			},
			expMsg: `{
				"alias": "6e3538104c14b583da237e9693b76debbc17f0f8058ef20492e5853096cf8733",
				"description": "[FIRING:1]  (val1)\nhttp://localhost/alerting/list\n\n**Firing**\n\nLabels:\n - alertname = alert1\n - lbl1 = val1\nAnnotations:\n - ann1 = annv1\nSilence: http://localhost/alerting/silence/new?alertmanager=grafana&matchers=alertname%3Dalert1%2Clbl1%3Dval1\n",
				"details": {
					"url": "http://localhost/alerting/list"
				},
				"message": "[FIRING:1]  (val1)",
				"source": "Grafana",
				"tags": ["alertname:alert1", "lbl1:val1"]
			}`,
		},
		{
			name: "Default config with one alert and send tags as details",
			settings: `{
				"apiKey": "abcdefgh0123456789",
				"sendTagsAs": "details"
			}`,
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1"},
					},
				},
			},
			expMsg: `{
				"alias": "6e3538104c14b583da237e9693b76debbc17f0f8058ef20492e5853096cf8733",
				"description": "[FIRING:1]  (val1)\nhttp://localhost/alerting/list\n\n**Firing**\n\nLabels:\n - alertname = alert1\n - lbl1 = val1\nAnnotations:\n - ann1 = annv1\nSilence: http://localhost/alerting/silence/new?alertmanager=grafana&matchers=alertname%3Dalert1%2Clbl1%3Dval1\n",
				"details": {
					"alertname": "alert1",
					"lbl1": "val1",
					"url": "http://localhost/alerting/list"
				},
				"message": "[FIRING:1]  (val1)",
				"source": "Grafana",
				"tags": []
			}`,
		},
		{
			name: "Custom config with multiple alerts and send tags as both details and tag",
			settings: `{
				"apiKey": "abcdefgh0123456789",
				"sendTagsAs": "both"
			}`,
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1"},
					},
				}, {
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val2"},
						Annotations: model.LabelSet{"ann1": "annv1"},
					},
				},
			},
			expMsg: `{
				"alias": "6e3538104c14b583da237e9693b76debbc17f0f8058ef20492e5853096cf8733",
				"description": "[FIRING:2]  \nhttp://localhost/alerting/list\n\n**Firing**\n\nLabels:\n - alertname = alert1\n - lbl1 = val1\nAnnotations:\n - ann1 = annv1\nSilence: http://localhost/alerting/silence/new?alertmanager=grafana&matchers=alertname%3Dalert1%2Clbl1%3Dval1\n\nLabels:\n - alertname = alert1\n - lbl1 = val2\nAnnotations:\n - ann1 = annv1\nSilence: http://localhost/alerting/silence/new?alertmanager=grafana&matchers=alertname%3Dalert1%2Clbl1%3Dval2\n",
				"details": {
					"alertname": "alert1",
					"url": "http://localhost/alerting/list"
				},
				"message": "[FIRING:2]  ",
				"source": "Grafana",
				"tags": ["alertname:alert1"]
			}`,
			expMsgError: nil,
		},
		{
			name:     "Resolved is not sent when auto close is false",
			settings: `{"apiKey": "abcdefgh0123456789", "autoClose": false}`,
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1"},
						EndsAt:      time.Now().Add(-1 * time.Minute),
					},
				},
			},
		},
		{
			name:         "Error when incorrect settings",
			settings:     `{}`,
			expInitError: `failed to validate receiver "opsgenie_testing" of type "opsgenie": could not find api key property in settings`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			settingsJSON, err := simplejson.NewJson([]byte(c.settings))
			require.NoError(t, err)

			m := &NotificationChannelConfig{
				Name:     "opsgenie_testing",
				Type:     "opsgenie",
				Settings: settingsJSON,
			}

			pn, err := NewOpsgenieNotifier(m, tmpl)
			if c.expInitError != "" {
				require.Error(t, err)
				require.Equal(t, c.expInitError, err.Error())
				return
			}
			require.NoError(t, err)

			body := "<not-sent>"
			bus.AddHandlerCtx("test", func(ctx context.Context, webhook *models.SendWebhookSync) error {
				body = webhook.Body
				return nil
			})

			ctx := notify.WithGroupKey(context.Background(), "alertname")
			ctx = notify.WithGroupLabels(ctx, model.LabelSet{"alertname": ""})
			ok, err := pn.Notify(ctx, c.alerts...)
			if c.expMsgError != nil {
				require.False(t, ok)
				require.Error(t, err)
				require.Equal(t, c.expMsgError.Error(), err.Error())
				return
			}
			require.True(t, ok)
			require.NoError(t, err)

			if c.expMsg == "" {
				// No notification was expected.
				require.Equal(t, "<not-sent>", body)
			} else {
				require.JSONEq(t, c.expMsg, body)
			}
		})
	}
}
