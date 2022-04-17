package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
	apimodels "github.com/grafana/grafana/pkg/services/ngalert/api/tooling/definitions"
	ngmodels "github.com/grafana/grafana/pkg/services/ngalert/models"
	"github.com/grafana/grafana/pkg/tests/testinfra"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusRules(t *testing.T) {
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		DisableLegacyAlerting: true,
		EnableUnifiedAlerting: true,
		DisableAnonymous:      true,
	})

	grafanaListedAddr, store := testinfra.StartGrafana(t, dir, path)
	// override bus to get the GetSignedInUserQuery handler
	store.Bus = bus.GetBus()

	// Create the namespace under default organisation (orgID = 1) where we'll save our alerts to.
	_, err := createFolder(t, store, 0, "default")
	require.NoError(t, err)

	// Create a user to make authenticated requests
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_EDITOR),
		Password:       "password",
		Login:          "grafana",
	})

	interval, err := model.ParseDuration("10s")
	require.NoError(t, err)

	// an unauthenticated request to get rules should fail
	{
		promRulesURL := fmt.Sprintf("http://%s/api/prometheus/grafana/api/v1/rules", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		require.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	}

	// When we have no alerting rules, it returns an empty list.
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		require.JSONEq(t, `{"status": "success", "data": {"groups": []}}`, string(b))
	}

	// Now, let's create some rules
	{
		rules := apimodels.PostableRuleGroupConfig{
			Name: "arulegroup",
			Rules: []apimodels.PostableExtendedRuleNode{
				{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For:         interval,
						Labels:      map[string]string{"label1": "val1"},
						Annotations: map[string]string{"annotation1": "val1"},
					},
					// this rule does not explicitly set no data and error states
					// therefore it should get the default values
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title:     "AlwaysFiring",
						Condition: "A",
						Data: []ngmodels.AlertQuery{
							{
								RefID: "A",
								RelativeTimeRange: ngmodels.RelativeTimeRange{
									From: ngmodels.Duration(time.Duration(5) * time.Hour),
									To:   ngmodels.Duration(time.Duration(3) * time.Hour),
								},
								DatasourceUID: "-100",
								Model: json.RawMessage(`{
									"type": "math",
									"expression": "2 + 3 > 1"
									}`),
							},
						},
					},
				},
				{
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title:     "AlwaysFiringButSilenced",
						Condition: "A",
						Data: []ngmodels.AlertQuery{
							{
								RefID: "A",
								RelativeTimeRange: ngmodels.RelativeTimeRange{
									From: ngmodels.Duration(time.Duration(5) * time.Hour),
									To:   ngmodels.Duration(time.Duration(3) * time.Hour),
								},
								DatasourceUID: "-100",
								Model: json.RawMessage(`{
									"type": "math",
									"expression": "2 + 3 > 1"
									}`),
							},
						},
						NoDataState:  apimodels.NoDataState(ngmodels.Alerting),
						ExecErrState: apimodels.ExecutionErrorState(ngmodels.AlertingErrState),
					},
				},
			},
		}
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err := enc.Encode(&rules)
		require.NoError(t, err)

		u := fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Post(u, "application/json", &buf)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, resp.StatusCode, 202)
		require.JSONEq(t, `{"message":"rule group updated successfully"}`, string(b))
	}

	// Check that we cannot create a rule that has a panel_id and no dashboard_uid
	{
		rules := apimodels.PostableRuleGroupConfig{
			Name: "anotherrulegroup",
			Rules: []apimodels.PostableExtendedRuleNode{
				{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For:         interval,
						Labels:      map[string]string{},
						Annotations: map[string]string{"__panelId__": "1"},
					},
					// this rule does not explicitly set no data and error states
					// therefore it should get the default values
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title:     "NeverCreated",
						Condition: "A",
						Data: []ngmodels.AlertQuery{
							{
								RefID: "A",
								RelativeTimeRange: ngmodels.RelativeTimeRange{
									From: ngmodels.Duration(time.Duration(5) * time.Hour),
									To:   ngmodels.Duration(time.Duration(3) * time.Hour),
								},
								DatasourceUID: "-100",
								Model: json.RawMessage(`{
									"type": "math",
									"expression": "2 + 3 > 1"
									}`),
							},
						},
					},
				},
			},
		}
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err := enc.Encode(&rules)
		require.NoError(t, err)

		u := fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Post(u, "application/json", &buf)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, 400, resp.StatusCode)
		require.JSONEq(t, `{"message":"failed to update rule group: invalid alert rule: cannot have Panel ID without a Dashboard UID"}`, string(b))
	}

	// Now, let's see how this looks like.
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.JSONEq(t, `
{
	"status": "success",
	"data": {
		"groups": [{
			"name": "arulegroup",
			"file": "default",
			"rules": [{
				"state": "inactive",
				"name": "AlwaysFiring",
				"query": "[{\"refId\":\"A\",\"queryType\":\"\",\"relativeTimeRange\":{\"from\":18000,\"to\":10800},\"datasourceUid\":\"-100\",\"model\":{\"expression\":\"2 + 3 \\u003e 1\",\"intervalMs\":1000,\"maxDataPoints\":43200,\"type\":\"math\"}}]",
				"duration": 10,
				"annotations": {
					"annotation1": "val1"
				},
				"labels": {
					"label1": "val1"
				},
				"health": "ok",
				"lastError": "",
				"type": "alerting",
				"lastEvaluation": "0001-01-01T00:00:00Z",
				"evaluationTime": 0
			}, {
				"state": "inactive",
				"name": "AlwaysFiringButSilenced",
				"query": "[{\"refId\":\"A\",\"queryType\":\"\",\"relativeTimeRange\":{\"from\":18000,\"to\":10800},\"datasourceUid\":\"-100\",\"model\":{\"expression\":\"2 + 3 \\u003e 1\",\"intervalMs\":1000,\"maxDataPoints\":43200,\"type\":\"math\"}}]",
				"labels": null,
				"health": "ok",
				"lastError": "",
				"type": "alerting",
				"lastEvaluation": "0001-01-01T00:00:00Z",
				"evaluationTime": 0
			}],
			"interval": 60,
			"lastEvaluation": "0001-01-01T00:00:00Z",
			"evaluationTime": 0
		}]
	}
}`, string(b))
	}

	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules", grafanaListedAddr)
		// nolint:gosec
		require.Eventually(t, func() bool {
			resp, err := http.Get(promRulesURL)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := resp.Body.Close()
				require.NoError(t, err)
			})
			b, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode)
			require.JSONEq(t, `
{
	"status": "success",
	"data": {
		"groups": [{
			"name": "arulegroup",
			"file": "default",
			"rules": [{
				"state": "inactive",
				"name": "AlwaysFiring",
				"query": "[{\"refId\":\"A\",\"queryType\":\"\",\"relativeTimeRange\":{\"from\":18000,\"to\":10800},\"datasourceUid\":\"-100\",\"model\":{\"expression\":\"2 + 3 \\u003e 1\",\"intervalMs\":1000,\"maxDataPoints\":43200,\"type\":\"math\"}}]",
				"duration": 10,
				"annotations": {
					"annotation1": "val1"
				},
				"labels": {
					"label1": "val1"
				},
				"health": "ok",
				"lastError": "",
				"type": "alerting",
				"lastEvaluation": "0001-01-01T00:00:00Z",
				"evaluationTime": 0
			}, {
				"state": "inactive",
				"name": "AlwaysFiringButSilenced",
				"query": "[{\"refId\":\"A\",\"queryType\":\"\",\"relativeTimeRange\":{\"from\":18000,\"to\":10800},\"datasourceUid\":\"-100\",\"model\":{\"expression\":\"2 + 3 \\u003e 1\",\"intervalMs\":1000,\"maxDataPoints\":43200,\"type\":\"math\"}}]",
				"labels": null,
				"health": "ok",
				"lastError": "",
				"type": "alerting",
				"lastEvaluation": "0001-01-01T00:00:00Z",
				"evaluationTime": 0
			}],
			"interval": 60,
			"lastEvaluation": "0001-01-01T00:00:00Z",
			"evaluationTime": 0
		}]
	}
}`, string(b))
			return true
		}, 18*time.Second, 2*time.Second)
	}
}

