package middleware

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/remotecache"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/auth"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	macaron "gopkg.in/macaron.v1"
)

func TestRecoveryMiddleware(t *testing.T) {
	t.Run("Given an API route that panics", func(t *testing.T) {
		apiURL := "/api/whatever"
		recoveryScenario(t, "recovery middleware should return JSON", apiURL, func(t *testing.T, sc *scenarioContext) {
			sc.handlerFunc = panicHandler
			sc.fakeReq("GET", apiURL).exec()
			sc.req.Header.Set("content-type", "application/json")

			assert.Equal(t, 500, sc.resp.Code)
			assert.Equal(t, "Internal Server Error - Check the Grafana server logs for the detailed error message.", sc.respJson["message"])
			assert.True(t, strings.HasPrefix(sc.respJson["error"].(string), "Server Error"))
		})
	})

	t.Run("Given a non-API route that panics", func(t *testing.T) {
		apiURL := "/whatever"
		recoveryScenario(t, "recovery middleware should return html", apiURL, func(t *testing.T, sc *scenarioContext) {
			sc.handlerFunc = panicHandler
			sc.fakeReq("GET", apiURL).exec()

			assert.Equal(t, 500, sc.resp.Code)
			assert.Equal(t, "text/html; charset=UTF-8", sc.resp.Header().Get("content-type"))
			assert.Contains(t, sc.resp.Body.String(), "<title>Grafana - Error</title>")
		})
	})
}

func panicHandler(c *models.ReqContext) {
	panic("Handler has panicked")
}

func recoveryScenario(t *testing.T, desc string, url string, fn scenarioFunc) {
	t.Run(desc, func(t *testing.T) {
		defer bus.ClearBusHandlers()

		cfg := setting.NewCfg()
		cfg.ErrTemplateName = "error-template"
		sc := &scenarioContext{
			t:   t,
			url: url,
			cfg: cfg,
		}

		viewsPath, err := filepath.Abs("../../public/views")
		require.NoError(t, err)

		sc.m = macaron.New()
		sc.m.Use(Recovery(cfg))

		sc.m.Use(AddDefaultResponseHeaders(cfg))
		sc.m.UseMiddleware(macaron.Renderer(viewsPath, "[[", "]]"))

		sc.userAuthTokenService = auth.NewFakeUserAuthTokenService()
		sc.remoteCacheService = remotecache.NewFakeStore(t)

		contextHandler := getContextHandler(t, nil)
		sc.m.Use(contextHandler.Middleware)
		// mock out gc goroutine
		sc.m.Use(OrgRedirect(cfg))

		sc.defaultHandler = func(c *models.ReqContext) {
			sc.context = c
			if sc.handlerFunc != nil {
				sc.handlerFunc(sc.context)
			}
		}

		sc.m.Get(url, sc.defaultHandler)

		fn(t, sc)
	})
}
