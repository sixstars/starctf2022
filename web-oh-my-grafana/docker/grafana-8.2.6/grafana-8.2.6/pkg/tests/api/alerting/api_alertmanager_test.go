package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/bus"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	apimodels "github.com/grafana/grafana/pkg/services/ngalert/api/tooling/definitions"
	ngmodels "github.com/grafana/grafana/pkg/services/ngalert/models"
	ngstore "github.com/grafana/grafana/pkg/services/ngalert/store"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/tests/testinfra"
)

func TestAMConfigAccess(t *testing.T) {
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		DisableLegacyAlerting: true,
		EnableUnifiedAlerting: true,
		DisableAnonymous:      true,
	})

	grafanaListedAddr, store := testinfra.StartGrafana(t, dir, path)
	// override bus to get the GetSignedInUserQuery handler
	store.Bus = bus.GetBus()

	// Create a users to make authenticated requests
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_VIEWER),
		Password:       "viewer",
		Login:          "viewer",
	})
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_EDITOR),
		Password:       "editor",
		Login:          "editor",
	})
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_ADMIN),
		Password:       "admin",
		Login:          "admin",
	})

	type testCase struct {
		desc      string
		url       string
		expStatus int
		expBody   string
	}

	t.Run("when creating alertmanager configuration", func(t *testing.T) {
		body := `
		{
			"alertmanager_config": {
				"route": {
					"receiver": "grafana-default-email"
				},
				"receivers": [{
					"name": "grafana-default-email",
					"grafana_managed_receiver_configs": [{
						"uid": "",
						"name": "email receiver",
						"type": "email",
						"isDefault": true,
						"settings": {
							"addresses": "<example@email.com>"
						}
					}]
				}]
			}
		}
		`

		testCases := []testCase{
			{
				desc:      "un-authenticated request should fail",
				url:       "http://%s/api/alertmanager/grafana/config/api/v1/alerts",
				expStatus: http.StatusUnauthorized,
				expBody:   `{"message": "Unauthorized"}`,
			},
			{
				desc:      "viewer request should fail",
				url:       "http://viewer:viewer@%s/api/alertmanager/grafana/config/api/v1/alerts",
				expStatus: http.StatusForbidden,
				expBody:   `{"message": "permission denied"}`,
			},
			{
				desc:      "editor request should succeed",
				url:       "http://editor:editor@%s/api/alertmanager/grafana/config/api/v1/alerts",
				expStatus: http.StatusAccepted,
				expBody:   `{"message":"configuration created"}`,
			},
			{
				desc:      "admin request should succeed",
				url:       "http://admin:admin@%s/api/alertmanager/grafana/config/api/v1/alerts",
				expStatus: http.StatusAccepted,
				expBody:   `{"message":"configuration created"}`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.desc, func(t *testing.T) {
				url := fmt.Sprintf(tc.url, grafanaListedAddr)
				buf := bytes.NewReader([]byte(body))
				// nolint:gosec
				resp, err := http.Post(url, "application/json", buf)
				t.Cleanup(func() {
					require.NoError(t, resp.Body.Close())
				})
				require.NoError(t, err)
				require.Equal(t, tc.expStatus, resp.StatusCode)
				b, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)
				require.JSONEq(t, tc.expBody, string(b))
			})
		}
	})

	t.Run("when retrieve alertmanager configuration", func(t *testing.T) {
		cfgBody := `
		{
			"template_files": null,
			"alertmanager_config": {
				"route": {
					"receiver": "grafana-default-email"
				},
				"templates": null,
				"receivers": [{
					"name": "grafana-default-email",
					"grafana_managed_receiver_configs": [{
						"disableResolveMessage": false,
						"uid": "",
						"name": "email receiver",
						"type": "email",
						"secureFields": {},
						"settings": {
							"addresses": "<example@email.com>"
						}
					}]
				}]
			}
		}
		`
		testCases := []testCase{
			{
				desc:      "un-authenticated request should fail",
				url:       "http://%s/api/alertmanager/grafana/config/api/v1/alerts",
				expStatus: http.StatusUnauthorized,
				expBody:   `{"message": "Unauthorized"}`,
			},
			{
				desc:      "viewer request should fail",
				url:       "http://viewer:viewer@%s/api/alertmanager/grafana/config/api/v1/alerts",
				expStatus: http.StatusForbidden,
				expBody:   `{"message": "permission denied"}`,
			},
			{
				desc:      "editor request should succeed",
				url:       "http://editor:editor@%s/api/alertmanager/grafana/config/api/v1/alerts",
				expStatus: http.StatusOK,
				expBody:   cfgBody,
			},
			{
				desc:      "admin request should succeed",
				url:       "http://admin:admin@%s/api/alertmanager/grafana/config/api/v1/alerts",
				expStatus: http.StatusOK,
				expBody:   cfgBody,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.desc, func(t *testing.T) {
				resp, err := http.Get(fmt.Sprintf(tc.url, grafanaListedAddr))
				t.Cleanup(func() {
					require.NoError(t, resp.Body.Close())
				})
				require.NoError(t, err)
				require.Equal(t, tc.expStatus, resp.StatusCode)
				b, err := ioutil.ReadAll(resp.Body)
				if tc.expStatus == http.StatusOK {
					re := regexp.MustCompile(`"uid":"([\w|-]+)"`)
					b = re.ReplaceAll(b, []byte(`"uid":""`))
				}
				require.NoError(t, err)
				require.JSONEq(t, tc.expBody, string(b))
			})
		}
	})

	t.Run("when creating silence", func(t *testing.T) {
		body := `
		{
			"comment": "string",
			"createdBy": "string",
			"endsAt": "2023-03-31T14:17:04.419Z",
			"matchers": [
			  {
				"isRegex": true,
				"name": "string",
				"value": "string"
			  }
			],
			"startsAt": "2021-03-31T13:17:04.419Z"
		  }
		`

		testCases := []testCase{
			{
				desc:      "un-authenticated request should fail",
				url:       "http://%s/api/alertmanager/grafana/config/api/v2/silences",
				expStatus: http.StatusUnauthorized,
				expBody:   `{"message": "Unauthorized"}`,
			},
			{
				desc:      "viewer request should fail",
				url:       "http://viewer:viewer@%s/api/alertmanager/grafana/api/v2/silences",
				expStatus: http.StatusForbidden,
				expBody:   `{"message": "permission denied"}`,
			},
			{
				desc:      "editor request should succeed",
				url:       "http://editor:editor@%s/api/alertmanager/grafana/api/v2/silences",
				expStatus: http.StatusAccepted,
				expBody:   `{"id": "0", "message":"silence created"}`,
			},
			{
				desc:      "admin request should succeed",
				url:       "http://admin:admin@%s/api/alertmanager/grafana/api/v2/silences",
				expStatus: http.StatusAccepted,
				expBody:   `{"id": "0", "message":"silence created"}`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.desc, func(t *testing.T) {
				url := fmt.Sprintf(tc.url, grafanaListedAddr)
				buf := bytes.NewReader([]byte(body))
				// nolint:gosec
				resp, err := http.Post(url, "application/json", buf)
				t.Cleanup(func() {
					require.NoError(t, resp.Body.Close())
				})
				require.NoError(t, err)
				require.Equal(t, tc.expStatus, resp.StatusCode)
				b, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)
				if tc.expStatus == http.StatusAccepted {
					re := regexp.MustCompile(`"id":"([\w|-]+)"`)
					b = re.ReplaceAll(b, []byte(`"id":"0"`))
				}
				require.JSONEq(t, tc.expBody, string(b))
			})
		}
	})

	var blob []byte
	t.Run("when getting silences", func(t *testing.T) {
		testCases := []testCase{
			{
				desc:      "un-authenticated request should fail",
				url:       "http://%s/api/alertmanager/grafana/api/v2/silences",
				expStatus: http.StatusUnauthorized,
				expBody:   `{"message": "Unauthorized"}`,
			},
			{
				desc:      "viewer request should succeed",
				url:       "http://viewer:viewer@%s/api/alertmanager/grafana/api/v2/silences",
				expStatus: http.StatusOK,
			},
			{
				desc:      "editor request should succeed",
				url:       "http://editor:editor@%s/api/alertmanager/grafana/api/v2/silences",
				expStatus: http.StatusOK,
			},
			{
				desc:      "admin request should succeed",
				url:       "http://admin:admin@%s/api/alertmanager/grafana/api/v2/silences",
				expStatus: http.StatusOK,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.desc, func(t *testing.T) {
				url := fmt.Sprintf(tc.url, grafanaListedAddr)
				// nolint:gosec
				resp, err := http.Get(url)
				t.Cleanup(func() {
					require.NoError(t, resp.Body.Close())
				})
				require.NoError(t, err)
				require.Equal(t, tc.expStatus, resp.StatusCode)
				require.NoError(t, err)
				if tc.expStatus == http.StatusOK {
					b, err := ioutil.ReadAll(resp.Body)
					require.NoError(t, err)
					blob = b
				}
			})
		}
	})

	var silences apimodels.GettableSilences
	err := json.Unmarshal(blob, &silences)
	require.NoError(t, err)
	assert.Len(t, silences, 2)
	silenceIDs := make([]string, 0, len(silences))
	for _, s := range silences {
		silenceIDs = append(silenceIDs, *s.ID)
	}

	unconsumedSilenceIdx := 0
	t.Run("when deleting a silence", func(t *testing.T) {
		testCases := []testCase{
			{
				desc:      "un-authenticated request should fail",
				url:       "http://%s/api/alertmanager/grafana/api/v2/silence/%s",
				expStatus: http.StatusUnauthorized,
				expBody:   `{"message": "Unauthorized"}`,
			},
			{
				desc:      "viewer request should fail",
				url:       "http://viewer:viewer@%s/api/alertmanager/grafana/api/v2/silence/%s",
				expStatus: http.StatusForbidden,
				expBody:   `{"message": "permission denied"}`,
			},
			{
				desc:      "editor request should succeed",
				url:       "http://editor:editor@%s/api/alertmanager/grafana/api/v2/silence/%s",
				expStatus: http.StatusOK,
				expBody:   `{"message": "silence deleted"}`,
			},
			{
				desc:      "admin request should succeed",
				url:       "http://admin:admin@%s/api/alertmanager/grafana/api/v2/silence/%s",
				expStatus: http.StatusOK,
				expBody:   `{"message": "silence deleted"}`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.desc, func(t *testing.T) {
				url := fmt.Sprintf(tc.url, grafanaListedAddr, silenceIDs[unconsumedSilenceIdx])

				// Create client
				client := &http.Client{}

				// Create request
				req, err := http.NewRequest("DELETE", url, nil)
				if err != nil {
					fmt.Println(err)
					return
				}

				// Fetch Request
				resp, err := client.Do(req)
				if err != nil {
					return
				}
				t.Cleanup(func() {
					require.NoError(t, resp.Body.Close())
				})
				require.NoError(t, err)
				require.Equal(t, tc.expStatus, resp.StatusCode)
				b, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)
				if tc.expStatus == http.StatusOK {
					unconsumedSilenceIdx++
				}
				require.JSONEq(t, tc.expBody, string(b))
			})
		}
	})
}

