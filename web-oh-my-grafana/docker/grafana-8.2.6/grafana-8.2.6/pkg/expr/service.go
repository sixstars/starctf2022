package expr

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/setting"
)

// DatasourceName is the string constant used as the datasource name in requests
// to identify it as an expression command.
const DatasourceName = "__expr__"

// DatasourceID is the fake datasource id used in requests to identify it as an
// expression command.
const DatasourceID = -100

// DatasourceUID is the fake datasource uid used in requests to identify it as an
// expression command.
const DatasourceUID = "-100"

// Service is service representation for expression handling.
type Service struct {
	Cfg         *setting.Cfg
	DataService plugins.DataRequestHandler
}

func (s *Service) isDisabled() bool {
	if s.Cfg == nil {
		return true
	}
	return !s.Cfg.ExpressionsEnabled
}

// BuildPipeline builds a pipeline from a request.
func (s *Service) BuildPipeline(req *Request) (DataPipeline, error) {
	return s.buildPipeline(req)
}

// ExecutePipeline executes an expression pipeline and returns all the results.
func (s *Service) ExecutePipeline(ctx context.Context, pipeline DataPipeline) (*backend.QueryDataResponse, error) {
	res := backend.NewQueryDataResponse()
	vars, err := pipeline.execute(ctx, s)
	if err != nil {
		return nil, err
	}
	for refID, val := range vars {
		res.Responses[refID] = backend.DataResponse{
			Frames: val.Values.AsDataFrames(refID),
		}
	}
	return res, nil
}
