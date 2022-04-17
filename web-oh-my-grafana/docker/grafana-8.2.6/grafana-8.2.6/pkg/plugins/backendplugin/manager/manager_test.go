package manager

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/grafana/grafana-aws-sdk/pkg/awsds"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/plugins/backendplugin"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/require"
)

const testPluginID = "test-plugin"

func TestManager(t *testing.T) {
	newManagerScenario(t, false, func(t *testing.T, ctx *managerScenarioCtx) {
		t.Run("Unregistered plugin scenario", func(t *testing.T) {
			err := ctx.manager.StartPlugin(context.Background(), testPluginID)
			require.Equal(t, backendplugin.ErrPluginNotRegistered, err)

			_, err = ctx.manager.CollectMetrics(context.Background(), testPluginID)
			require.Equal(t, backendplugin.ErrPluginNotRegistered, err)

			_, err = ctx.manager.CheckHealth(context.Background(), backend.PluginContext{PluginID: testPluginID})
			require.Equal(t, backendplugin.ErrPluginNotRegistered, err)

			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			err = ctx.manager.callResourceInternal(w, req, backend.PluginContext{PluginID: testPluginID})
			require.Equal(t, backendplugin.ErrPluginNotRegistered, err)
		})
	})

	newManagerScenario(t, true, func(t *testing.T, ctx *managerScenarioCtx) {
		t.Run("Managed plugin scenario", func(t *testing.T) {
			ctx.license.edition = "Open Source"
			ctx.license.hasLicense = false
			ctx.cfg.BuildVersion = "7.0.0"

			t.Run("Should be able to register plugin", func(t *testing.T) {
				err := ctx.manager.RegisterAndStart(context.Background(), testPluginID, ctx.factory)
				require.NoError(t, err)
				require.NotNil(t, ctx.plugin)
				require.Equal(t, testPluginID, ctx.plugin.pluginID)
				require.NotNil(t, ctx.plugin.logger)
				require.Equal(t, 1, ctx.plugin.startCount)
				require.True(t, ctx.manager.IsRegistered(testPluginID))

				t.Run("Should not be able to register an already registered plugin", func(t *testing.T) {
					err := ctx.manager.RegisterAndStart(context.Background(), testPluginID, ctx.factory)
					require.Equal(t, 1, ctx.plugin.startCount)
					require.Error(t, err)
				})

				t.Run("Should provide expected host environment variables", func(t *testing.T) {
					require.Len(t, ctx.env, 7)
					require.EqualValues(t, []string{
						"GF_VERSION=7.0.0",
						"GF_EDITION=Open Source",
						fmt.Sprintf("%s=true", awsds.AssumeRoleEnabledEnvVarKeyName),
						fmt.Sprintf("%s=keys,credentials", awsds.AllowedAuthProvidersEnvVarKeyName),
						"AZURE_CLOUD=AzureCloud",
						"AZURE_MANAGED_IDENTITY_CLIENT_ID=client-id",
						"AZURE_MANAGED_IDENTITY_ENABLED=true"},
						ctx.env)
				})

				t.Run("When manager runs should start and stop plugin", func(t *testing.T) {
					pCtx := context.Background()
					cCtx, cancel := context.WithCancel(pCtx)
					var wg sync.WaitGroup
					wg.Add(1)
					var runErr error
					go func() {
						runErr = ctx.manager.Run(cCtx)
						wg.Done()
					}()
					time.Sleep(time.Millisecond)
					cancel()
					wg.Wait()
					require.Equal(t, context.Canceled, runErr)
					require.Equal(t, 1, ctx.plugin.startCount)
					require.Equal(t, 1, ctx.plugin.stopCount)
				})

				t.Run("When manager runs should restart plugin process when killed", func(t *testing.T) {
					ctx.plugin.stopCount = 0
					ctx.plugin.startCount = 0
					pCtx := context.Background()
					cCtx, cancel := context.WithCancel(pCtx)
					var wgRun sync.WaitGroup
					wgRun.Add(1)
					var runErr error
					go func() {
						runErr = ctx.manager.Run(cCtx)
						wgRun.Done()
					}()

					time.Sleep(time.Millisecond)

					var wgKill sync.WaitGroup
					wgKill.Add(1)
					go func() {
						ctx.plugin.kill()
						for {
							if !ctx.plugin.Exited() {
								break
							}
						}
						cancel()
						wgKill.Done()
					}()
					wgKill.Wait()
					wgRun.Wait()
					require.Equal(t, context.Canceled, runErr)
					require.Equal(t, 1, ctx.plugin.stopCount)
					require.Equal(t, 1, ctx.plugin.startCount)
				})

				t.Run("Shouldn't be able to start managed plugin", func(t *testing.T) {
					err := ctx.manager.StartPlugin(context.Background(), testPluginID)
					require.NotNil(t, err)
				})

				t.Run("Unimplemented handlers", func(t *testing.T) {
					t.Run("Collect metrics should return method not implemented error", func(t *testing.T) {
						_, err = ctx.manager.CollectMetrics(context.Background(), testPluginID)
						require.Equal(t, backendplugin.ErrMethodNotImplemented, err)
					})

					t.Run("Check health should return method not implemented error", func(t *testing.T) {
						_, err = ctx.manager.CheckHealth(context.Background(), backend.PluginContext{PluginID: testPluginID})
						require.Equal(t, backendplugin.ErrMethodNotImplemented, err)
					})

					t.Run("Call resource should return method not implemented error", func(t *testing.T) {
						req, err := http.NewRequest(http.MethodGet, "/test", bytes.NewReader([]byte{}))
						require.NoError(t, err)
						w := httptest.NewRecorder()
						err = ctx.manager.callResourceInternal(w, req, backend.PluginContext{PluginID: testPluginID})
						require.Equal(t, backendplugin.ErrMethodNotImplemented, err)
					})
				})

				t.Run("Implemented handlers", func(t *testing.T) {
					t.Run("Collect metrics should return expected result", func(t *testing.T) {
						ctx.plugin.CollectMetricsHandlerFunc = func(ctx context.Context) (*backend.CollectMetricsResult, error) {
							return &backend.CollectMetricsResult{
								PrometheusMetrics: []byte("hello"),
							}, nil
						}

						res, err := ctx.manager.CollectMetrics(context.Background(), testPluginID)
						require.NoError(t, err)
						require.NotNil(t, res)
						require.Equal(t, "hello", string(res.PrometheusMetrics))
					})

					t.Run("Check health should return expected result", func(t *testing.T) {
						json := []byte(`{
							"key": "value"
						}`)
						ctx.plugin.CheckHealthHandlerFunc = func(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
							return &backend.CheckHealthResult{
								Status:      backend.HealthStatusOk,
								Message:     "All good",
								JSONDetails: json,
							}, nil
						}

						res, err := ctx.manager.CheckHealth(context.Background(), backend.PluginContext{PluginID: testPluginID})
						require.NoError(t, err)
						require.NotNil(t, res)
						require.Equal(t, backend.HealthStatusOk, res.Status)
						require.Equal(t, "All good", res.Message)
						require.Equal(t, json, res.JSONDetails)
					})

					t.Run("Call resource should return expected response", func(t *testing.T) {
						ctx.plugin.CallResourceHandlerFunc = func(ctx context.Context,
							req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
							return sender.Send(&backend.CallResourceResponse{
								Status: http.StatusOK,
							})
						}

						req, err := http.NewRequest(http.MethodGet, "/test", bytes.NewReader([]byte{}))
						require.NoError(t, err)
						w := httptest.NewRecorder()
						err = ctx.manager.callResourceInternal(w, req, backend.PluginContext{PluginID: testPluginID})
						require.NoError(t, err)
						require.Equal(t, http.StatusOK, w.Code)
					})
				})

				t.Run("Should be able to decommission a running plugin", func(t *testing.T) {
					require.True(t, ctx.manager.IsRegistered(testPluginID))

					err := ctx.manager.UnregisterAndStop(context.Background(), testPluginID)
					require.NoError(t, err)

					require.Equal(t, 2, ctx.plugin.stopCount)
					require.False(t, ctx.manager.IsRegistered(testPluginID))
					p := ctx.manager.plugins[testPluginID]
					require.Nil(t, p)

					err = ctx.manager.StartPlugin(context.Background(), testPluginID)
					require.Equal(t, backendplugin.ErrPluginNotRegistered, err)
				})
			})
		})
	})

	newManagerScenario(t, false, func(t *testing.T, ctx *managerScenarioCtx) {
		t.Run("Unmanaged plugin scenario", func(t *testing.T) {
			ctx.license.edition = "Open Source"
			ctx.license.hasLicense = false
			ctx.cfg.BuildVersion = "7.0.0"

			t.Run("Should be able to register plugin", func(t *testing.T) {
				err := ctx.manager.RegisterAndStart(context.Background(), testPluginID, ctx.factory)
				require.NoError(t, err)
				require.True(t, ctx.manager.IsRegistered(testPluginID))
				require.False(t, ctx.plugin.managed)

				t.Run("When manager runs should not start plugin", func(t *testing.T) {
					pCtx := context.Background()
					cCtx, cancel := context.WithCancel(pCtx)
					var wg sync.WaitGroup
					wg.Add(1)
					var runErr error
					go func() {
						runErr = ctx.manager.Run(cCtx)
						wg.Done()
					}()
					go func() {
						cancel()
					}()
					wg.Wait()
					require.Equal(t, context.Canceled, runErr)
					require.Equal(t, 0, ctx.plugin.startCount)
					require.Equal(t, 1, ctx.plugin.stopCount)
				})

				t.Run("Should be able to start unmanaged plugin and be restarted when process is killed", func(t *testing.T) {
					pCtx := context.Background()
					cCtx, cancel := context.WithCancel(pCtx)
					defer cancel()
					err := ctx.manager.StartPlugin(cCtx, testPluginID)
					require.Nil(t, err)
					require.Equal(t, 1, ctx.plugin.startCount)

					var wg sync.WaitGroup
					wg.Add(1)
					go func() {
						ctx.plugin.kill()
						for {
							if !ctx.plugin.Exited() {
								break
							}
						}
						wg.Done()
					}()
					wg.Wait()
					require.Equal(t, 2, ctx.plugin.startCount)
				})
			})
		})
	})

	newManagerScenario(t, true, func(t *testing.T, ctx *managerScenarioCtx) {
		t.Run("Plugin registration scenario when Grafana is licensed", func(t *testing.T) {
			ctx.license.edition = "Enterprise"
			ctx.license.hasLicense = true
			ctx.license.tokenRaw = "testtoken"
			ctx.cfg.BuildVersion = "7.0.0"
			ctx.cfg.EnterpriseLicensePath = "/license.txt"

			err := ctx.manager.RegisterAndStart(context.Background(), testPluginID, ctx.factory)
			require.NoError(t, err)

			t.Run("Should provide expected host environment variables", func(t *testing.T) {
				require.Len(t, ctx.env, 9)
				require.EqualValues(t, []string{
					"GF_VERSION=7.0.0",
					"GF_EDITION=Enterprise",
					"GF_ENTERPRISE_LICENSE_PATH=/license.txt",
					"GF_ENTERPRISE_LICENSE_TEXT=testtoken",
					fmt.Sprintf("%s=true", awsds.AssumeRoleEnabledEnvVarKeyName),
					fmt.Sprintf("%s=keys,credentials", awsds.AllowedAuthProvidersEnvVarKeyName),
					"AZURE_CLOUD=AzureCloud",
					"AZURE_MANAGED_IDENTITY_CLIENT_ID=client-id",
					"AZURE_MANAGED_IDENTITY_ENABLED=true"},
					ctx.env)
			})
		})
	})
}