func TestAlertAndGroupsQuery(t *testing.T) {
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		DisableLegacyAlerting: true,
		EnableUnifiedAlerting: true,
		DisableAnonymous:      true,
	})

	grafanaListedAddr, store := testinfra.StartGrafana(t, dir, path)
	// override bus to get the GetSignedInUserQuery handler
	store.Bus = bus.GetBus()

	// unauthenticated request to get the alerts should fail
	{
		alertsURL := fmt.Sprintf("http://%s/api/alertmanager/grafana/api/v2/alerts", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(alertsURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		require.JSONEq(t, `{"message": "Unauthorized"}`, string(b))
	}

	// Create a user to make authenticated requests
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_EDITOR),
		Password:       "password",
		Login:          "grafana",
	})

	// invalid credentials request to get the alerts should fail
	{
		alertsURL := fmt.Sprintf("http://grafana:invalid@%s/api/alertmanager/grafana/api/v2/alerts", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(alertsURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		require.JSONEq(t, `{"error": "invalid username or password","message": "invalid username or password"}`, string(b))
	}

	// When there are no alerts available, it returns an empty list.
	{
		alertsURL := fmt.Sprintf("http://grafana:password@%s/api/alertmanager/grafana/api/v2/alerts", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(alertsURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.JSONEq(t, "[]", string(b))
	}

	// When are there no alerts available, it returns an empty list of groups.
	{
		alertsURL := fmt.Sprintf("http://grafana:password@%s/api/alertmanager/grafana/api/v2/alerts/groups", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(alertsURL)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.JSONEq(t, "[]", string(b))
	}

	// Now, let's test the endpoint with some alerts.
	{
		// Create the namespace we'll save our alerts to.
		_, err := createFolder(t, store, 0, "default")
		require.NoError(t, err)
	}

	// Create an alert that will fire as quickly as possible
	{
		interval, err := model.ParseDuration("10s")
		require.NoError(t, err)
		rules := apimodels.PostableRuleGroupConfig{
			Name:     "arulegroup",
			Interval: interval,
			Rules: []apimodels.PostableExtendedRuleNode{
				{
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
			},
		}
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err = enc.Encode(&rules)
		require.NoError(t, err)

		u := fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Post(u, "application/json", &buf)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		require.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 202)
	}

	// Eventually, we'll get an alert with its state being active.
	{
		alertsURL := fmt.Sprintf("http://grafana:password@%s/api/alertmanager/grafana/api/v2/alerts", grafanaListedAddr)
		// nolint:gosec
		require.Eventually(t, func() bool {
			resp, err := http.Get(alertsURL)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := resp.Body.Close()
				require.NoError(t, err)
			})
			b, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode)

			var alerts apimodels.GettableAlerts
			err = json.Unmarshal(b, &alerts)
			require.NoError(t, err)

			if len(alerts) > 0 {
				status := alerts[0].Status
				return status != nil && status.State != nil && *status.State == "active"
			}

			return false
		}, 18*time.Second, 2*time.Second)
	}
}

func TestRulerAccess(t *testing.T) {
	// Setup Grafana and its Database
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		DisableLegacyAlerting: true,
		EnableUnifiedAlerting: true,
		EnableQuota:           true,
		DisableAnonymous:      true,
		ViewersCanEdit:        true,
	})

	grafanaListedAddr, store := testinfra.StartGrafana(t, dir, path)
	// override bus to get the GetSignedInUserQuery handler
	store.Bus = bus.GetBus()

	// Create the namespace we'll save our alerts to.
	_, err := createFolder(t, store, 0, "default")
	require.NoError(t, err)

	// Create a users to make authenticated requests
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_VIEWER),
		Password:       "viewer",
		Login:          "viewer",
	})
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_EDITOR),
		Password:       "editor",
		Login:          "editor",
	})
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_ADMIN),
		Password:       "admin",
		Login:          "admin",
	})

	// Now, let's test the access policies.
	testCases := []struct {
		desc             string
		url              string
		expStatus        int
		expectedResponse string
	}{
		{
			desc:             "un-authenticated request should fail",
			url:              "http://%s/api/ruler/grafana/api/v1/rules/default",
			expStatus:        http.StatusUnauthorized,
			expectedResponse: `{"message": "Unauthorized"}`,
		},
		{
			desc:             "viewer request should fail",
			url:              "http://viewer:viewer@%s/api/ruler/grafana/api/v1/rules/default",
			expStatus:        http.StatusForbidden,
			expectedResponse: `{"message":"user does not have permissions to edit the namespace: user does not have permissions to edit the namespace"}`,
		},
		{
			desc:             "editor request should succeed",
			url:              "http://editor:editor@%s/api/ruler/grafana/api/v1/rules/default",
			expStatus:        http.StatusAccepted,
			expectedResponse: `{"message":"rule group updated successfully"}`,
		},
		{
			desc:             "admin request should succeed",
			url:              "http://admin:admin@%s/api/ruler/grafana/api/v1/rules/default",
			expStatus:        http.StatusAccepted,
			expectedResponse: `{"message":"rule group updated successfully"}`,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			interval, err := model.ParseDuration("1m")
			require.NoError(t, err)

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
							Title:     fmt.Sprintf("AlwaysFiring %d", i),
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
			err = enc.Encode(&rules)
			require.NoError(t, err)

			u := fmt.Sprintf(tc.url, grafanaListedAddr)
			// nolint:gosec
			resp, err := http.Post(u, "application/json", &buf)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := resp.Body.Close()
				require.NoError(t, err)
			})
			b, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expStatus, resp.StatusCode)
			require.JSONEq(t, tc.expectedResponse, string(b))
		})
	}
}

