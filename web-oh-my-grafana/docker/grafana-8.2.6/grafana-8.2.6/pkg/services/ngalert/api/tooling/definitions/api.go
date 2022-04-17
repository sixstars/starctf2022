// Package definitions includes the types required for generating or consuming an OpenAPI
// spec for the Unified Alerting API.
// Documentation of the API.
//
//     Schemes: http, https
//     BasePath: /api/v1
//     Version: 1.1.0
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - basic
//
//    SecurityDefinitions:
//    basic:
//      type: basic
//
// swagger:meta
package definitions

// swagger:model
type ValidationError struct {
	Msg string `json:"msg"`
}

// swagger:model
type Ack struct{}

type Backend int

const (
	GrafanaBackend Backend = iota
	AlertmanagerBackend
	LoTexRulerBackend
)

func (b Backend) String() string {
	switch b {
	case GrafanaBackend:
		return "grafana"
	case AlertmanagerBackend:
		return "alertmanager"
	case LoTexRulerBackend:
		return "lotex"
	default:
		return ""
	}
}