func TestPrometheusRulesFilterByDashboard(t *testing.T) {
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		EnableFeatureToggles: []string{"ngalert"},
		DisableAnonymous:     true,
	})

	grafanaListedAddr, store := testinfra.StartGrafana(t, dir, path)
	// override bus to get the GetSignedInUserQuery handler
	store.Bus = bus.GetBus()

	// Create the namespace under default organisation (orgID = 1) where we'll save our alerts to.
	dashboardUID, err := createFolder(t, store, 0, "default")
	require.NoError(t, err)

	// Create a user to make authenticated requests
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_EDITOR),
		Password:       "password",
		Login:          "grafana",
	})

	interval, err := model.ParseDuration("10s")
	require.NoError(t, err)

	// Now, let's create some rules
	{
		rules := apimodels.PostableRuleGroupConfig{
			Name: "anotherrulegroup",
			Rules: []apimodels.PostableExtendedRuleNode{
				{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For:    interval,
						Labels: map[string]string{},
						Annotations: map[string]string{
							"__dashboardUid__": dashboardUID,
							"__panelId__":      "1",
						},
					},
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title:     "AlwaysFiring",
						Condition: "A",
						Data: []ngmodels.AlertQuery{
							{
								RefID: "A",
								RelativeTimeRange: ngmodels.RelativeTimeRange{
									From: ngmodels.Duration(time.Duration(5) * time.Hour),
									To:   ngmodels.Duration(time.Duration(3) * time.Hour),
								},
								DatasourceUID: "-100",
								Model: json.RawMessage(`{
									"type": "math",
									"expression": "2 + 3 > 1"
									}`),
							},
						},
					},
				},
				{
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title:     "AlwaysFiringButSilenced",
						Condition: "A",
						Data: []ngmodels.AlertQuery{
							{
								RefID: "A",
								RelativeTimeRange: ngmodels.RelativeTimeRange{
									From: ngmodels.Duration(time.Duration(5) * time.Hour),
									To:   ngmodels.Duration(time.Duration(3) * time.Hour),
								},
								DatasourceUID: "-100",
								Model: json.RawMessage(`{
									"type": "math",
									"expression": "2 + 3 > 1"
									}`),
							},
						},
						NoDataState:  apimodels.NoDataState(ngmodels.Alerting),
						ExecErrState: apimodels.ExecutionErrorState(ngmodels.AlertingErrState),
					},
				},
			},
		}
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err := enc.Encode(&rules)
		require.NoError(t, err)

		u := fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Post(u, "application/json", &buf)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, resp.StatusCode, 202)
		require.JSONEq(t, `{"message":"rule group updated successfully"}`, string(b))
	}

	expectedAllJSON := fmt.Sprintf(`
{
	"status": "success",
	"data": {
		"groups": [{
			"name": "anotherrulegroup",
			"file": "default",
			"rules": [{
				"state": "inactive",
				"name": "AlwaysFiring",
				"query": "[{\"refId\":\"A\",\"queryType\":\"\",\"relativeTimeRange\":{\"from\":18000,\"to\":10800},\"datasourceUid\":\"-100\",\"model\":{\"expression\":\"2 + 3 \\u003e 1\",\"intervalMs\":1000,\"maxDataPoints\":43200,\"type\":\"math\"}}]",
				"duration": 10,
				"annotations": {
					"__dashboardUid__": "%s",
					"__panelId__": "1"
				},
				"labels": null,
				"health": "ok",
				"lastError": "",
				"type": "alerting",
				"lastEvaluation": "0001-01-01T00:00:00Z",
				"evaluationTime": 0
			}, {
				"state": "inactive",
				"name": "AlwaysFiringButSilenced",
				"query": "[{\"refId\":\"A\",\"queryType\":\"\",\"relativeTimeRange\":{\"from\":18000,\"to\":10800},\"datasourceUid\":\"-100\",\"model\":{\"expression\":\"2 + 3 \\u003e 1\",\"intervalMs\":1000,\"maxDataPoints\":43200,\"type\":\"math\"}}]",
				"labels": null,
				"health": "ok",
				"lastError": "",
				"type": "alerting",
				"lastEvaluation": "0001-01-01T00:00:00Z",
				"evaluationTime": 0
			}],
			"interval": 60,
			"lastEvaluation": "0001-01-01T00:00:00Z",
			"evaluationTime": 0
		}]
	}
}`, dashboardUID)
	expectedFilteredByJSON := fmt.Sprintf(`
{
	"status": "success",
	"data": {
		"groups": [{
			"name": "anotherrulegroup",
			"file": "default",
			"rules": [{
				"state": "inactive",
				"name": "AlwaysFiring",
				"query": "[{\"refId\":\"A\",\"queryType\":\"\",\"relativeTimeRange\":{\"from\":18000,\"to\":10800},\"datasourceUid\":\"-100\",\"model\":{\"expression\":\"2 + 3 \\u003e 1\",\"intervalMs\":1000,\"maxDataPoints\":43200,\"type\":\"math\"}}]",
				"duration": 10,
				"annotations": {
					"__dashboardUid__": "%s",
					"__panelId__": "1"
				},
				"labels": null,
				"health": "ok",
				"lastError": "",
				"type": "alerting",
				"lastEvaluation": "0001-01-01T00:00:00Z",
				"evaluationTime": 0
			}],
			"interval": 60,
			"lastEvaluation": "0001-01-01T00:00:00Z",
			"evaluationTime": 0
		}]
	}
}`, dashboardUID)
	expectedNoneJSON := `
{
	"status": "success",
	"data": {
		"groups": []
	}
}`

	// Now, let's see how this looks like.
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.JSONEq(t, expectedAllJSON, string(b))
	}

	// Now, let's check we get the same rule when filtering by dashboard_uid
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules?dashboard_uid=%s", grafanaListedAddr, dashboardUID)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.JSONEq(t, expectedFilteredByJSON, string(b))
	}

	// Now, let's check we get no rules when filtering by an unknown dashboard_uid
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules?dashboard_uid=%s", grafanaListedAddr, "abc")
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.JSONEq(t, expectedNoneJSON, string(b))
	}

	// Now, let's check we get the same rule when filtering by dashboard_uid and panel_id
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules?dashboard_uid=%s&panel_id=1", grafanaListedAddr, dashboardUID)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.JSONEq(t, expectedFilteredByJSON, string(b))
	}

	// Now, let's check we get no rules when filtering by dashboard_uid and unknown panel_id
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules?dashboard_uid=%s&panel_id=2", grafanaListedAddr, dashboardUID)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.JSONEq(t, expectedNoneJSON, string(b))
	}

	// Now, let's check an invalid panel_id returns a 400 Bad Request response
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules?dashboard_uid=%s&panel_id=invalid", grafanaListedAddr, dashboardUID)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"message":"invalid panel_id: strconv.ParseInt: parsing \"invalid\": invalid syntax"}`, string(b))
	}

	// Now, let's check a panel_id without dashboard_uid returns a 400 Bad Request response
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules?panel_id=1", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"message":"panel_id must be set with dashboard_uid"}`, string(b))
	}
}

