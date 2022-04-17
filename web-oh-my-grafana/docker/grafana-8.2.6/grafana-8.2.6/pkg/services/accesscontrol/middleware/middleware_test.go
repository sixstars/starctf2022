package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/macaron.v1"

	"github.com/stretchr/testify/assert"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/accesscontrol/mock"
)

type middlewareTestCase struct {
	desc           string
	expectFallback bool
	expectEndpoint bool
	evaluator      accesscontrol.Evaluator
	ac             accesscontrol.AccessControl
}

func TestMiddleware(t *testing.T) {
	tests := []middlewareTestCase{
		{
			desc:           "should use fallback if access control is disabled",
			ac:             mock.New().WithDisabled(),
			expectFallback: true,
			expectEndpoint: true,
		},
		{
			desc: "should pass middleware for correct permissions",
			ac: mock.New().WithPermissions(
				[]*accesscontrol.Permission{{Action: "users:read", Scope: "users:*"}},
			),
			evaluator:      accesscontrol.EvalPermission("users:read", "users:*"),
			expectFallback: false,
			expectEndpoint: true,
		},
		{
			desc: "should not reach endpoint when missing permissions",
			ac: mock.New().WithPermissions(
				[]*accesscontrol.Permission{{Action: "users:read", Scope: "users:1"}},
			),
			evaluator:      accesscontrol.EvalPermission("users:read", "users:*"),
			expectFallback: false,
			expectEndpoint: false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			fallbackCalled := false
			fallback := func(c *models.ReqContext) {
				fallbackCalled = true
			}

			server := macaron.New()
			server.UseMiddleware(macaron.Renderer("../../public/views", "[[", "]]"))

			server.Use(contextProvider())
			server.Use(Middleware(test.ac)(fallback, test.evaluator))

			endpointCalled := false
			server.Get("/", func(c *models.ReqContext) {
				endpointCalled = true
			})

			request, err := http.NewRequest(http.MethodGet, "/", nil)
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, request)

			assert.Equal(t, test.expectFallback, fallbackCalled)
			assert.Equal(t, test.expectEndpoint, endpointCalled)
		})
	}
}

func contextProvider() macaron.Handler {
	return func(c *macaron.Context) {
		reqCtx := &models.ReqContext{
			Context:      c,
			Logger:       log.New(""),
			SignedInUser: &models.SignedInUser{},
			IsSignedIn:   true,
			SkipCache:    true,
		}
		c.Map(reqCtx)
	}
}
