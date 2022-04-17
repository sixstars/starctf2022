package opentsdb

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/httpclient"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/plugins/backendplugin"
	"github.com/grafana/grafana/pkg/plugins/backendplugin/coreplugin"
	"github.com/grafana/grafana/pkg/setting"
	"golang.org/x/net/context/ctxhttp"
)

type Service struct {
	logger log.Logger
	im     instancemgmt.InstanceManager
}

func ProvideService(httpClientProvider httpclient.Provider, manager backendplugin.Manager) (*Service, error) {
	im := datasource.NewInstanceManager(newInstanceSettings(httpClientProvider))
	s := &Service{
		logger: log.New("tsdb.opentsdb"),
		im:     im,
	}

	factory := coreplugin.New(backend.ServeOpts{
		QueryDataHandler: s,
	})
	err := manager.RegisterAndStart(context.Background(), "opentsdb", factory)
	if err != nil {
		return nil, err
	}

	return s, nil
}

type datasourceInfo struct {
	HTTPClient *http.Client
	URL        string
}

type DsAccess string

func newInstanceSettings(httpClientProvider httpclient.Provider) datasource.InstanceFactoryFunc {
	return func(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
		opts, err := settings.HTTPClientOptions()
		if err != nil {
			return nil, err
		}

		client, err := httpClientProvider.New(opts)
		if err != nil {
			return nil, err
		}

		model := &datasourceInfo{
			HTTPClient: client,
			URL:        settings.URL,
		}

		return model, nil
	}
}

func (s *Service) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	var tsdbQuery OpenTsdbQuery

	q := req.Queries[0]

	tsdbQuery.Start = q.TimeRange.From.UnixNano() / int64(time.Millisecond)
	tsdbQuery.End = q.TimeRange.To.UnixNano() / int64(time.Millisecond)

	for _, query := range req.Queries {
		metric := s.buildMetric(query)
		tsdbQuery.Queries = append(tsdbQuery.Queries, metric)
	}

	// TODO: Don't use global variable
	if setting.Env == setting.Dev {
		s.logger.Debug("OpenTsdb request", "params", tsdbQuery)
	}

	dsInfo, err := s.getDSInfo(req.PluginContext)
	if err != nil {
		return nil, err
	}

	request, err := s.createRequest(dsInfo, tsdbQuery)
	if err != nil {
		return &backend.QueryDataResponse{}, err
	}

	res, err := ctxhttp.Do(ctx, dsInfo.HTTPClient, request)
	if err != nil {
		return &backend.QueryDataResponse{}, err
	}

	result, err := s.parseResponse(res)
	if err != nil {
		return &backend.QueryDataResponse{}, err
	}

	return result, nil
}

func (s *Service) createRequest(dsInfo *datasourceInfo, data OpenTsdbQuery) (*http.Request, error) {
	u, err := url.Parse(dsInfo.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "api/query")

	postData, err := json.Marshal(data)
	if err != nil {
		s.logger.Info("Failed marshaling data", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(string(postData)))
	if err != nil {
		s.logger.Info("Failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (s *Service) parseResponse(res *http.Response) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			s.logger.Warn("Failed to close response body", "err", err)
		}
	}()

	if res.StatusCode/100 != 2 {
		s.logger.Info("Request failed", "status", res.Status, "body", string(body))
		return nil, fmt.Errorf("request failed, status: %s", res.Status)
	}

	var responseData []OpenTsdbResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		s.logger.Info("Failed to unmarshal opentsdb response", "error", err, "status", res.Status, "body", string(body))
		return nil, err
	}

	frames := data.Frames{}
	for _, val := range responseData {
		timeVector := make([]time.Time, 0, len(val.DataPoints))
		values := make([]float64, 0, len(val.DataPoints))
		name := val.Metric

		for timeString, value := range val.DataPoints {
			timestamp, err := strconv.ParseInt(timeString, 10, 64)
			if err != nil {
				s.logger.Info("Failed to unmarshal opentsdb timestamp", "timestamp", timeString)
				return nil, err
			}
			timeVector = append(timeVector, time.Unix(timestamp, 0).UTC())
			values = append(values, value)
		}
		frames = append(frames, data.NewFrame(name,
			data.NewField("time", nil, timeVector),
			data.NewField("value", nil, values)))
	}
	result := resp.Responses["A"]
	result.Frames = frames
	resp.Responses["A"] = result
	return resp, nil
}

func (s *Service) buildMetric(query backend.DataQuery) map[string]interface{} {
	metric := make(map[string]interface{})

	model, err := simplejson.NewJson(query.JSON)
	if err != nil {
		return nil
	}

	// Setting metric and aggregator
	metric["metric"] = model.Get("metric").MustString()
	metric["aggregator"] = model.Get("aggregator").MustString()

	// Setting downsampling options
	disableDownsampling := model.Get("disableDownsampling").MustBool()
	if !disableDownsampling {
		downsampleInterval := model.Get("downsampleInterval").MustString()
		if downsampleInterval == "" {
			downsampleInterval = "1m" // default value for blank
		}
		downsample := downsampleInterval + "-" + model.Get("downsampleAggregator").MustString()
		if model.Get("downsampleFillPolicy").MustString() != "none" {
			metric["downsample"] = downsample + "-" + model.Get("downsampleFillPolicy").MustString()
		} else {
			metric["downsample"] = downsample
		}
	}

	// Setting rate options
	if model.Get("shouldComputeRate").MustBool() {
		metric["rate"] = true
		rateOptions := make(map[string]interface{})
		rateOptions["counter"] = model.Get("isCounter").MustBool()

		counterMax, counterMaxCheck := model.CheckGet("counterMax")
		if counterMaxCheck {
			rateOptions["counterMax"] = counterMax.MustFloat64()
		}

		resetValue, resetValueCheck := model.CheckGet("counterResetValue")
		if resetValueCheck {
			rateOptions["resetValue"] = resetValue.MustFloat64()
		}

		if !counterMaxCheck && (!resetValueCheck || resetValue.MustFloat64() == 0) {
			rateOptions["dropResets"] = true
		}

		metric["rateOptions"] = rateOptions
	}

	// Setting tags
	tags, tagsCheck := model.CheckGet("tags")
	if tagsCheck && len(tags.MustMap()) > 0 {
		metric["tags"] = tags.MustMap()
	}

	// Setting filters
	filters, filtersCheck := model.CheckGet("filters")
	if filtersCheck && len(filters.MustArray()) > 0 {
		metric["filters"] = filters.MustArray()
	}

	return metric
}

func (s *Service) getDSInfo(pluginCtx backend.PluginContext) (*datasourceInfo, error) {
	i, err := s.im.Get(pluginCtx)
	if err != nil {
		return nil, err
	}

	instance, ok := i.(*datasourceInfo)
	if !ok {
		return nil, fmt.Errorf("failed to cast datsource info")
	}

	return instance, nil
}