func TestPrometheusRulesPermissions(t *testing.T) {
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		DisableLegacyAlerting: true,
		EnableUnifiedAlerting: true,
		DisableAnonymous:      true,
	})

	grafanaListedAddr, store := testinfra.StartGrafana(t, dir, path)
	// override bus to get the GetSignedInUserQuery handler
	store.Bus = bus.GetBus()

	// Create a user to make authenticated requests
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_EDITOR),
		Password:       "password",
		Login:          "grafana",
	})

	// Create a namespace under default organisation (orgID = 1) where we'll save some alerts.
	_, err := createFolder(t, store, 0, "folder1")
	require.NoError(t, err)

	// Create another namespace under default organisation (orgID = 1) where we'll save some alerts.
	_, err = createFolder(t, store, 0, "folder2")
	require.NoError(t, err)

	// Create rule under folder1
	createRule(t, grafanaListedAddr, "folder1", "grafana", "password")

	// Create rule under folder2
	createRule(t, grafanaListedAddr, "folder2", "grafana", "password")

	// Now, let's see how this looks like.
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.JSONEq(t, `
{
	"status": "success",
	"data": {
		"groups": [{
			"name": "arulegroup",
			"file": "folder1",
			"rules": [{
				"state": "inactive",
				"name": "rule under folder folder1",
				"query": "[{\"refId\":\"A\",\"queryType\":\"\",\"relativeTimeRange\":{\"from\":18000,\"to\":10800},\"datasourceUid\":\"-100\",\"model\":{\"expression\":\"2 + 3 \\u003e 1\",\"intervalMs\":1000,\"maxDataPoints\":43200,\"type\":\"math\"}}]",
				"duration": 120,
				"annotations": {
					"annotation1": "val1"
				},
				"labels": {
					"label1": "val1"
				},
				"health": "ok",
				"lastError": "",
				"type": "alerting",
				"lastEvaluation": "0001-01-01T00:00:00Z",
				"evaluationTime": 0
			}],
			"interval": 60,
			"lastEvaluation": "0001-01-01T00:00:00Z",
			"evaluationTime": 0
		},
		{
			"name": "arulegroup",
			"file": "folder2",
			"rules": [{
				"state": "inactive",
				"name": "rule under folder folder2",
				"query": "[{\"refId\":\"A\",\"queryType\":\"\",\"relativeTimeRange\":{\"from\":18000,\"to\":10800},\"datasourceUid\":\"-100\",\"model\":{\"expression\":\"2 + 3 \\u003e 1\",\"intervalMs\":1000,\"maxDataPoints\":43200,\"type\":\"math\"}}]",
				"duration": 120,
				"annotations": {
					"annotation1": "val1"
				},
				"labels": {
					"label1": "val1"
				},
				"health": "ok",
				"lastError": "",
				"type": "alerting",
				"lastEvaluation": "0001-01-01T00:00:00Z",
				"evaluationTime": 0
			}],
			"interval": 60,
			"lastEvaluation": "0001-01-01T00:00:00Z",
			"evaluationTime": 0
		}]
	}
}`, string(b))
	}

	// remove permissions from folder2
	require.NoError(t, store.UpdateDashboardACL(2, nil))

	// make sure that folder2 is not included in the response
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.JSONEq(t, `
{
	"status": "success",
	"data": {
		"groups": [{
			"name": "arulegroup",
			"file": "folder1",
			"rules": [{
				"state": "inactive",
				"name": "rule under folder folder1",
				"query": "[{\"refId\":\"A\",\"queryType\":\"\",\"relativeTimeRange\":{\"from\":18000,\"to\":10800},\"datasourceUid\":\"-100\",\"model\":{\"expression\":\"2 + 3 \\u003e 1\",\"intervalMs\":1000,\"maxDataPoints\":43200,\"type\":\"math\"}}]",
				"duration": 120,
				"annotations": {
					"annotation1": "val1"
				},
				"labels": {
					"label1": "val1"
				},
				"health": "ok",
				"lastError": "",
				"type": "alerting",
				"lastEvaluation": "0001-01-01T00:00:00Z",
				"evaluationTime": 0
			}],
			"interval": 60,
			"lastEvaluation": "0001-01-01T00:00:00Z",
			"evaluationTime": 0
		}]
	}
}`, string(b))
	}

	// remove permissions from _ALL_ folders
	require.NoError(t, store.UpdateDashboardACL(1, nil))

	// make sure that no folders are included in the response
	{
		promRulesURL := fmt.Sprintf("http://grafana:password@%s/api/prometheus/grafana/api/v1/rules", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(promRulesURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		require.JSONEq(t, `
{
	"status": "success",
	"data": {
		"groups": []
	}
}`, string(b))
	}
}
