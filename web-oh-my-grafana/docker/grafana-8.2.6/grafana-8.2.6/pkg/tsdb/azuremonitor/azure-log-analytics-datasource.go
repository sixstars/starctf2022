package azuremonitor

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/util/errutil"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context/ctxhttp"
)

// AzureLogAnalyticsDatasource calls the Azure Log Analytics API's
type AzureLogAnalyticsDatasource struct {
	proxy serviceProxy
}

// AzureLogAnalyticsQuery is the query request that is built from the saved values for
// from the UI
type AzureLogAnalyticsQuery struct {
	RefID        string
	ResultFormat string
	URL          string
	JSON         json.RawMessage
	Params       url.Values
	Target       string
	TimeRange    backend.TimeRange
}

func (e *AzureLogAnalyticsDatasource) resourceRequest(rw http.ResponseWriter, req *http.Request, cli *http.Client) {
	e.proxy.Do(rw, req, cli)
}

// executeTimeSeriesQuery does the following:
// 1. build the AzureMonitor url and querystring for each query
// 2. executes each query by calling the Azure Monitor API
// 3. parses the responses for each query into data frames
func (e *AzureLogAnalyticsDatasource) executeTimeSeriesQuery(ctx context.Context, originalQueries []backend.DataQuery, dsInfo datasourceInfo, client *http.Client, url string) (*backend.QueryDataResponse, error) {
	result := backend.NewQueryDataResponse()

	queries, err := e.buildQueries(originalQueries, dsInfo)
	if err != nil {
		return nil, err
	}

	for _, query := range queries {
		result.Responses[query.RefID] = e.executeQuery(ctx, query, dsInfo, client, url)
	}

	return result, nil
}

func getApiURL(queryJSONModel logJSONQuery) string {
	// Legacy queries only specify a Workspace GUID, which we need to use the old workspace-centric
	// API URL for, and newer queries specifying a resource URI should use resource-centric API.
	// However, legacy workspace queries using a `workspaces()` template variable will be resolved
	// to a resource URI, so they should use the new resource-centric.
	azureLogAnalyticsTarget := queryJSONModel.AzureLogAnalytics
	var resourceOrWorkspace string

	if azureLogAnalyticsTarget.Resource != "" {
		resourceOrWorkspace = azureLogAnalyticsTarget.Resource
	} else {
		resourceOrWorkspace = azureLogAnalyticsTarget.Workspace
	}

	matchesResourceURI, _ := regexp.MatchString("^/subscriptions/", resourceOrWorkspace)

	if matchesResourceURI {
		return fmt.Sprintf("v1%s/query", resourceOrWorkspace)
	} else {
		return fmt.Sprintf("v1/workspaces/%s/query", resourceOrWorkspace)
	}
}

func (e *AzureLogAnalyticsDatasource) buildQueries(queries []backend.DataQuery, dsInfo datasourceInfo) ([]*AzureLogAnalyticsQuery, error) {
	azureLogAnalyticsQueries := []*AzureLogAnalyticsQuery{}

	for _, query := range queries {
		queryJSONModel := logJSONQuery{}
		err := json.Unmarshal(query.JSON, &queryJSONModel)
		if err != nil {
			return nil, fmt.Errorf("failed to decode the Azure Log Analytics query object from JSON: %w", err)
		}

		azureLogAnalyticsTarget := queryJSONModel.AzureLogAnalytics
		azlog.Debug("AzureLogAnalytics", "target", azureLogAnalyticsTarget)

		resultFormat := azureLogAnalyticsTarget.ResultFormat
		if resultFormat == "" {
			resultFormat = timeSeries
		}

		apiURL := getApiURL(queryJSONModel)

		params := url.Values{}
		rawQuery, err := KqlInterpolate(query, dsInfo, azureLogAnalyticsTarget.Query, "TimeGenerated")
		if err != nil {
			return nil, err
		}
		params.Add("query", rawQuery)

		azureLogAnalyticsQueries = append(azureLogAnalyticsQueries, &AzureLogAnalyticsQuery{
			RefID:        query.RefID,
			ResultFormat: resultFormat,
			URL:          apiURL,
			JSON:         query.JSON,
			Params:       params,
			Target:       params.Encode(),
			TimeRange:    query.TimeRange,
		})
	}

	return azureLogAnalyticsQueries, nil
}