type managerScenarioCtx struct {
	cfg     *setting.Cfg
	license *testLicensingService
	manager *Manager
	factory backendplugin.PluginFactoryFunc
	plugin  *testPlugin
	env     []string
}

func newManagerScenario(t *testing.T, managed bool, fn func(t *testing.T, ctx *managerScenarioCtx)) {
	t.Helper()
	cfg := setting.NewCfg()
	cfg.AWSAllowedAuthProviders = []string{"keys", "credentials"}
	cfg.AWSAssumeRoleEnabled = true

	cfg.Azure.ManagedIdentityEnabled = true
	cfg.Azure.Cloud = "AzureCloud"
	cfg.Azure.ManagedIdentityClientId = "client-id"

	license := &testLicensingService{}
	validator := &testPluginRequestValidator{}
	ctx := &managerScenarioCtx{
		cfg:     cfg,
		license: license,
		manager: &Manager{
			Cfg:                    cfg,
			License:                license,
			PluginRequestValidator: validator,
			logger:                 log.New("test"),
			plugins:                map[string]backendplugin.Plugin{},
		},
	}

	ctx.factory = func(pluginID string, logger log.Logger, env []string) (backendplugin.Plugin, error) {
		ctx.plugin = &testPlugin{
			pluginID: pluginID,
			logger:   logger,
			managed:  managed,
		}
		ctx.env = env

		return ctx.plugin, nil
	}

	fn(t, ctx)
}

