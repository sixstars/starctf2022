package testdatasource

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/util/errutil"
)

const (
	randomWalkQuery                   queryType = "random_walk"
	randomWalkSlowQuery               queryType = "slow_query"
	randomWalkWithErrorQuery          queryType = "random_walk_with_error"
	randomWalkTableQuery              queryType = "random_walk_table"
	exponentialHeatmapBucketDataQuery queryType = "exponential_heatmap_bucket_data"
	linearHeatmapBucketDataQuery      queryType = "linear_heatmap_bucket_data"
	noDataPointsQuery                 queryType = "no_data_points"
	datapointsOutsideRangeQuery       queryType = "datapoints_outside_range"
	csvMetricValuesQuery              queryType = "csv_metric_values"
	predictablePulseQuery             queryType = "predictable_pulse"
	predictableCSVWaveQuery           queryType = "predictable_csv_wave"
	streamingClientQuery              queryType = "streaming_client"
	flightPath                        queryType = "flight_path"
	usaQueryKey                       queryType = "usa"
	liveQuery                         queryType = "live"
	grafanaAPIQuery                   queryType = "grafana_api"
	arrowQuery                        queryType = "arrow"
	annotationsQuery                  queryType = "annotations"
	tableStaticQuery                  queryType = "table_static"
	serverError500Query               queryType = "server_error_500"
	logsQuery                         queryType = "logs"
	nodeGraphQuery                    queryType = "node_graph"
	csvFileQueryType                  queryType = "csv_file"
	csvContentQueryType               queryType = "csv_content"
)

type queryType string

type Scenario struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	StringInput string `json:"stringInput"`
	Description string `json:"description"`
	handler     backend.QueryDataHandlerFunc
}

func (p *TestDataPlugin) registerScenario(scenario *Scenario) {
	p.scenarios[scenario.ID] = scenario
	p.queryMux.HandleFunc(scenario.ID, scenario.handler)
}

func (p *TestDataPlugin) registerScenarios() {
	p.registerScenario(&Scenario{
		ID:      string(exponentialHeatmapBucketDataQuery),
		Name:    "Exponential heatmap bucket data",
		handler: p.handleExponentialHeatmapBucketDataScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(linearHeatmapBucketDataQuery),
		Name:    "Linear heatmap bucket data",
		handler: p.handleLinearHeatmapBucketDataScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(randomWalkQuery),
		Name:    "Random Walk",
		handler: p.handleRandomWalkScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(predictablePulseQuery),
		Name:    "Predictable Pulse",
		handler: p.handlePredictablePulseScenario,
		Description: `Predictable Pulse returns a pulse wave where there is a datapoint every timeStepSeconds.
The wave cycles at timeStepSeconds*(onCount+offCount).
The cycle of the wave is based off of absolute time (from the epoch) which makes it predictable.
Timestamps will line up evenly on timeStepSeconds (For example, 60 seconds means times will all end in :00 seconds).`,
	})

	p.registerScenario(&Scenario{
		ID:      string(predictableCSVWaveQuery),
		Name:    "Predictable CSV Wave",
		handler: p.handlePredictableCSVWaveScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(randomWalkTableQuery),
		Name:    "Random Walk Table",
		handler: p.handleRandomWalkTableScenario,
	})

	p.registerScenario(&Scenario{
		ID:          string(randomWalkSlowQuery),
		Name:        "Slow Query",
		StringInput: "5s",
		handler:     p.handleRandomWalkSlowScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(noDataPointsQuery),
		Name:    "No Data Points",
		handler: p.handleClientSideScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(datapointsOutsideRangeQuery),
		Name:    "Datapoints Outside Range",
		handler: p.handleDatapointsOutsideRangeScenario,
	})

	p.registerScenario(&Scenario{
		ID:          string(csvMetricValuesQuery),
		Name:        "CSV Metric Values",
		StringInput: "1,20,90,30,5,0",
		handler:     p.handleCSVMetricValuesScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(streamingClientQuery),
		Name:    "Streaming Client",
		handler: p.handleClientSideScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(liveQuery),
		Name:    "Grafana Live",
		handler: p.handleClientSideScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(flightPath),
		Name:    "Flight path",
		handler: p.handleFlightPathScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(usaQueryKey),
		Name:    "USA generated data",
		handler: p.handleUSAScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(grafanaAPIQuery),
		Name:    "Grafana API",
		handler: p.handleClientSideScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(arrowQuery),
		Name:    "Load Apache Arrow Data",
		handler: p.handleArrowScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(annotationsQuery),
		Name:    "Annotations",
		handler: p.handleClientSideScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(tableStaticQuery),
		Name:    "Table Static",
		handler: p.handleTableStaticScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(randomWalkWithErrorQuery),
		Name:    "Random Walk (with error)",
		handler: p.handleRandomWalkWithErrorScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(serverError500Query),
		Name:    "Server Error (500)",
		handler: p.handleServerError500Scenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(logsQuery),
		Name:    "Logs",
		handler: p.handleLogsScenario,
	})

	p.registerScenario(&Scenario{
		ID:   string(nodeGraphQuery),
		Name: "Node Graph",
	})

	p.registerScenario(&Scenario{
		ID:      string(csvFileQueryType),
		Name:    "CSV File",
		handler: p.handleCsvFileScenario,
	})

	p.registerScenario(&Scenario{
		ID:      string(csvContentQueryType),
		Name:    "CSV Content",
		handler: p.handleCsvContentScenario,
	})

	p.queryMux.HandleFunc("", p.handleFallbackScenario)
}

