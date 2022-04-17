package plugins

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/grafana/grafana/pkg/plugins/backendplugin"
	"github.com/grafana/grafana/pkg/plugins/backendplugin/grpcplugin"
	"github.com/grafana/grafana/pkg/util/errutil"
)

// DataSourcePlugin contains all metadata about a datasource plugin
type DataSourcePlugin struct {
	FrontendPluginBase
	Annotations  bool              `json:"annotations"`
	Metrics      bool              `json:"metrics"`
	Alerting     bool              `json:"alerting"`
	Explore      bool              `json:"explore"`
	Table        bool              `json:"tables"`
	Logs         bool              `json:"logs"`
	Tracing      bool              `json:"tracing"`
	QueryOptions map[string]bool   `json:"queryOptions,omitempty"`
	BuiltIn      bool              `json:"builtIn,omitempty"`
	Mixed        bool              `json:"mixed,omitempty"`
	Routes       []*AppPluginRoute `json:"routes"`
	Streaming    bool              `json:"streaming"`

	Backend    bool   `json:"backend,omitempty"`
	Executable string `json:"executable,omitempty"`
	SDK        bool   `json:"sdk,omitempty"`
}

func (p *DataSourcePlugin) Load(decoder *json.Decoder, base *PluginBase, backendPluginManager backendplugin.Manager) (
	interface{}, error) {
	if err := decoder.Decode(p); err != nil {
		return nil, errutil.Wrapf(err, "Failed to decode datasource plugin")
	}

	if p.Backend {
		cmd := ComposePluginStartCommand(p.Executable)
		fullpath := filepath.Join(base.PluginDir, cmd)
		factory := grpcplugin.NewBackendPlugin(p.Id, fullpath)
		if err := backendPluginManager.RegisterAndStart(context.Background(), p.Id, factory); err != nil {
			return nil, errutil.Wrapf(err, "failed to register backend plugin")
		}
	}

	return p, nil
}