func TestDeleteFolderWithRules(t *testing.T) {
	// Setup Grafana and its Database
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		DisableLegacyAlerting: true,
		EnableUnifiedAlerting: true,
		EnableQuota:           true,
		DisableAnonymous:      true,
		ViewersCanEdit:        true,
	})

	grafanaListedAddr, store := testinfra.StartGrafana(t, dir, path)
	// override bus to get the GetSignedInUserQuery handler
	store.Bus = bus.GetBus()

	// Create the namespace we'll save our alerts to.
	namespaceUID, err := createFolder(t, store, 0, "default")
	require.NoError(t, err)

	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_VIEWER),
		Password:       "viewer",
		Login:          "viewer",
	})
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_EDITOR),
		Password:       "editor",
		Login:          "editor",
	})

	createRule(t, grafanaListedAddr, "default", "editor", "editor")

	// First, let's have an editor create a rule within the folder/namespace.
	{
		u := fmt.Sprintf("http://editor:editor@%s/api/ruler/grafana/api/v1/rules", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(u)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode)

		re := regexp.MustCompile(`"uid":"([\w|-]+)"`)
		b = re.ReplaceAll(b, []byte(`"uid":""`))
		re = regexp.MustCompile(`"updated":"(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)"`)
		b = re.ReplaceAll(b, []byte(`"updated":"2021-05-19T19:47:55Z"`))

		expectedGetRulesResponseBody := fmt.Sprintf(`{
			"default": [
				{
					"name": "arulegroup",
					"interval": "1m",
					"rules": [
						{
							"expr": "",
							"for": "2m",
							"labels": {
								"label1": "val1"
							},
							"annotations": {
								"annotation1": "val1"
							},
							"grafana_alert": {
								"id": 1,
								"orgId": 1,
								"title": "rule under folder default",
								"condition": "A",
								"data": [
									{
										"refId": "A",
										"queryType": "",
										"relativeTimeRange": {
											"from": 18000,
											"to": 10800
										},
										"datasourceUid": "-100",
										"model": {
											"expression": "2 + 3 > 1",
											"intervalMs": 1000,
											"maxDataPoints": 43200,
											"type": "math"
										}
									}
								],
								"updated": "2021-05-19T19:47:55Z",
								"intervalSeconds": 60,
								"version": 1,
								"uid": "",
								"namespace_uid": %q,
								"namespace_id": 1,
								"rule_group": "arulegroup",
								"no_data_state": "NoData",
								"exec_err_state": "Alerting"
							}
						}
					]
				}
			]
		}`, namespaceUID)
		assert.JSONEq(t, expectedGetRulesResponseBody, string(b))
	}

	// Next, the editor can not delete the folder because it contains Grafana 8 alerts.
	{
		u := fmt.Sprintf("http://editor:editor@%s/api/folders/%s", grafanaListedAddr, namespaceUID)
		req, err := http.NewRequest(http.MethodDelete, u, nil)
		require.NoError(t, err)
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.JSONEq(t, `{"message":"folder cannot be deleted: folder contains alert rules"}`, string(b))
	}

	// Next, the editor can delete the folder if forceDeleteRules is true.
	{
		u := fmt.Sprintf("http://editor:editor@%s/api/folders/%s?forceDeleteRules=true", grafanaListedAddr, namespaceUID)
		req, err := http.NewRequest(http.MethodDelete, u, nil)
		require.NoError(t, err)
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.JSONEq(t, `{"id":1,"message":"Folder default deleted","title":"default"}`, string(b))
	}

	// Finally, we ensure the rules were deleted.
	{
		u := fmt.Sprintf("http://editor:editor@%s/api/ruler/grafana/api/v1/rules", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(u)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode)
		assert.JSONEq(t, "{}", string(b))
	}
}