// handleFallbackScenario handles the scenario where queryType is not set and fallbacks to scenarioId.
func (p *TestDataPlugin) handleFallbackScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	scenarioQueries := map[string][]backend.DataQuery{}

	for _, q := range req.Queries {
		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			p.logger.Error("Failed to unmarshal query model to JSON", "error", err)
			continue
		}

		scenarioID := model.Get("scenarioId").MustString(string(randomWalkQuery))
		if _, exist := p.scenarios[scenarioID]; exist {
			if _, ok := scenarioQueries[scenarioID]; !ok {
				scenarioQueries[scenarioID] = []backend.DataQuery{}
			}

			scenarioQueries[scenarioID] = append(scenarioQueries[scenarioID], q)
		} else {
			p.logger.Error("Scenario not found", "scenarioId", scenarioID)
		}
	}

	resp := backend.NewQueryDataResponse()
	for scenarioID, queries := range scenarioQueries {
		if scenario, exist := p.scenarios[scenarioID]; exist {
			sReq := &backend.QueryDataRequest{
				PluginContext: req.PluginContext,
				Headers:       req.Headers,
				Queries:       queries,
			}
			if sResp, err := scenario.handler(ctx, sReq); err != nil {
				p.logger.Error("Failed to handle scenario", "scenarioId", scenarioID, "error", err)
			} else {
				for refID, dr := range sResp.Responses {
					resp.Responses[refID] = dr
				}
			}
		}
	}

	return resp, nil
}

func (p *TestDataPlugin) handleRandomWalkScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			continue
		}
		seriesCount := model.Get("seriesCount").MustInt(1)

		for i := 0; i < seriesCount; i++ {
			respD := resp.Responses[q.RefID]
			respD.Frames = append(respD.Frames, RandomWalk(q, model, i))
			resp.Responses[q.RefID] = respD
		}
	}

	return resp, nil
}