type testPlugin struct {
	pluginID       string
	logger         log.Logger
	startCount     int
	stopCount      int
	managed        bool
	exited         bool
	decommissioned bool
	backend.CollectMetricsHandlerFunc
	backend.CheckHealthHandlerFunc
	backend.QueryDataHandlerFunc
	backend.CallResourceHandlerFunc
	mutex sync.RWMutex
}

func (tp *testPlugin) PluginID() string {
	return tp.pluginID
}

func (tp *testPlugin) Logger() log.Logger {
	return tp.logger
}

func (tp *testPlugin) Start(ctx context.Context) error {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()
	tp.exited = false
	tp.startCount++
	return nil
}

func (tp *testPlugin) Stop(ctx context.Context) error {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()
	tp.stopCount++
	return nil
}

func (tp *testPlugin) IsManaged() bool {
	return tp.managed
}

func (tp *testPlugin) Exited() bool {
	tp.mutex.RLock()
	defer tp.mutex.RUnlock()
	return tp.exited
}

func (tp *testPlugin) Decommission() error {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	tp.decommissioned = true

	return nil
}

func (tp *testPlugin) IsDecommissioned() bool {
	tp.mutex.RLock()
	defer tp.mutex.RUnlock()
	return tp.decommissioned
}

func (tp *testPlugin) kill() {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()
	tp.exited = true
}