func (e *AzureLogAnalyticsDatasource) executeQuery(ctx context.Context, query *AzureLogAnalyticsQuery, dsInfo datasourceInfo, client *http.Client, url string) backend.DataResponse {
	dataResponse := backend.DataResponse{}

	dataResponseErrorWithExecuted := func(err error) backend.DataResponse {
		dataResponse.Error = err
		dataResponse.Frames = data.Frames{
			&data.Frame{
				RefID: query.RefID,
				Meta: &data.FrameMeta{
					ExecutedQueryString: query.Params.Get("query"),
				},
			},
		}
		return dataResponse
	}

	// If azureLogAnalyticsSameAs is defined and set to false, return an error
	if sameAs, ok := dsInfo.JSONData["azureLogAnalyticsSameAs"]; ok && !sameAs.(bool) {
		return dataResponseErrorWithExecuted(fmt.Errorf("Log Analytics credentials are no longer supported. Go to the data source configuration to update Azure Monitor credentials")) //nolint:golint,stylecheck
	}

	req, err := e.createRequest(ctx, dsInfo, url)
	if err != nil {
		dataResponse.Error = err
		return dataResponse
	}

	req.URL.Path = path.Join(req.URL.Path, query.URL)
	req.URL.RawQuery = query.Params.Encode()

	span, ctx := opentracing.StartSpanFromContext(ctx, "azure log analytics query")
	span.SetTag("target", query.Target)
	span.SetTag("from", query.TimeRange.From.UnixNano()/int64(time.Millisecond))
	span.SetTag("until", query.TimeRange.To.UnixNano()/int64(time.Millisecond))
	span.SetTag("datasource_id", dsInfo.DatasourceID)
	span.SetTag("org_id", dsInfo.OrgID)

	defer span.Finish()

	if err := opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
		return dataResponseErrorWithExecuted(err)
	}

	azlog.Debug("AzureLogAnalytics", "Request ApiURL", req.URL.String())
	res, err := ctxhttp.Do(ctx, client, req)
	if err != nil {
		return dataResponseErrorWithExecuted(err)
	}

	logResponse, err := e.unmarshalResponse(res)
	if err != nil {
		return dataResponseErrorWithExecuted(err)
	}

	t, err := logResponse.GetPrimaryResultTable()
	if err != nil {
		return dataResponseErrorWithExecuted(err)
	}

	frame, err := ResponseTableToFrame(t)
	if err != nil {
		return dataResponseErrorWithExecuted(err)
	}

	model, err := simplejson.NewJson(query.JSON)
	if err != nil {
		return dataResponseErrorWithExecuted(err)
	}

	err = setAdditionalFrameMeta(frame,
		query.Params.Get("query"),
		model.Get("subscriptionId").MustString(),
		model.Get("azureLogAnalytics").Get("workspace").MustString())
	if err != nil {
		frame.AppendNotices(data.Notice{Severity: data.NoticeSeverityWarning, Text: "could not add custom metadata: " + err.Error()})
		azlog.Warn("failed to add custom metadata to azure log analytics response", err)
	}

	if query.ResultFormat == timeSeries {
		tsSchema := frame.TimeSeriesSchema()
		if tsSchema.Type == data.TimeSeriesTypeLong {
			wideFrame, err := data.LongToWide(frame, nil)
			if err == nil {
				frame = wideFrame
			} else {
				frame.AppendNotices(data.Notice{Severity: data.NoticeSeverityWarning, Text: "could not convert frame to time series, returning raw table: " + err.Error()})
			}
		}
	}

	dataResponse.Frames = data.Frames{frame}
	return dataResponse
}

func (e *AzureLogAnalyticsDatasource) createRequest(ctx context.Context, dsInfo datasourceInfo, url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		azlog.Debug("Failed to create request", "error", err)
		return nil, errutil.Wrap("failed to create request", err)
	}
	req.URL.Path = "/"
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// GetPrimaryResultTable returns the first table in the response named "PrimaryResult", or an
// error if there is no table by that name.
func (ar *AzureLogAnalyticsResponse) GetPrimaryResultTable() (*AzureResponseTable, error) {
	for _, t := range ar.Tables {
		if t.Name == "PrimaryResult" {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("no data as PrimaryResult table is missing from the response")
}

func (e *AzureLogAnalyticsDatasource) unmarshalResponse(res *http.Response) (AzureLogAnalyticsResponse, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return AzureLogAnalyticsResponse{}, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			azlog.Warn("Failed to close response body", "err", err)
		}
	}()

	if res.StatusCode/100 != 2 {
		azlog.Debug("Request failed", "status", res.Status, "body", string(body))
		return AzureLogAnalyticsResponse{}, fmt.Errorf("request failed, status: %s, body: %s", res.Status, string(body))
	}

	var data AzureLogAnalyticsResponse
	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	err = d.Decode(&data)
	if err != nil {
		azlog.Debug("Failed to unmarshal Azure Log Analytics response", "error", err, "status", res.Status, "body", string(body))
		return AzureLogAnalyticsResponse{}, err
	}

	return data, nil
}

// LogAnalyticsMeta is a type for the a Frame's Meta's Custom property.
type LogAnalyticsMeta struct {
	ColumnTypes  []string `json:"azureColumnTypes"`
	Subscription string   `json:"subscription"`
	Workspace    string   `json:"workspace"`
	EncodedQuery []byte   `json:"encodedQuery"` // EncodedQuery is used for deep links.
}

func setAdditionalFrameMeta(frame *data.Frame, query, subscriptionID, workspace string) error {
	frame.Meta.ExecutedQueryString = query
	la, ok := frame.Meta.Custom.(*LogAnalyticsMeta)
	if !ok {
		return fmt.Errorf("unexpected type found for frame's custom metadata")
	}
	la.Subscription = subscriptionID
	la.Workspace = workspace
	encodedQuery, err := encodeQuery(query)
	if err == nil {
		la.EncodedQuery = encodedQuery
		return nil
	}
	return fmt.Errorf("failed to encode the query into the encodedQuery property")
}

// encodeQuery encodes the query in gzip so the frontend can build links.
func encodeQuery(rawQuery string) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(rawQuery)); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