func (p *TestDataPlugin) handleDatapointsOutsideRangeScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			continue
		}

		frame := newSeriesForQuery(q, model, 0)
		outsideTime := q.TimeRange.From.Add(-1 * time.Hour)
		frame.Fields = data.Fields{
			data.NewField(data.TimeSeriesTimeFieldName, nil, []time.Time{outsideTime}),
			data.NewField(data.TimeSeriesValueFieldName, nil, []float64{10}),
		}

		respD := resp.Responses[q.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handleCSVMetricValuesScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			continue
		}

		stringInput := model.Get("stringInput").MustString()

		valueField, err := csvLineToField(stringInput)
		if err != nil {
			return nil, err
		}
		valueField.Name = frameNameForQuery(q, model, 0)

		timeField := data.NewFieldFromFieldType(data.FieldTypeTime, valueField.Len())
		timeField.Name = "time"

		startTime := q.TimeRange.From.UnixNano() / int64(time.Millisecond)
		endTime := q.TimeRange.To.UnixNano() / int64(time.Millisecond)
		count := valueField.Len()
		var step int64 = 0
		if count > 1 {
			step = (endTime - startTime) / int64(count-1)
		}

		for i := 0; i < count; i++ {
			t := time.Unix(startTime/int64(1e+3), (startTime%int64(1e+3))*int64(1e+6))
			timeField.Set(i, t)
			startTime += step
		}

		frame := data.NewFrame("", timeField, valueField)

		respD := resp.Responses[q.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handleRandomWalkWithErrorScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			continue
		}

		respD := resp.Responses[q.RefID]
		respD.Frames = append(respD.Frames, RandomWalk(q, model, 0))
		respD.Error = fmt.Errorf("this is an error and it can include URLs http://grafana.com/")
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handleRandomWalkSlowScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			continue
		}

		stringInput := model.Get("stringInput").MustString()
		parsedInterval, _ := time.ParseDuration(stringInput)
		time.Sleep(parsedInterval)

		respD := resp.Responses[q.RefID]
		respD.Frames = append(respD.Frames, RandomWalk(q, model, 0))
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handleRandomWalkTableScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			continue
		}

		respD := resp.Responses[q.RefID]
		respD.Frames = append(respD.Frames, randomWalkTable(q, model))
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handlePredictableCSVWaveScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			return nil, err
		}

		respD := resp.Responses[q.RefID]
		frames, err := predictableCSVWave(q, model)
		if err != nil {
			return nil, err
		}
		respD.Frames = append(respD.Frames, frames...)
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handlePredictablePulseScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			continue
		}

		respD := resp.Responses[q.RefID]
		frame, err := predictablePulse(q, model)
		if err != nil {
			continue
		}
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handleServerError500Scenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	panic("Test Data Panic!")
}

func (p *TestDataPlugin) handleClientSideScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return backend.NewQueryDataResponse(), nil
}

