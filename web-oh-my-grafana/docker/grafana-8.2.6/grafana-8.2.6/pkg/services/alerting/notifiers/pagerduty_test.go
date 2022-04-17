package notifiers

import (
	"context"
	"strings"
	"testing"

	"github.com/grafana/grafana/pkg/services/validations"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/grafana/pkg/components/null"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/alerting"
	. "github.com/smartystreets/goconvey/convey"
)

func presenceComparer(a, b string) bool {
	if a == "<<PRESENCE>>" {
		return b != ""
	}
	if b == "<<PRESENCE>>" {
		return a != ""
	}
	return a == b
}

func TestPagerdutyNotifier(t *testing.T) {
	Convey("Pagerduty notifier tests", t, func() {
		Convey("Parsing alert notification from settings", func() {
			Convey("empty settings should return error", func() {
				json := `{ }`

				settingsJSON, err := simplejson.NewJson([]byte(json))
				So(err, ShouldBeNil)

				model := &models.AlertNotification{
					Name:     "pageduty_testing",
					Type:     "pagerduty",
					Settings: settingsJSON,
				}

				_, err = NewPagerdutyNotifier(model)
				So(err, ShouldNotBeNil)
			})

			Convey("severity should override default", func() {
				json := `{ "integrationKey": "abcdefgh0123456789", "severity": "info", "tags": ["foo"]}`

				settingsJSON, err := simplejson.NewJson([]byte(json))
				So(err, ShouldBeNil)

				model := &models.AlertNotification{
					Name:     "pagerduty_testing",
					Type:     "pagerduty",
					Settings: settingsJSON,
				}

				not, err := NewPagerdutyNotifier(model)
				pagerdutyNotifier := not.(*PagerdutyNotifier)

				So(err, ShouldBeNil)
				So(pagerdutyNotifier.Name, ShouldEqual, "pagerduty_testing")
				So(pagerdutyNotifier.Type, ShouldEqual, "pagerduty")
				So(pagerdutyNotifier.Key, ShouldEqual, "abcdefgh0123456789")
				So(pagerdutyNotifier.Severity, ShouldEqual, "info")
				So(pagerdutyNotifier.AutoResolve, ShouldBeFalse)
			})

			Convey("auto resolve and severity should have expected defaults", func() {
				json := `{ "integrationKey": "abcdefgh0123456789" }`

				settingsJSON, err := simplejson.NewJson([]byte(json))
				So(err, ShouldBeNil)

				model := &models.AlertNotification{
					Name:     "pagerduty_testing",
					Type:     "pagerduty",
					Settings: settingsJSON,
				}

				not, err := NewPagerdutyNotifier(model)
				pagerdutyNotifier := not.(*PagerdutyNotifier)

				So(err, ShouldBeNil)
				So(pagerdutyNotifier.Name, ShouldEqual, "pagerduty_testing")
				So(pagerdutyNotifier.Type, ShouldEqual, "pagerduty")
				So(pagerdutyNotifier.Key, ShouldEqual, "abcdefgh0123456789")
				So(pagerdutyNotifier.Severity, ShouldEqual, "critical")
				So(pagerdutyNotifier.AutoResolve, ShouldBeFalse)
			})

			Convey("settings should trigger incident", func() {
				json := `
				{
		  			"integrationKey": "abcdefgh0123456789",
					"autoResolve": false
				}`

				settingsJSON, err := simplejson.NewJson([]byte(json))
				So(err, ShouldBeNil)

				model := &models.AlertNotification{
					Name:     "pagerduty_testing",
					Type:     "pagerduty",
					Settings: settingsJSON,
				}

				not, err := NewPagerdutyNotifier(model)
				pagerdutyNotifier := not.(*PagerdutyNotifier)

				So(err, ShouldBeNil)
				So(pagerdutyNotifier.Name, ShouldEqual, "pagerduty_testing")
				So(pagerdutyNotifier.Type, ShouldEqual, "pagerduty")
				So(pagerdutyNotifier.Key, ShouldEqual, "abcdefgh0123456789")
				So(pagerdutyNotifier.AutoResolve, ShouldBeFalse)
			})

			Convey("should return properly formatted default v2 event payload", func() {
				json := `{
					"integrationKey": "abcdefgh0123456789",
					"autoResolve": false
				}`

				settingsJSON, err := simplejson.NewJson([]byte(json))
				So(err, ShouldBeNil)

				model := &models.AlertNotification{
					Name:     "pagerduty_testing",
					Type:     "pagerduty",
					Settings: settingsJSON,
				}

				not, err := NewPagerdutyNotifier(model)
				So(err, ShouldBeNil)

				pagerdutyNotifier := not.(*PagerdutyNotifier)
				evalContext := alerting.NewEvalContext(context.Background(), &alerting.Rule{
					ID:      0,
					Name:    "someRule",
					Message: "someMessage",
					State:   models.AlertStateAlerting,
				}, &validations.OSSPluginRequestValidator{})
				evalContext.IsTestRun = true

				payloadJSON, err := pagerdutyNotifier.buildEventPayload(evalContext)
				So(err, ShouldBeNil)
				payload, err := simplejson.NewJson(payloadJSON)
				So(err, ShouldBeNil)

				diff := cmp.Diff(map[string]interface{}{
					"client":       "Grafana",
					"client_url":   "",
					"dedup_key":    "alertId-0",
					"event_action": "trigger",
					"links": []interface{}{
						map[string]interface{}{
							"href": "",
						},
					},
					"payload": map[string]interface{}{
						"component": "Grafana",
						"source":    "<<PRESENCE>>",
						"custom_details": map[string]interface{}{
							"state": "alerting",
						},
						"severity":  "critical",
						"summary":   "someRule - someMessage",
						"timestamp": "<<PRESENCE>>",
					},
					"routing_key": "abcdefgh0123456789",
				}, payload.Interface(), cmp.Comparer(presenceComparer))
				So(diff, ShouldBeEmpty)
			})

			Convey("should return properly formatted default v2 event payload with empty message", func() {
				json := `{
					"integrationKey": "abcdefgh0123456789",
					"autoResolve": false
				}`

				settingsJSON, err := simplejson.NewJson([]byte(json))
				So(err, ShouldBeNil)

				model := &models.AlertNotification{
					Name:     "pagerduty_testing",
					Type:     "pagerduty",
					Settings: settingsJSON,
				}

				not, err := NewPagerdutyNotifier(model)
				So(err, ShouldBeNil)

				pagerdutyNotifier := not.(*PagerdutyNotifier)
				evalContext := alerting.NewEvalContext(context.Background(), &alerting.Rule{
					ID:    0,
					Name:  "someRule",
					State: models.AlertStateAlerting,
				}, &validations.OSSPluginRequestValidator{})
				evalContext.IsTestRun = true

				payloadJSON, err := pagerdutyNotifier.buildEventPayload(evalContext)
				So(err, ShouldBeNil)
				payload, err := simplejson.NewJson(payloadJSON)
				So(err, ShouldBeNil)

				diff := cmp.Diff(map[string]interface{}{
					"client":       "Grafana",
					"client_url":   "",
					"dedup_key":    "alertId-0",
					"event_action": "trigger",
					"links": []interface{}{
						map[string]interface{}{
							"href": "",
						},
					},
					"payload": map[string]interface{}{
						"component": "Grafana",
						"source":    "<<PRESENCE>>",
						"custom_details": map[string]interface{}{
							"state": "alerting",
						},
						"severity":  "critical",
						"summary":   "someRule",
						"timestamp": "<<PRESENCE>>",
					},
					"routing_key": "abcdefgh0123456789",
				}, payload.Interface(), cmp.Comparer(presenceComparer))
				So(diff, ShouldBeEmpty)
			})

			Convey("should return properly formatted payload with message moved to details", func() {
				json := `{
					"integrationKey": "abcdefgh0123456789",
					"autoResolve": false,
					"messageInDetails": true
				}`

				settingsJSON, err := simplejson.NewJson([]byte(json))
				So(err, ShouldBeNil)

				model := &models.AlertNotification{
					Name:     "pagerduty_testing",
					Type:     "pagerduty",
					Settings: settingsJSON,
				}

				not, err := NewPagerdutyNotifier(model)
				So(err, ShouldBeNil)

				pagerdutyNotifier := not.(*PagerdutyNotifier)
				evalContext := alerting.NewEvalContext(context.Background(), &alerting.Rule{
					ID:      0,
					Name:    "someRule",
					Message: "someMessage",
					State:   models.AlertStateAlerting,
				}, &validations.OSSPluginRequestValidator{})
				evalContext.IsTestRun = true
				evalContext.EvalMatches = []*alerting.EvalMatch{
					{
						// nil is a terrible value to test with, but the cmp.Diff doesn't
						// like comparing actual floats. So this is roughly the equivalent
						// of <<PRESENCE>>
						Value:  null.FloatFromPtr(nil),
						Metric: "someMetric",
					},
				}

				payloadJSON, err := pagerdutyNotifier.buildEventPayload(evalContext)
				So(err, ShouldBeNil)
				payload, err := simplejson.NewJson(payloadJSON)
				So(err, ShouldBeNil)

				diff := cmp.Diff(map[string]interface{}{
					"client":       "Grafana",
					"client_url":   "",
					"dedup_key":    "alertId-0",
					"event_action": "trigger",
					"links": []interface{}{
						map[string]interface{}{
							"href": "",
						},
					},
					"payload": map[string]interface{}{
						"component": "Grafana",
						"source":    "<<PRESENCE>>",
						"custom_details": map[string]interface{}{
							"message": "someMessage",
							"queries": map[string]interface{}{
								"someMetric": nil,
							},
							"state": "alerting",
						},
						"severity":  "critical",
						"summary":   "someRule",
						"timestamp": "<<PRESENCE>>",
					},
					"routing_key": "abcdefgh0123456789",
				}, payload.Interface(), cmp.Comparer(presenceComparer))
				So(diff, ShouldBeEmpty)
			})

			Convey("should return properly formatted v2 event payload when using override tags", func() {
				json := `{
					"integrationKey": "abcdefgh0123456789",
					"autoResolve": false
				}`

				settingsJSON, err := simplejson.NewJson([]byte(json))
				So(err, ShouldBeNil)

				model := &models.AlertNotification{
					Name:     "pagerduty_testing",
					Type:     "pagerduty",
					Settings: settingsJSON,
				}

				not, err := NewPagerdutyNotifier(model)
				So(err, ShouldBeNil)

				pagerdutyNotifier := not.(*PagerdutyNotifier)

				evalContext := alerting.NewEvalContext(context.Background(), &alerting.Rule{
					ID:      0,
					Name:    "someRule",
					Message: "someMessage",
					State:   models.AlertStateAlerting,
					AlertRuleTags: []*models.Tag{
						{Key: "keyOnly"},
						{Key: "group", Value: "aGroup"},
						{Key: "class", Value: "aClass"},
						{Key: "component", Value: "aComponent"},
						{Key: "severity", Value: "warning"},
						{Key: "dedup_key", Value: "key-" + strings.Repeat("x", 260)},
					},
				}, &validations.OSSPluginRequestValidator{})
				evalContext.ImagePublicURL = "http://somewhere.com/omg_dont_panic.png"
				evalContext.IsTestRun = true

				payloadJSON, err := pagerdutyNotifier.buildEventPayload(evalContext)
				So(err, ShouldBeNil)
				payload, err := simplejson.NewJson(payloadJSON)
				So(err, ShouldBeNil)

				diff := cmp.Diff(map[string]interface{}{
					"client":       "Grafana",
					"client_url":   "",
					"dedup_key":    "key-" + strings.Repeat("x", 250),
					"event_action": "trigger",
					"links": []interface{}{
						map[string]interface{}{
							"href": "",
						},
					},
					"payload": map[string]interface{}{
						"source":    "<<PRESENCE>>",
						"component": "aComponent",
						"custom_details": map[string]interface{}{
							"group":     "aGroup",
							"class":     "aClass",
							"component": "aComponent",
							"severity":  "warning",
							"dedup_key": "key-" + strings.Repeat("x", 250),
							"keyOnly":   "",
							"state":     "alerting",
						},
						"severity":  "warning",
						"summary":   "someRule - someMessage",
						"timestamp": "<<PRESENCE>>",
						"class":     "aClass",
						"group":     "aGroup",
					},
					"images": []interface{}{
						map[string]interface{}{
							"src": "http://somewhere.com/omg_dont_panic.png",
						},
					},
					"routing_key": "abcdefgh0123456789",
				}, payload.Interface(), cmp.Comparer(presenceComparer))
				So(diff, ShouldBeEmpty)
			})

			Convey("should support multiple levels of severity", func() {
				json := `{
					"integrationKey": "abcdefgh0123456789",
					"autoResolve": false
				}`

				settingsJSON, err := simplejson.NewJson([]byte(json))
				So(err, ShouldBeNil)

				model := &models.AlertNotification{
					Name:     "pagerduty_testing",
					Type:     "pagerduty",
					Settings: settingsJSON,
				}

				not, err := NewPagerdutyNotifier(model)
				So(err, ShouldBeNil)

				pagerdutyNotifier := not.(*PagerdutyNotifier)

				evalContext := alerting.NewEvalContext(context.Background(), &alerting.Rule{
					ID:      0,
					Name:    "someRule",
					Message: "someMessage",
					State:   models.AlertStateAlerting,
					AlertRuleTags: []*models.Tag{
						{Key: "keyOnly"},
						{Key: "group", Value: "aGroup"},
						{Key: "class", Value: "aClass"},
						{Key: "component", Value: "aComponent"},
						{Key: "severity", Value: "info"},
					},
				}, &validations.OSSPluginRequestValidator{})
				evalContext.ImagePublicURL = "http://somewhere.com/omg_dont_panic.png"
				evalContext.IsTestRun = true

				payloadJSON, err := pagerdutyNotifier.buildEventPayload(evalContext)
				So(err, ShouldBeNil)
				payload, err := simplejson.NewJson(payloadJSON)
				So(err, ShouldBeNil)

				diff := cmp.Diff(map[string]interface{}{
					"client":       "Grafana",
					"client_url":   "",
					"dedup_key":    "alertId-0",
					"event_action": "trigger",
					"links": []interface{}{
						map[string]interface{}{
							"href": "",
						},
					},
					"payload": map[string]interface{}{
						"source":    "<<PRESENCE>>",
						"component": "aComponent",
						"custom_details": map[string]interface{}{
							"group":     "aGroup",
							"class":     "aClass",
							"component": "aComponent",
							"severity":  "info",
							"keyOnly":   "",
							"state":     "alerting",
						},
						"severity":  "info",
						"summary":   "someRule - someMessage",
						"timestamp": "<<PRESENCE>>",
						"class":     "aClass",
						"group":     "aGroup",
					},
					"images": []interface{}{
						map[string]interface{}{
							"src": "http://somewhere.com/omg_dont_panic.png",
						},
					},
					"routing_key": "abcdefgh0123456789",
				}, payload.Interface(), cmp.Comparer(presenceComparer))
				So(diff, ShouldBeEmpty)
			})

			Convey("should ignore invalid severity for PD but keep the tag", func() {
				json := `{
					"integrationKey": "abcdefgh0123456789",
					"autoResolve": false,
					"severity": "critical"
				}`

				settingsJSON, err := simplejson.NewJson([]byte(json))
				So(err, ShouldBeNil)

				model := &models.AlertNotification{
					Name:     "pagerduty_testing",
					Type:     "pagerduty",
					Settings: settingsJSON,
				}

				not, err := NewPagerdutyNotifier(model)
				So(err, ShouldBeNil)

				pagerdutyNotifier := not.(*PagerdutyNotifier)

				evalContext := alerting.NewEvalContext(context.Background(), &alerting.Rule{
					ID:      0,
					Name:    "someRule",
					Message: "someMessage",
					State:   models.AlertStateAlerting,
					AlertRuleTags: []*models.Tag{
						{Key: "keyOnly"},
						{Key: "group", Value: "aGroup"},
						{Key: "class", Value: "aClass"},
						{Key: "component", Value: "aComponent"},
						{Key: "severity", Value: "llama"},
					},
				}, &validations.OSSPluginRequestValidator{})
				evalContext.ImagePublicURL = "http://somewhere.com/omg_dont_panic.png"
				evalContext.IsTestRun = true

				payloadJSON, err := pagerdutyNotifier.buildEventPayload(evalContext)
				So(err, ShouldBeNil)
				payload, err := simplejson.NewJson(payloadJSON)
				So(err, ShouldBeNil)

				diff := cmp.Diff(map[string]interface{}{
					"client":       "Grafana",
					"client_url":   "",
					"dedup_key":    "alertId-0",
					"event_action": "trigger",
					"links": []interface{}{
						map[string]interface{}{
							"href": "",
						},
					},
					"payload": map[string]interface{}{
						"source":    "<<PRESENCE>>",
						"component": "aComponent",
						"custom_details": map[string]interface{}{
							"group":     "aGroup",
							"class":     "aClass",
							"component": "aComponent",
							"severity":  "llama",
							"keyOnly":   "",
							"state":     "alerting",
						},
						"severity":  "critical",
						"summary":   "someRule - someMessage",
						"timestamp": "<<PRESENCE>>",
						"class":     "aClass",
						"group":     "aGroup",
					},
					"images": []interface{}{
						map[string]interface{}{
							"src": "http://somewhere.com/omg_dont_panic.png",
						},
					},
					"routing_key": "abcdefgh0123456789",
				}, payload.Interface(), cmp.Comparer(presenceComparer))
				So(diff, ShouldBeEmpty)
			})
		})
	})
}