func TestAlertRuleCRUD(t *testing.T) {
	// Setup Grafana and its Database
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		DisableLegacyAlerting: true,
		EnableUnifiedAlerting: true,
		EnableQuota:           true,
		DisableAnonymous:      true,
	})

	grafanaListedAddr, store := testinfra.StartGrafana(t, dir, path)
	// override bus to get the GetSignedInUserQuery handler
	store.Bus = bus.GetBus()

	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_EDITOR),
		Password:       "password",
		Login:          "grafana",
	})

	// Create the namespace we'll save our alerts to.
	_, err := createFolder(t, store, 0, "default")
	require.NoError(t, err)

	interval, err := model.ParseDuration("1m")
	require.NoError(t, err)

	invalidInterval, err := model.ParseDuration("1s")
	require.NoError(t, err)

	// Now, let's try to create some invalid alert rules.
	{
		testCases := []struct {
			desc             string
			rulegroup        string
			interval         model.Duration
			rule             apimodels.PostableExtendedRuleNode
			expectedResponse string
		}{
			{
				desc:      "alert rule without queries and expressions",
				rulegroup: "arulegroup",
				rule: apimodels.PostableExtendedRuleNode{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For:         interval,
						Labels:      map[string]string{"label1": "val1"},
						Annotations: map[string]string{"annotation1": "val1"},
					},
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title: "AlwaysFiring",
						Data:  []ngmodels.AlertQuery{},
					},
				},
				expectedResponse: `{"message":"failed to update rule group: invalid alert rule: no queries or expressions are found"}`,
			},
			{
				desc:      "alert rule with empty title",
				rulegroup: "arulegroup",
				rule: apimodels.PostableExtendedRuleNode{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For:         interval,
						Labels:      map[string]string{"label1": "val1"},
						Annotations: map[string]string{"annotation1": "val1"},
					},
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title:     "",
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
				expectedResponse: `{"message":"failed to update rule group: invalid alert rule: title is empty"}`,
			},
			{
				desc:      "alert rule with too long name",
				rulegroup: "arulegroup",
				rule: apimodels.PostableExtendedRuleNode{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For:         interval,
						Labels:      map[string]string{"label1": "val1"},
						Annotations: map[string]string{"annotation1": "val1"},
					},
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title:     getLongString(t, ngstore.AlertRuleMaxTitleLength+1),
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
				expectedResponse: `{"message":"failed to update rule group: invalid alert rule: name length should not be greater than 190"}`,
			},
			{
				desc:      "alert rule with too long rulegroup",
				rulegroup: getLongString(t, ngstore.AlertRuleMaxTitleLength+1),
				rule: apimodels.PostableExtendedRuleNode{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For:         interval,
						Labels:      map[string]string{"label1": "val1"},
						Annotations: map[string]string{"annotation1": "val1"},
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
				expectedResponse: `{"message":"failed to update rule group: invalid alert rule: rule group name length should not be greater than 190"}`,
			},
			{
				desc:      "alert rule with invalid interval",
				rulegroup: "arulegroup",
				interval:  invalidInterval,
				rule: apimodels.PostableExtendedRuleNode{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For:         interval,
						Labels:      map[string]string{"label1": "val1"},
						Annotations: map[string]string{"annotation1": "val1"},
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
				expectedResponse: `{"message":"failed to update rule group: invalid alert rule: interval (1s) should be non-zero and divided exactly by scheduler interval: 10s"}`,
			},
			{
				desc:      "alert rule with unknown datasource",
				rulegroup: "arulegroup",
				rule: apimodels.PostableExtendedRuleNode{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For:         interval,
						Labels:      map[string]string{"label1": "val1"},
						Annotations: map[string]string{"annotation1": "val1"},
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
								DatasourceUID: "unknown",
								Model: json.RawMessage(`{
									"type": "math",
									"expression": "2 + 3 > 1"
									}`),
							},
						},
					},
				},
				expectedResponse: `{"message":"failed to validate alert rule \"AlwaysFiring\": invalid query A: data source not found: unknown"}`,
			},
			{
				desc:      "alert rule with invalid condition",
				rulegroup: "arulegroup",
				rule: apimodels.PostableExtendedRuleNode{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For:         interval,
						Labels:      map[string]string{"label1": "val1"},
						Annotations: map[string]string{"annotation1": "val1"},
					},
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title:     "AlwaysFiring",
						Condition: "B",
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
				expectedResponse: `{"message":"failed to validate alert rule \"AlwaysFiring\": condition B not found in any query or expression: it should be one of: [A]"}`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.desc, func(t *testing.T) {
				rules := apimodels.PostableRuleGroupConfig{
					Name:     tc.rulegroup,
					Interval: tc.interval,
					Rules: []apimodels.PostableExtendedRuleNode{
						tc.rule,
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

				assert.Equal(t, resp.StatusCode, http.StatusBadRequest)
				require.JSONEq(t, tc.expectedResponse, string(b))
			})
		}
	}

	var ruleUID string
	var expectedGetNamespaceResponseBody string
	// Now, let's create two alerts.
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

	// With the rules created, let's make sure that rule definition is stored correctly.
	{
		u := fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(u)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, resp.StatusCode, 202)

		body, m := rulesNamespaceWithoutVariableValues(t, b)
		generatedUIDs, ok := m["default,arulegroup"]
		assert.True(t, ok)
		assert.Equal(t, 2, len(generatedUIDs))
		// assert that generated UIDs are unique
		assert.NotEqual(t, generatedUIDs[0], generatedUIDs[1])
		// copy result to a variable with a wider scope
		// to be used by the next test
		ruleUID = generatedUIDs[0]
		expectedGetNamespaceResponseBody = `
		{
		   "default":[
			  {
				 "name":"arulegroup",
				 "interval":"1m",
				 "rules":[
					{
						"annotations": {
							"annotation1": "val1"
					   },
					   "expr":"",
					   "for": "1m",
					   "labels": {
							"label1": "val1"
					   },
					   "grafana_alert":{
						  "id":1,
						  "orgId":1,
						  "title":"AlwaysFiring",
						  "condition":"A",
						  "data":[
							 {
								"refId":"A",
								"queryType":"",
								"relativeTimeRange":{
								   "from":18000,
								   "to":10800
								},
								"datasourceUid":"-100",
								"model":{
								   "expression":"2 + 3 \u003e 1",
								   "intervalMs":1000,
								   "maxDataPoints":43200,
								   "type":"math"
								}
							 }
						  ],
						  "updated":"2021-02-21T01:10:30Z",
						  "intervalSeconds":60,
						  "version":1,
						  "uid":"uid",
						  "namespace_uid":"nsuid",
						  "namespace_id":1,
						  "rule_group":"arulegroup",
						  "no_data_state":"NoData",
						  "exec_err_state":"Alerting"
					   }
					},
					{
					   "expr":"",
					   "grafana_alert":{
						  "id":2,
						  "orgId":1,
						  "title":"AlwaysFiringButSilenced",
						  "condition":"A",
						  "data":[
							 {
								"refId":"A",
								"queryType":"",
								"relativeTimeRange":{
								   "from":18000,
								   "to":10800
								},
								"datasourceUid":"-100",
								"model":{
								   "expression":"2 + 3 \u003e 1",
								   "intervalMs":1000,
								   "maxDataPoints":43200,
								   "type":"math"
								}
							 }
						  ],
						  "updated":"2021-02-21T01:10:30Z",
						  "intervalSeconds":60,
						  "version":1,
						  "uid":"uid",
						  "namespace_uid":"nsuid",
						  "namespace_id":1,
						  "rule_group":"arulegroup",
						  "no_data_state":"Alerting",
						  "exec_err_state":"Alerting"
					   }
					}
				 ]
			  }
		   ]
		}`
		assert.JSONEq(t, expectedGetNamespaceResponseBody, body)
	}

	// try to update by pass an invalid UID
	{
		interval, err := model.ParseDuration("30s")
		require.NoError(t, err)

		rules := apimodels.PostableRuleGroupConfig{
			Name: "arulegroup",
			Rules: []apimodels.PostableExtendedRuleNode{
				{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For: interval,
						Labels: map[string]string{
							"label1": "val42",
							"foo":    "bar",
						},
						Annotations: map[string]string{
							"annotation1": "val42",
							"foo":         "bar",
						},
					},
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						UID:       "unknown",
						Title:     "AlwaysNormal",
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
											"expression": "2 + 3 < 1"
											}`),
							},
						},
						NoDataState:  apimodels.NoDataState(ngmodels.Alerting),
						ExecErrState: apimodels.ExecutionErrorState(ngmodels.AlertingErrState),
					},
				},
			},
			Interval: interval,
		}
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err = enc.Encode(&rules)
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

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		require.JSONEq(t, `{"message":"failed to update rule group: failed to get alert rule unknown: could not find alert rule"}`, string(b))

		// let's make sure that rule definitions are not affected by the failed POST request.
		u = fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err = http.Get(u)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err = ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, resp.StatusCode, 202)

		body, m := rulesNamespaceWithoutVariableValues(t, b)
		returnedUIDs, ok := m["default,arulegroup"]
		assert.True(t, ok)
		assert.Equal(t, 2, len(returnedUIDs))
		assert.JSONEq(t, expectedGetNamespaceResponseBody, body)
	}

	// try to update by pass two rules with conflicting UIDs
	{
		interval, err := model.ParseDuration("30s")
		require.NoError(t, err)

		rules := apimodels.PostableRuleGroupConfig{
			Name: "arulegroup",
			Rules: []apimodels.PostableExtendedRuleNode{
				{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For: interval,
						Labels: map[string]string{
							"label1": "val42",
							"foo":    "bar",
						},
						Annotations: map[string]string{
							"annotation1": "val42",
							"foo":         "bar",
						},
					},
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						UID:       ruleUID,
						Title:     "AlwaysNormal",
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
												"expression": "2 + 3 < 1"
												}`),
							},
						},
						NoDataState:  apimodels.NoDataState(ngmodels.Alerting),
						ExecErrState: apimodels.ExecutionErrorState(ngmodels.AlertingErrState),
					},
				},
				{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For: interval,
						Labels: map[string]string{
							"label1": "val42",
							"foo":    "bar",
						},
						Annotations: map[string]string{
							"annotation1": "val42",
							"foo":         "bar",
						},
					},
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						UID:       ruleUID,
						Title:     "AlwaysAlerting",
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
			Interval: interval,
		}
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err = enc.Encode(&rules)
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

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.JSONEq(t, fmt.Sprintf(`{"message":"failed to validate alert rule \"AlwaysAlerting\": conflicting UID \"%s\" found"}`, ruleUID), string(b))

		// let's make sure that rule definitions are not affected by the failed POST request.
		u = fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err = http.Get(u)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err = ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, resp.StatusCode, 202)

		body, m := rulesNamespaceWithoutVariableValues(t, b)
		returnedUIDs, ok := m["default,arulegroup"]
		assert.True(t, ok)
		assert.Equal(t, 2, len(returnedUIDs))
		assert.JSONEq(t, expectedGetNamespaceResponseBody, body)
	}

	// update the first rule and completely remove the other
	{
		forValue, err := model.ParseDuration("30s")
		require.NoError(t, err)

		rules := apimodels.PostableRuleGroupConfig{
			Name: "arulegroup",
			Rules: []apimodels.PostableExtendedRuleNode{
				{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For: forValue,
						Labels: map[string]string{
							// delete foo label
							"label1": "val1", // update label value
							"label2": "val2", // new label
						},
						Annotations: map[string]string{
							// delete foo annotation
							"annotation1": "val1", // update annotation value
							"annotation2": "val2", // new annotation
						},
					},
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						UID:       ruleUID, // Including the UID in the payload makes the endpoint update the existing rule.
						Title:     "AlwaysNormal",
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
											"expression": "2 + 3 < 1"
											}`),
							},
						},
						NoDataState:  apimodels.NoDataState(ngmodels.Alerting),
						ExecErrState: apimodels.ExecutionErrorState(ngmodels.AlertingErrState),
					},
				},
			},
			Interval: interval,
		}
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err = enc.Encode(&rules)
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

		// let's make sure that rule definitions are updated correctly.
		u = fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err = http.Get(u)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err = ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, resp.StatusCode, 202)

		body, m := rulesNamespaceWithoutVariableValues(t, b)
		returnedUIDs, ok := m["default,arulegroup"]
		assert.True(t, ok)
		assert.Equal(t, 1, len(returnedUIDs))
		assert.Equal(t, ruleUID, returnedUIDs[0])
		assert.JSONEq(t, `
		{
		   "default":[
		      {
		         "name":"arulegroup",
		         "interval":"1m",
		         "rules":[
		            {
						"annotations": {
							"annotation1": "val1",
							"annotation2": "val2"
					   },
		               "expr":"",
					   "for": "30s",
					   "labels": {
							"label1": "val1",
							"label2": "val2"
					   },
		               "grafana_alert":{
		                  "id":1,
		                  "orgId":1,
		                  "title":"AlwaysNormal",
		                  "condition":"A",
		                  "data":[
		                     {
		                        "refId":"A",
		                        "queryType":"",
		                        "relativeTimeRange":{
		                           "from":18000,
		                           "to":10800
		                        },
		                        "datasourceUid":"-100",
								"model":{
		                           "expression":"2 + 3 \u003C 1",
		                           "intervalMs":1000,
		                           "maxDataPoints":43200,
		                           "type":"math"
		                        }
		                     }
		                  ],
		                  "updated":"2021-02-21T01:10:30Z",
		                  "intervalSeconds":60,
		                  "version":2,
		                  "uid":"uid",
		                  "namespace_uid":"nsuid",
		                  "namespace_id":1,
		                  "rule_group":"arulegroup",
		                  "no_data_state":"Alerting",
		                  "exec_err_state":"Alerting"
		               }
		            }
		         ]
		      }
		   ]
		}`, body)
	}

	// update the rule; delete labels and annotations
	{
		forValue, err := model.ParseDuration("30s")
		require.NoError(t, err)

		rules := apimodels.PostableRuleGroupConfig{
			Name: "arulegroup",
			Rules: []apimodels.PostableExtendedRuleNode{
				{
					ApiRuleNode: &apimodels.ApiRuleNode{
						For: forValue,
					},
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						UID:       ruleUID, // Including the UID in the payload makes the endpoint update the existing rule.
						Title:     "AlwaysNormal",
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
												"expression": "2 + 3 < 1"
												}`),
							},
						},
						NoDataState:  apimodels.NoDataState(ngmodels.Alerting),
						ExecErrState: apimodels.ExecutionErrorState(ngmodels.AlertingErrState),
					},
				},
			},
			Interval: interval,
		}
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err = enc.Encode(&rules)
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

		// let's make sure that rule definitions are updated correctly.
		u = fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err = http.Get(u)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err = ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, resp.StatusCode, 202)

		body, m := rulesNamespaceWithoutVariableValues(t, b)
		returnedUIDs, ok := m["default,arulegroup"]
		assert.True(t, ok)
		assert.Equal(t, 1, len(returnedUIDs))
		assert.Equal(t, ruleUID, returnedUIDs[0])
		assert.JSONEq(t, `
			{
			   "default":[
			      {
				 "name":"arulegroup",
				 "interval":"1m",
				 "rules":[
				    {
				       "expr":"",
				       "for": "30s",
				       "grafana_alert":{
					  "id":1,
					  "orgId":1,
					  "title":"AlwaysNormal",
					  "condition":"A",
					  "data":[
					     {
						"refId":"A",
						"queryType":"",
						"relativeTimeRange":{
						   "from":18000,
						   "to":10800
						},
						"datasourceUid":"-100",
									"model":{
						   "expression":"2 + 3 \u003C 1",
						   "intervalMs":1000,
						   "maxDataPoints":43200,
						   "type":"math"
						}
					     }
					  ],
					  "updated":"2021-02-21T01:10:30Z",
					  "intervalSeconds":60,
					  "version":3,
					  "uid":"uid",
					  "namespace_uid":"nsuid",
					  "namespace_id":1,
					  "rule_group":"arulegroup",
					  "no_data_state":"Alerting",
					  "exec_err_state":"Alerting"
				       }
				    }
				 ]
			      }
			   ]
			}`, body)
	}

	// update the rule; keep title, condition, no data state, error state, queries and expressions if not provided
	{
		rules := apimodels.PostableRuleGroupConfig{
			Name: "arulegroup",
			Rules: []apimodels.PostableExtendedRuleNode{
				{
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						UID: ruleUID, // Including the UID in the payload makes the endpoint update the existing rule.
					},
				},
			},
			Interval: interval,
		}
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err = enc.Encode(&rules)
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

		// let's make sure that rule definitions are updated correctly.
		u = fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err = http.Get(u)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err = ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, resp.StatusCode, 202)

		body, m := rulesNamespaceWithoutVariableValues(t, b)
		returnedUIDs, ok := m["default,arulegroup"]
		assert.True(t, ok)
		assert.Equal(t, 1, len(returnedUIDs))
		assert.Equal(t, ruleUID, returnedUIDs[0])
		assert.JSONEq(t, `
			{
			   "default":[
			      {
				 "name":"arulegroup",
				 "interval":"1m",
				 "rules":[
				    {
				       "expr":"",
				       "grafana_alert":{
					  "id":1,
					  "orgId":1,
					  "title":"AlwaysNormal",
					  "condition":"A",
					  "data":[
					     {
						"refId":"A",
						"queryType":"",
						"relativeTimeRange":{
						   "from":18000,
						   "to":10800
						},
						"datasourceUid":"-100",
									"model":{
						   "expression":"2 + 3 \u003C 1",
						   "intervalMs":1000,
						   "maxDataPoints":43200,
						   "type":"math"
						}
					     }
					  ],
					  "updated":"2021-02-21T01:10:30Z",
					  "intervalSeconds":60,
					  "version":4,
					  "uid":"uid",
					  "namespace_uid":"nsuid",
					  "namespace_id":1,
					  "rule_group":"arulegroup",
					  "no_data_state":"Alerting",
					  "exec_err_state":"Alerting"
				       }
				    }
				 ]
			      }
			   ]
			}`, body)
	}

	client := &http.Client{}
	// Finally, make sure we can delete it.
	{
		t.Run("fail if he rule group name does not exists", func(t *testing.T) {
			u := fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default/groupnotexist", grafanaListedAddr)
			req, err := http.NewRequest(http.MethodDelete, u, nil)
			require.NoError(t, err)
			resp, err := client.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := resp.Body.Close()
				require.NoError(t, err)
			})
			b, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, http.StatusNotFound, resp.StatusCode)
			require.JSONEq(t, `{"message":"failed to delete rule group: rule group not found under this namespace"}`, string(b))
		})

		t.Run("succeed if the rule group name does exist", func(t *testing.T) {
			u := fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default/arulegroup", grafanaListedAddr)
			req, err := http.NewRequest(http.MethodDelete, u, nil)
			require.NoError(t, err)
			resp, err := client.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := resp.Body.Close()
				require.NoError(t, err)
			})
			b, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, http.StatusAccepted, resp.StatusCode)
			require.JSONEq(t, `{"message":"rule group deleted"}`, string(b))
		})
	}
}