func (p *TestDataPlugin) handleArrowScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			return nil, err
		}

		respD := resp.Responses[q.RefID]
		frame, err := doArrowQuery(q, model)
		if err != nil {
			return nil, err
		}
		if frame == nil {
			continue
		}
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handleExponentialHeatmapBucketDataScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		respD := resp.Responses[q.RefID]
		frame := randomHeatmapData(q, func(index int) float64 {
			return math.Exp2(float64(index))
		})
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handleLinearHeatmapBucketDataScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		respD := resp.Responses[q.RefID]
		frame := randomHeatmapData(q, func(index int) float64 {
			return float64(index * 10)
		})
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handleTableStaticScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		timeWalkerMs := q.TimeRange.From.UnixNano() / int64(time.Millisecond)
		to := q.TimeRange.To.UnixNano() / int64(time.Millisecond)
		step := q.Interval.Milliseconds()

		frame := data.NewFrame(q.RefID,
			data.NewField("Time", nil, []time.Time{}),
			data.NewField("Message", nil, []string{}),
			data.NewField("Description", nil, []string{}),
			data.NewField("Value", nil, []float64{}),
		)

		for i := int64(0); i < 10 && timeWalkerMs < to; i++ {
			t := time.Unix(timeWalkerMs/int64(1e+3), (timeWalkerMs%int64(1e+3))*int64(1e+6))
			frame.AppendRow(t, "This is a message", "Description", 23.1)
			timeWalkerMs += step
		}

		respD := resp.Responses[q.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func (p *TestDataPlugin) handleLogsScenario(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		from := q.TimeRange.From.UnixNano() / int64(time.Millisecond)
		to := q.TimeRange.To.UnixNano() / int64(time.Millisecond)

		model, err := simplejson.NewJson(q.JSON)
		if err != nil {
			continue
		}

		lines := model.Get("lines").MustInt64(10)
		includeLevelColumn := model.Get("levelColumn").MustBool(false)

		logLevelGenerator := newRandomStringProvider([]string{
			"emerg",
			"alert",
			"crit",
			"critical",
			"warn",
			"warning",
			"err",
			"eror",
			"error",
			"info",
			"notice",
			"dbug",
			"debug",
			"trace",
			"",
		})
		containerIDGenerator := newRandomStringProvider([]string{
			"f36a9eaa6d34310686f2b851655212023a216de955cbcc764210cefa71179b1a",
			"5a354a630364f3742c602f315132e16def594fe68b1e4a195b2fce628e24c97a",
		})
		hostnameGenerator := newRandomStringProvider([]string{
			"srv-001",
			"srv-002",
		})

		frame := data.NewFrame(q.RefID,
			data.NewField("time", nil, []time.Time{}),
			data.NewField("message", nil, []string{}),
			data.NewField("container_id", nil, []string{}),
			data.NewField("hostname", nil, []string{}),
		).SetMeta(&data.FrameMeta{
			PreferredVisualization: "logs",
		})

		if includeLevelColumn {
			frame.Fields = append(frame.Fields, data.NewField("level", nil, []string{}))
		}

		for i := int64(0); i < lines && to > from; i++ {
			logLevel := logLevelGenerator.Next()
			timeFormatted := time.Unix(to/1000, 0).Format(time.RFC3339)
			lvlString := ""
			if !includeLevelColumn {
				lvlString = fmt.Sprintf("lvl=%s ", logLevel)
			}

			message := fmt.Sprintf("t=%s %smsg=\"Request Completed\" logger=context userId=1 orgId=1 uname=admin method=GET path=/api/datasources/proxy/152/api/prom/label status=502 remote_addr=[::1] time_ms=1 size=0 referer=\"http://localhost:3000/explore?left=%%5B%%22now-6h%%22,%%22now%%22,%%22Prometheus%%202.x%%22,%%7B%%7D,%%7B%%22ui%%22:%%5Btrue,true,true,%%22none%%22%%5D%%7D%%5D\"", timeFormatted, lvlString)
			containerID := containerIDGenerator.Next()
			hostname := hostnameGenerator.Next()

			t := time.Unix(to/int64(1e+3), (to%int64(1e+3))*int64(1e+6))

			if includeLevelColumn {
				frame.AppendRow(t, message, containerID, hostname, logLevel)
			} else {
				frame.AppendRow(t, message, containerID, hostname)
			}

			to -= q.Interval.Milliseconds()
		}

		respD := resp.Responses[q.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[q.RefID] = respD
	}

	return resp, nil
}

func RandomWalk(query backend.DataQuery, model *simplejson.Json, index int) *data.Frame {
	timeWalkerMs := query.TimeRange.From.UnixNano() / int64(time.Millisecond)
	to := query.TimeRange.To.UnixNano() / int64(time.Millisecond)
	startValue := model.Get("startValue").MustFloat64(rand.Float64() * 100)
	spread := model.Get("spread").MustFloat64(1)
	noise := model.Get("noise").MustFloat64(0)

	min, err := model.Get("min").Float64()
	hasMin := err == nil
	max, err := model.Get("max").Float64()
	hasMax := err == nil

	timeVec := make([]*time.Time, 0)
	floatVec := make([]*float64, 0)

	walker := startValue

	for i := int64(0); i < 10000 && timeWalkerMs < to; i++ {
		nextValue := walker + (rand.Float64() * noise)

		if hasMin && nextValue < min {
			nextValue = min
			walker = min
		}

		if hasMax && nextValue > max {
			nextValue = max
			walker = max
		}

		t := time.Unix(timeWalkerMs/int64(1e+3), (timeWalkerMs%int64(1e+3))*int64(1e+6))
		timeVec = append(timeVec, &t)
		floatVec = append(floatVec, &nextValue)

		walker += (rand.Float64() - 0.5) * spread
		timeWalkerMs += query.Interval.Milliseconds()
	}

	return data.NewFrame("",
		data.NewField("time", nil, timeVec),
		data.NewField(frameNameForQuery(query, model, index), parseLabels(model), floatVec),
	)
}

func randomWalkTable(query backend.DataQuery, model *simplejson.Json) *data.Frame {
	timeWalkerMs := query.TimeRange.From.UnixNano() / int64(time.Millisecond)
	to := query.TimeRange.To.UnixNano() / int64(time.Millisecond)
	withNil := model.Get("withNil").MustBool(false)
	walker := model.Get("startValue").MustFloat64(rand.Float64() * 100)
	spread := 2.5

	frame := data.NewFrame(query.RefID,
		data.NewField("Time", nil, []*time.Time{}),
		data.NewField("Value", nil, []*float64{}),
		data.NewField("Min", nil, []*float64{}),
		data.NewField("Max", nil, []*float64{}),
		data.NewField("Info", nil, []*string{}),
	)

	var info strings.Builder

	for i := int64(0); i < query.MaxDataPoints && timeWalkerMs < to; i++ {
		delta := rand.Float64() - 0.5
		walker += delta

		info.Reset()
		if delta > 0 {
			info.WriteString("up")
		} else {
			info.WriteString("down")
		}
		if math.Abs(delta) > .4 {
			info.WriteString(" fast")
		}

		t := time.Unix(timeWalkerMs/int64(1e+3), (timeWalkerMs%int64(1e+3))*int64(1e+6))
		val := walker
		min := walker - ((rand.Float64() * spread) + 0.01)
		max := walker + ((rand.Float64() * spread) + 0.01)
		infoString := info.String()

		vals := []*float64{&val, &min, &max}
		// Add some random null values
		if withNil && rand.Float64() > 0.8 {
			for i := range vals {
				if rand.Float64() > .2 {
					vals[i] = nil
				}
			}
		}

		frame.AppendRow(&t, vals[0], vals[1], vals[2], &infoString)

		timeWalkerMs += query.Interval.Milliseconds()
	}

	return frame
}

type pCSVOptions struct {
	TimeStep  int64  `json:"timeStep"`
	ValuesCSV string `json:"valuesCSV"`
	Labels    string `json:"labels"`
	Name      string `json:"name"`
}

func predictableCSVWave(query backend.DataQuery, model *simplejson.Json) ([]*data.Frame, error) {
	rawQueries, err := model.Get("csvWave").ToDB()
	if err != nil {
		return nil, err
	}

	queries := []pCSVOptions{}
	err = json.Unmarshal(rawQueries, &queries)
	if err != nil {
		return nil, err
	}

	frames := make([]*data.Frame, 0, len(queries))

	for _, subQ := range queries {
		var err error

		rawValues := strings.TrimRight(strings.TrimSpace(subQ.ValuesCSV), ",") // Strip Trailing Comma
		rawValesCSV := strings.Split(rawValues, ",")
		values := make([]*float64, len(rawValesCSV))

		for i, rawValue := range rawValesCSV {
			var val *float64
			rawValue = strings.TrimSpace(rawValue)

			switch rawValue {
			case "null":
				// val stays nil
			case "nan":
				f := math.NaN()
				val = &f
			default:
				f, err := strconv.ParseFloat(rawValue, 64)
				if err != nil {
					return nil, errutil.Wrapf(err, "failed to parse value '%v' into nullable float", rawValue)
				}
				val = &f
			}
			values[i] = val
		}

		subQ.TimeStep *= 1000 // Seconds to Milliseconds
		valuesLen := int64(len(values))
		getValue := func(mod int64) (*float64, error) {
			var i int64
			for i = 0; i < valuesLen; i++ {
				if mod == i*subQ.TimeStep {
					return values[i], nil
				}
			}
			return nil, fmt.Errorf("did not get value at point in waveform - should not be here")
		}
		fields, err := predictableSeries(query.TimeRange, subQ.TimeStep, valuesLen, getValue)
		if err != nil {
			return nil, err
		}

		frame := newSeriesForQuery(query, model, 0)
		frame.Fields = fields
		frame.Fields[1].Labels = parseLabelsString(subQ.Labels)
		if subQ.Name != "" {
			frame.Name = subQ.Name
		}
		frames = append(frames, frame)
	}
	return frames, nil
}

func predictableSeries(timeRange backend.TimeRange, timeStep, length int64, getValue func(mod int64) (*float64, error)) (data.Fields, error) {
	from := timeRange.From.UnixNano() / int64(time.Millisecond)
	to := timeRange.To.UnixNano() / int64(time.Millisecond)

	timeCursor := from - (from % timeStep) // Truncate Start
	wavePeriod := timeStep * length
	maxPoints := 10000 // Don't return too many points

	timeVec := make([]*time.Time, 0)
	floatVec := make([]*float64, 0)

	for i := 0; i < maxPoints && timeCursor < to; i++ {
		val, err := getValue(timeCursor % wavePeriod)
		if err != nil {
			return nil, err
		}

		t := time.Unix(timeCursor/int64(1e+3), (timeCursor%int64(1e+3))*int64(1e+6))
		timeVec = append(timeVec, &t)
		floatVec = append(floatVec, val)

		timeCursor += timeStep
	}

	return data.Fields{
		data.NewField(data.TimeSeriesTimeFieldName, nil, timeVec),
		data.NewField(data.TimeSeriesValueFieldName, nil, floatVec),
	}, nil
}

func predictablePulse(query backend.DataQuery, model *simplejson.Json) (*data.Frame, error) {
	// Process Input
	var timeStep int64
	var onCount int64
	var offCount int64
	var onValue *float64
	var offValue *float64

	options := model.Get("pulseWave")

	var err error
	if timeStep, err = options.Get("timeStep").Int64(); err != nil {
		return nil, fmt.Errorf("failed to parse timeStep value '%v' into integer: %v", options.Get("timeStep"), err)
	}
	if onCount, err = options.Get("onCount").Int64(); err != nil {
		return nil, fmt.Errorf("failed to parse onCount value '%v' into integer: %v", options.Get("onCount"), err)
	}
	if offCount, err = options.Get("offCount").Int64(); err != nil {
		return nil, fmt.Errorf("failed to parse offCount value '%v' into integer: %v", options.Get("offCount"), err)
	}

	onValue, err = fromStringOrNumber(options.Get("onValue"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse onValue value '%v' into float: %v", options.Get("onValue"), err)
	}
	offValue, err = fromStringOrNumber(options.Get("offValue"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse offValue value '%v' into float: %v", options.Get("offValue"), err)
	}

	timeStep *= 1000                             // Seconds to Milliseconds
	onFor := func(mod int64) (*float64, error) { // How many items in the cycle should get the on value
		var i int64
		for i = 0; i < onCount; i++ {
			if mod == i*timeStep {
				return onValue, nil
			}
		}
		return offValue, nil
	}
	fields, err := predictableSeries(query.TimeRange, timeStep, onCount+offCount, onFor)
	if err != nil {
		return nil, err
	}

	frame := newSeriesForQuery(query, model, 0)
	frame.Fields = fields
	frame.Fields[1].Labels = parseLabels(model)

	return frame, nil
}

func randomHeatmapData(query backend.DataQuery, fnBucketGen func(index int) float64) *data.Frame {
	frame := data.NewFrame("data", data.NewField("time", nil, []*time.Time{}))
	for i := 0; i < 10; i++ {
		frame.Fields = append(frame.Fields, data.NewField(strconv.FormatInt(int64(fnBucketGen(i)), 10), nil, []*float64{}))
	}

	timeWalkerMs := query.TimeRange.From.UnixNano() / int64(time.Millisecond)
	to := query.TimeRange.To.UnixNano() / int64(time.Millisecond)

	for j := int64(0); j < 100 && timeWalkerMs < to; j++ {
		t := time.Unix(timeWalkerMs/int64(1e+3), (timeWalkerMs%int64(1e+3))*int64(1e+6))
		vals := []interface{}{&t}
		for n := 1; n < len(frame.Fields); n++ {
			v := float64(rand.Int63n(100))
			vals = append(vals, &v)
		}
		frame.AppendRow(vals...)
		timeWalkerMs += query.Interval.Milliseconds() * 50
	}

	return frame
}

func doArrowQuery(query backend.DataQuery, model *simplejson.Json) (*data.Frame, error) {
	encoded := model.Get("stringInput").MustString("")
	if encoded == "" {
		return nil, nil
	}
	arrow, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return data.UnmarshalArrowFrame(arrow)
}

func newSeriesForQuery(query backend.DataQuery, model *simplejson.Json, index int) *data.Frame {
	alias := model.Get("alias").MustString("")
	suffix := ""

	if index > 0 {
		suffix = strconv.Itoa(index)
	}

	if alias == "" {
		alias = fmt.Sprintf("%s-series%s", query.RefID, suffix)
	}

	if alias == "__server_names" && len(serverNames) > index {
		alias = serverNames[index]
	}

	if alias == "__house_locations" && len(houseLocations) > index {
		alias = houseLocations[index]
	}

	return data.NewFrame(alias)
}

/**
 * Looks for a labels request and adds them as tags
 *
 * '{job="foo", instance="bar"} => {job: "foo", instance: "bar"}`
 */
func parseLabels(model *simplejson.Json) data.Labels {
	labelText := model.Get("labels").MustString("")
	return parseLabelsString(labelText)
}

func parseLabelsString(labelText string) data.Labels {
	if labelText == "" {
		return data.Labels{}
	}

	text := strings.Trim(labelText, `{}`)
	if len(text) < 2 {
		return data.Labels{}
	}

	tags := make(data.Labels)

	for _, keyval := range strings.Split(text, ",") {
		idx := strings.Index(keyval, "=")
		key := strings.TrimSpace(keyval[:idx])
		val := strings.TrimSpace(keyval[idx+1:])
		val = strings.Trim(val, "\"")
		tags[key] = val
	}

	return tags
}

func frameNameForQuery(query backend.DataQuery, model *simplejson.Json, index int) string {
	name := model.Get("alias").MustString("")
	suffix := ""

	if index > 0 {
		suffix = strconv.Itoa(index)
	}

	if name == "" {
		name = fmt.Sprintf("%s-series%s", query.RefID, suffix)
	}

	if name == "__server_names" && len(serverNames) > index {
		name = serverNames[index]
	}

	if name == "__house_locations" && len(houseLocations) > index {
		name = houseLocations[index]
	}

	return name
}

func fromStringOrNumber(val *simplejson.Json) (*float64, error) {
	switch v := val.Interface().(type) {
	case json.Number:
		fV, err := v.Float64()
		if err != nil {
			return nil, err
		}
		return &fV, nil
	case string:
		switch v {
		case "null":
			return nil, nil
		case "nan":
			v := math.NaN()
			return &v, nil
		default:
			return nil, fmt.Errorf("failed to extract value from %v", v)
		}
	default:
		return nil, fmt.Errorf("failed to extract value")
	}
}

var serverNames = []string{
	"Backend-ops-01",
	"Backend-ops-02",
	"Backend-ops-03",
	"Backend-ops-04",
	"Frontend-web-01",
	"Frontend-web-02",
	"Frontend-web-03",
	"Frontend-web-04",
	"MySQL-01",
	"MySQL-02",
	"MySQL-03",
	"MySQL-04",
	"Postgres-01",
	"Postgres-02",
	"Postgres-03",
	"Postgres-04",
	"DB-01",
	"DB-02",
	"SAN-01",
	"SAN-02",
	"SAN-02",
	"SAN-04",
	"Kaftka-01",
	"Kaftka-02",
	"Kaftka-03",
	"Zookeeper-01",
	"Zookeeper-02",
	"Zookeeper-03",
	"Zookeeper-04",
}

var houseLocations = []string{
	"Cellar",
	"Living room",
	"Porch",
	"Bedroom",
	"Guest room",
	"Kitchen",
	"Playroom",
	"Bathroom",
	"Outside",
	"Roof",
	"Terrace",
}