func (tp *testPlugin) CollectMetrics(ctx context.Context) (*backend.CollectMetricsResult, error) {
	if tp.CollectMetricsHandlerFunc != nil {
		return tp.CollectMetricsHandlerFunc(ctx)
	}

	return nil, backendplugin.ErrMethodNotImplemented
}

func (tp *testPlugin) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if tp.CheckHealthHandlerFunc != nil {
		return tp.CheckHealthHandlerFunc(ctx, req)
	}

	return nil, backendplugin.ErrMethodNotImplemented
}

func (tp *testPlugin) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	if tp.QueryDataHandlerFunc != nil {
		return tp.QueryDataHandlerFunc(ctx, req)
	}

	return nil, backendplugin.ErrMethodNotImplemented
}

func (tp *testPlugin) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	if tp.CallResourceHandlerFunc != nil {
		return tp.CallResourceHandlerFunc(ctx, req, sender)
	}

	return backendplugin.ErrMethodNotImplemented
}

func (tp *testPlugin) SubscribeStream(ctx context.Context, request *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	return nil, backendplugin.ErrMethodNotImplemented
}

func (tp *testPlugin) PublishStream(ctx context.Context, request *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	return nil, backendplugin.ErrMethodNotImplemented
}

func (tp *testPlugin) RunStream(ctx context.Context, request *backend.RunStreamRequest, sender *backend.StreamSender) error {
	return backendplugin.ErrMethodNotImplemented
}

type testLicensingService struct {
	edition    string
	hasLicense bool
	tokenRaw   string
}

func (t *testLicensingService) HasLicense() bool {
	return t.hasLicense
}

func (t *testLicensingService) Expiry() int64 {
	return 0
}

func (t *testLicensingService) Edition() string {
	return t.edition
}

func (t *testLicensingService) StateInfo() string {
	return ""
}

func (t *testLicensingService) ContentDeliveryPrefix() string {
	return ""
}

func (t *testLicensingService) LicenseURL(user *models.SignedInUser) string {
	return ""
}

func (t *testLicensingService) HasValidLicense() bool {
	return false
}

func (t *testLicensingService) Environment() map[string]string {
	return map[string]string{"GF_ENTERPRISE_LICENSE_TEXT": t.tokenRaw}
}

type testPluginRequestValidator struct{}

func (t *testPluginRequestValidator) Validate(string, *http.Request) error {
	return nil
}