func TestAlertmanagerStatus(t *testing.T) {
	// Setup Grafana and its Database
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		DisableLegacyAlerting: true,
		EnableUnifiedAlerting: true,
	})

	grafanaListedAddr, _ := testinfra.StartGrafana(t, dir, path)

	// Get the Alertmanager current status.
	{
		alertsURL := fmt.Sprintf("http://%s/api/alertmanager/grafana/api/v2/status", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(alertsURL)
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
	"cluster": {
		"peers": [],
		"status": "disabled"
	},
	"config": {
		"route": {
			"receiver": "grafana-default-email"
		},
		"templates": null,
		"receivers": [{
			"name": "grafana-default-email",
			"grafana_managed_receiver_configs": [{
				"uid": "",
				"name": "email receiver",
				"type": "email",
				"disableResolveMessage": false,
				"settings": {
					"addresses": "\u003cexample@email.com\u003e"
				},
				"secureSettings": null
			}]
		}]
	},
	"uptime": null,
	"versionInfo": {
		"branch": "N/A",
		"buildDate": "N/A",
		"buildUser": "N/A",
		"goVersion": "N/A",
		"revision": "N/A",
		"version": "N/A"
	}
}
`, string(b))
	}
}

func TestQuota(t *testing.T) {
	// Setup Grafana and its Database
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		DisableLegacyAlerting: true,
		EnableUnifiedAlerting: true,
		EnableQuota:           true,
		DisableAnonymous:      true,
	})

	grafanaListedAddr, store := testinfra.StartGrafana(t, dir, path)
	// override bus to get the GetSignedInUserQuery handler
	store.Bus = bus.GetBus()

	// Create the namespace we'll save our alerts to.
	_, err := createFolder(t, store, 0, "default")
	require.NoError(t, err)

	// Create a user to make authenticated requests
	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_EDITOR),
		Password:       "password",
		Login:          "grafana",
	})

	interval, err := model.ParseDuration("1m")
	require.NoError(t, err)

	// Create rule under folder1
	createRule(t, grafanaListedAddr, "default", "grafana", "password")

	// get the generated rule UID
	var ruleUID string
	{
		u := fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err := http.Get(u)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, resp.StatusCode, 202)

		_, m := rulesNamespaceWithoutVariableValues(t, b)
		generatedUIDs, ok := m["default,arulegroup"]
		assert.True(t, ok)
		assert.Equal(t, 1, len(generatedUIDs))
		ruleUID = generatedUIDs[0]
	}

	// check quota limits
	t.Run("when quota limit exceed creating new rule should fail", func(t *testing.T) {
		// get existing org quota
		query := models.GetOrgQuotaByTargetQuery{OrgId: 1, Target: "alert_rule"}
		err = sqlstore.GetOrgQuotaByTarget(&query)
		require.NoError(t, err)
		used := query.Result.Used
		limit := query.Result.Limit

		// set org quota limit to equal used
		orgCmd := models.UpdateOrgQuotaCmd{
			OrgId:  1,
			Target: "alert_rule",
			Limit:  used,
		}
		err := sqlstore.UpdateOrgQuota(&orgCmd)
		require.NoError(t, err)

		t.Cleanup(func() {
			// reset org quota to original value
			orgCmd := models.UpdateOrgQuotaCmd{
				OrgId:  1,
				Target: "alert_rule",
				Limit:  limit,
			}
			err := sqlstore.UpdateOrgQuota(&orgCmd)
			require.NoError(t, err)
		})

		// try to create an alert rule
		rules := apimodels.PostableRuleGroupConfig{
			Name:     "arulegroup",
			Interval: interval,
			Rules: []apimodels.PostableExtendedRuleNode{
				{
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title:     "One more alert rule",
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
		err = enc.Encode(&rules)
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
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		require.JSONEq(t, `{"message":"quota reached"}`, string(b))
	})

	t.Run("when quota limit exceed updating existing rule should succeed", func(t *testing.T) {
		// try to create an alert rule
		rules := apimodels.PostableRuleGroupConfig{
			Name:     "arulegroup",
			Interval: interval,
			Rules: []apimodels.PostableExtendedRuleNode{
				{
					GrafanaManagedAlert: &apimodels.PostableGrafanaRule{
						Title:     "Updated alert rule",
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
									"expression": "2 + 4 > 1"
									}`),
							},
						},
						UID: ruleUID,
					},
				},
			},
		}
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err = enc.Encode(&rules)
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
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		require.JSONEq(t, `{"message":"rule group updated successfully"}`, string(b))

		// let's make sure that rule definitions are updated correctly.
		u = fmt.Sprintf("http://grafana:password@%s/api/ruler/grafana/api/v1/rules/default", grafanaListedAddr)
		// nolint:gosec
		resp, err = http.Get(u)
		require.NoError(t, err)
		t.Cleanup(func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		})
		b, err = ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, resp.StatusCode, 202)

		body, m := rulesNamespaceWithoutVariableValues(t, b)
		returnedUIDs, ok := m["default,arulegroup"]
		assert.True(t, ok)
		assert.Equal(t, 1, len(returnedUIDs))
		assert.Equal(t, ruleUID, returnedUIDs[0])
		assert.JSONEq(t, `
				{
				   "default":[
				      {
					 "name":"arulegroup",
					 "interval":"1m",
					 "rules":[
					    {
					       "expr":"",
					       "grafana_alert":{
						  "id":1,
						  "orgId":1,
						  "title":"Updated alert rule",
						  "condition":"A",
						  "data":[
						     {
							"refId":"A",
							"queryType":"",
							"relativeTimeRange":{
							   "from":18000,
							   "to":10800
							},
							"datasourceUid":"-100",
										"model":{
							   "expression":"2 + 4 \u003E 1",
							   "intervalMs":1000,
							   "maxDataPoints":43200,
							   "type":"math"
							}
						     }
						  ],
						  "updated":"2021-02-21T01:10:30Z",
						  "intervalSeconds":60,
						  "version":2,
						  "uid":"uid",
						  "namespace_uid":"nsuid",
						  "namespace_id":1,
						  "rule_group":"arulegroup",
						  "no_data_state":"NoData",
						  "exec_err_state":"Alerting"
					       }
					    }
					 ]
				      }
				   ]
				}`, body)
	})
}

func TestEval(t *testing.T) {
	// Setup Grafana and its Database
	dir, path := testinfra.CreateGrafDir(t, testinfra.GrafanaOpts{
		DisableLegacyAlerting: true,
		EnableUnifiedAlerting: true,
		EnableQuota:           true,
		DisableAnonymous:      true,
	})

	grafanaListedAddr, store := testinfra.StartGrafana(t, dir, path)
	// override bus to get the GetSignedInUserQuery handler
	store.Bus = bus.GetBus()

	createUser(t, store, models.CreateUserCommand{
		DefaultOrgRole: string(models.ROLE_EDITOR),
		Password:       "password",
		Login:          "grafana",
	})

	// Create the namespace we'll save our alerts to.
	_, err := createFolder(t, store, 0, "default")
	require.NoError(t, err)

	// test eval conditions
	testCases := []struct {
		desc               string
		payload            string
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			desc: "alerting condition",
			payload: `
			{
				"grafana_condition": {
				"condition": "A",
				"data": [
					{
						"refId": "A",
						"relativeTimeRange": {
							"from": 18000,
							"to": 10800
						},
						"datasourceUid":"-100",
						"model": {
							"type":"math",
							"expression":"1 < 2"
						}
					}
				],
				"now": "2021-04-11T14:38:14Z"
				}
			}
			`,
			expectedStatusCode: http.StatusOK,
			expectedResponse: `{
			"instances": [
			  {
				"schema": {
				  "name": "evaluation results",
				  "fields": [
					{
					  "name": "State",
					  "type": "string",
					  "typeInfo": {
						"frame": "string"
					  }
					},
					{
					  "name": "Info",
					  "type": "string",
					  "typeInfo": {
					    "frame": "string"
					  }
					}
				  ]
				},
				"data": {
				  "values": [
					[
					  "Alerting"
					],
					[
					  "[ var='A' labels={} value=1 ]"
					]
				  ]
				}
			  }
			]
		  }`,
		},
		{
			desc: "normal condition",
			payload: `
			{
				"grafana_condition": {
				"condition": "A",
				"data": [
					{
						"refId": "A",
						"relativeTimeRange": {
							"from": 18000,
							"to": 10800
						},
						"datasourceUid": "-100",
						"model": {
							"type":"math",
							"expression":"1 > 2"
						}
					}
				],
				"now": "2021-04-11T14:38:14Z"
				}
			}
			`,
			expectedStatusCode: http.StatusOK,
			expectedResponse: `{
			"instances": [
			  {
				"schema": {
				  "name": "evaluation results",
				  "fields": [
					{
					  "name": "State",
					  "type": "string",
					  "typeInfo": {
						"frame": "string"
					  }
					},
					{
					  "name": "Info",
					  "type": "string",
					  "typeInfo": {
					    "frame": "string"
					  }
					}
				  ]
				},
				"data": {
				  "values": [
					[
					  "Normal"
					],
					[
					  "[ var='A' labels={} value=0 ]"
					]
				  ]
				}
			  }
			]
		  }`,
		},
		{
			desc: "condition not found in any query or expression",
			payload: `
			{
				"grafana_condition": {
				"condition": "B",
				"data": [
					{
						"refId": "A",
						"relativeTimeRange": {
							"from": 18000,
							"to": 10800
						},
						"datasourceUid": "-100",
						"model": {
							"type":"math",
							"expression":"1 > 2"
						}
					}
				],
				"now": "2021-04-11T14:38:14Z"
				}
			}
			`,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"message":"invalid condition: condition B not found in any query or expression: it should be one of: [A]"}`,
		},
		{
			desc: "unknown query datasource",
			payload: `
			{
				"grafana_condition": {
				"condition": "A",
				"data": [
					{
						"refId": "A",
						"relativeTimeRange": {
							"from": 18000,
							"to": 10800
						},
						"datasourceUid": "unknown",
						"model": {
						}
					}
				],
				"now": "2021-04-11T14:38:14Z"
				}
			}
			`,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"message":"invalid condition: invalid query A: data source not found: unknown"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			u := fmt.Sprintf("http://grafana:password@%s/api/v1/rule/test/grafana", grafanaListedAddr)
			r := strings.NewReader(tc.payload)
			// nolint:gosec
			resp, err := http.Post(u, "application/json", r)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := resp.Body.Close()
				require.NoError(t, err)
			})
			b, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)
			require.JSONEq(t, tc.expectedResponse, string(b))
		})
	}

	// test eval queries and expressions
	testCases = []struct {
		desc               string
		payload            string
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			desc: "alerting condition",
			payload: `
			{
				"data": [
						{
							"refId": "A",
							"relativeTimeRange": {
								"from": 18000,
								"to": 10800
							},
							"datasourceUid": "-100",
							"model": {
								"type":"math",
								"expression":"1 < 2"
							}
						}
					],
				"now": "2021-04-11T14:38:14Z"
			}
			`,
			expectedStatusCode: http.StatusOK,
			expectedResponse: `{
				"results": {
				  "A": {
					"frames": [
					  {
						"schema": {
						  "refId": "A",
						  "fields": [
							{
							  "name": "A",
							  "type": "number",
							  "typeInfo": {
								"frame": "float64",
								"nullable": true
							  }
							}
						  ]
						},
						"data": {
						  "values": [
							[
							  1
							]
						  ]
						}
					  }
					]
				  }
				}
			}`,
		},
		{
			desc: "normal condition",
			payload: `
			{
				"data": [
						{
							"refId": "A",
							"relativeTimeRange": {
								"from": 18000,
								"to": 10800
							},
							"datasourceUid": "-100",
							"model": {
								"type":"math",
								"expression":"1 > 2"
							}
						}
					],
				"now": "2021-04-11T14:38:14Z"
			}
			`,
			expectedStatusCode: http.StatusOK,
			expectedResponse: `{
				"results": {
				  "A": {
					"frames": [
					  {
						"schema": {
						  "refId": "A",
						  "fields": [
							{
							  "name": "A",
							  "type": "number",
							  "typeInfo": {
								"frame": "float64",
								"nullable": true
							  }
							}
						  ]
						},
						"data": {
						  "values": [
							[
							  0
							]
						  ]
						}
					  }
					]
				  }
				}
			}`,
		},
		{
			desc: "unknown query datasource",
			payload: `
			{
				"data": [
						{
							"refId": "A",
							"relativeTimeRange": {
								"from": 18000,
								"to": 10800
							},
							"datasourceUid": "unknown",
							"model": {
							}
						}
					],
				"now": "2021-04-11T14:38:14Z"
			}
			`,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"message":"invalid queries or expressions: invalid query A: data source not found: unknown"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			u := fmt.Sprintf("http://grafana:password@%s/api/v1/eval", grafanaListedAddr)
			r := strings.NewReader(tc.payload)
			// nolint:gosec
			resp, err := http.Post(u, "application/json", r)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := resp.Body.Close()
				require.NoError(t, err)
			})
			b, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)
			require.JSONEq(t, tc.expectedResponse, string(b))
		})
	}
}

// createFolder creates a folder for storing our alerts under. Grafana uses folders as a replacement for alert namespaces to match its permission model.
// We use the dashboard command using IsFolder = true to tell it's a folder, it takes the dashboard as the name of the folder.
func createFolder(t *testing.T, store *sqlstore.SQLStore, folderID int64, folderName string) (string, error) {
	t.Helper()

	cmd := models.SaveDashboardCommand{
		OrgId:    1, // default organisation
		FolderId: folderID,
		IsFolder: true,
		Dashboard: simplejson.NewFromAny(map[string]interface{}{
			"title": folderName,
		}),
	}
	f, err := store.SaveDashboard(cmd)

	if err != nil {
		return "", err
	}

	return f.Uid, nil
}

// rulesNamespaceWithoutVariableValues takes a apimodels.NamespaceConfigResponse JSON-based input and makes the dynamic fields static e.g. uid, dates, etc.
// it returns a map of the modified rule UIDs with the namespace,rule_group as a key
func rulesNamespaceWithoutVariableValues(t *testing.T, b []byte) (string, map[string][]string) {
	t.Helper()

	var r apimodels.NamespaceConfigResponse
	require.NoError(t, json.Unmarshal(b, &r))
	// create a map holding the created rule UIDs per namespace/group
	m := make(map[string][]string)
	for namespace, nodes := range r {
		for _, node := range nodes {
			compositeKey := strings.Join([]string{namespace, node.Name}, ",")
			_, ok := m[compositeKey]
			if !ok {
				m[compositeKey] = make([]string, 0, len(node.Rules))
			}
			for _, rule := range node.Rules {
				m[compositeKey] = append(m[compositeKey], rule.GrafanaManagedAlert.UID)
				rule.GrafanaManagedAlert.UID = "uid"
				rule.GrafanaManagedAlert.NamespaceUID = "nsuid"
				rule.GrafanaManagedAlert.Updated = time.Date(2021, time.Month(2), 21, 1, 10, 30, 0, time.UTC)
			}
		}
	}

	json, err := json.Marshal(&r)
	require.NoError(t, err)
	return string(json), m
}

func createUser(t *testing.T, store *sqlstore.SQLStore, cmd models.CreateUserCommand) int64 {
	t.Helper()

	u, err := store.CreateUser(context.Background(), cmd)
	require.NoError(t, err)
	return u.Id
}

func createOrg(t *testing.T, store *sqlstore.SQLStore, name string, userID int64) int64 {
	org, err := store.CreateOrgWithMember(name, userID)
	require.NoError(t, err)
	return org.Id
}

func getLongString(t *testing.T, n int) string {
	t.Helper()

	b := make([]rune, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
