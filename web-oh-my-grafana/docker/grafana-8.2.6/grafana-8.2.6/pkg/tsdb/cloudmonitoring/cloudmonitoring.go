package cloudmonitoring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/plugins"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/api/pluginproxy"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/httpclient"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"golang.org/x/oauth2/google"
)

var (
	slog = log.New("tsdb.cloudMonitoring")
)

var (
	matchAllCap                 = regexp.MustCompile("(.)([A-Z][a-z]*)")
	legendKeyFormat             = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)
	metricNameFormat            = regexp.MustCompile(`([\w\d_]+)\.(googleapis\.com|io)/(.+)`)
	wildcardRegexRe             = regexp.MustCompile(`[-\/^$+?.()|[\]{}]`)
	alignmentPeriodRe           = regexp.MustCompile("[0-9]+")
	cloudMonitoringUnitMappings = map[string]string{
		"bit":     "bits",
		"By":      "bytes",
		"s":       "s",
		"min":     "m",
		"h":       "h",
		"d":       "d",
		"us":      "µs",
		"ms":      "ms",
		"ns":      "ns",
		"%":       "percent",
		"percent": "percent",
		"MiBy":    "mbytes",
		"By/s":    "Bps",
		"GBy":     "decgbytes",
	}
)

const (
	gceAuthentication         string = "gce"
	jwtAuthentication         string = "jwt"
	metricQueryType           string = "metrics"
	sloQueryType              string = "slo"
	mqlEditorMode             string = "mql"
	crossSeriesReducerDefault string = "REDUCE_NONE"
	perSeriesAlignerDefault   string = "ALIGN_MEAN"
)

func ProvideService(cfg *setting.Cfg, pluginManager plugins.Manager, httpClientProvider httpclient.Provider) *Service {
	return &Service{
		PluginManager:      pluginManager,
		HTTPClientProvider: httpClientProvider,
		Cfg:                cfg,
	}
}

type Service struct {
	PluginManager      plugins.Manager
	HTTPClientProvider httpclient.Provider
	Cfg                *setting.Cfg
}

// Executor executes queries for the CloudMonitoring datasource.
type Executor struct {
	httpClient    *http.Client
	dsInfo        *models.DataSource
	pluginManager plugins.Manager
	cfg           *setting.Cfg
}

// NewExecutor returns an Executor.
//nolint: staticcheck // plugins.DataPlugin deprecated
func (s *Service) NewExecutor(dsInfo *models.DataSource) (plugins.DataPlugin, error) {
	httpClient, err := dsInfo.GetHTTPClient(s.HTTPClientProvider)
	if err != nil {
		return nil, err
	}

	return &Executor{
		httpClient:    httpClient,
		dsInfo:        dsInfo,
		pluginManager: s.PluginManager,
		cfg:           s.Cfg,
	}, nil
}

// Query takes in the frontend queries, parses them into the CloudMonitoring query format
// executes the queries against the CloudMonitoring API and parses the response into
// the time series or table format
//nolint: staticcheck // plugins.DataPlugin deprecated
func (e *Executor) DataQuery(ctx context.Context, dsInfo *models.DataSource, tsdbQuery plugins.DataQuery) (
	plugins.DataResponse, error) {
	var result plugins.DataResponse
	var err error
	queryType := tsdbQuery.Queries[0].Model.Get("type").MustString("")

	switch queryType {
	case "annotationQuery":
		result, err = e.executeAnnotationQuery(ctx, tsdbQuery)
	case "getGCEDefaultProject":
		result, err = e.getGCEDefaultProject(ctx, tsdbQuery)
	case "timeSeriesQuery":
		fallthrough
	default:
		result, err = e.executeTimeSeriesQuery(ctx, tsdbQuery)
	}

	return result, err
}

//nolint: staticcheck // plugins.DataPlugin deprecated
func (e *Executor) getGCEDefaultProject(ctx context.Context, tsdbQuery plugins.DataQuery) (plugins.DataResponse, error) {
	result := plugins.DataResponse{
		//nolint: staticcheck // plugins.DataPlugin deprecated
		Results: make(map[string]plugins.DataQueryResult),
	}
	refID := tsdbQuery.Queries[0].RefID
	//nolint: staticcheck // plugins.DataPlugin deprecated
	queryResult := plugins.DataQueryResult{Meta: simplejson.New(), RefID: refID}

	gceDefaultProject, err := e.getDefaultProject(ctx)
	if err != nil {
		return plugins.DataResponse{}, fmt.Errorf(
			"failed to retrieve default project from GCE metadata server, error: %w", err)
	}

	queryResult.Meta.Set("defaultProject", gceDefaultProject)
	result.Results[refID] = queryResult

	return result, nil
}

//nolint: staticcheck // plugins.DataPlugin deprecated
func (e *Executor) executeTimeSeriesQuery(ctx context.Context, tsdbQuery plugins.DataQuery) (
	plugins.DataResponse, error) {
	result := plugins.DataResponse{
		//nolint: staticcheck // plugins.DataPlugin deprecated
		Results: make(map[string]plugins.DataQueryResult),
	}

	queryExecutors, err := e.buildQueryExecutors(tsdbQuery)
	if err != nil {
		return plugins.DataResponse{}, err
	}

	for _, queryExecutor := range queryExecutors {
		queryRes, resp, executedQueryString, err := queryExecutor.run(ctx, tsdbQuery, e)
		if err != nil {
			return plugins.DataResponse{}, err
		}
		err = queryExecutor.parseResponse(&queryRes, resp, executedQueryString)
		if err != nil {
			queryRes.Error = err
		}

		result.Results[queryExecutor.getRefID()] = queryRes
	}

	return result, nil
}

func (e *Executor) buildQueryExecutors(tsdbQuery plugins.DataQuery) ([]cloudMonitoringQueryExecutor, error) {
	cloudMonitoringQueryExecutors := []cloudMonitoringQueryExecutor{}

	startTime, err := tsdbQuery.TimeRange.ParseFrom()
	if err != nil {
		return nil, err
	}

	endTime, err := tsdbQuery.TimeRange.ParseTo()
	if err != nil {
		return nil, err
	}

	durationSeconds := int(endTime.Sub(startTime).Seconds())

	for i := range tsdbQuery.Queries {
		migrateLegacyQueryModel(&tsdbQuery.Queries[i])
		query := tsdbQuery.Queries[i]
		q := grafanaQuery{}
		model, err := query.Model.MarshalJSON()
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(model, &q); err != nil {
			return nil, fmt.Errorf("could not unmarshal CloudMonitoringQuery json: %w", err)
		}
		q.MetricQuery.PreprocessorType = toPreprocessorType(q.MetricQuery.Preprocessor)
		var target string
		params := url.Values{}
		params.Add("interval.startTime", startTime.UTC().Format(time.RFC3339))
		params.Add("interval.endTime", endTime.UTC().Format(time.RFC3339))

		var queryInterface cloudMonitoringQueryExecutor
		cmtsf := &cloudMonitoringTimeSeriesFilter{
			RefID:    query.RefID,
			GroupBys: []string{},
		}
		switch q.QueryType {
		case metricQueryType:
			if q.MetricQuery.EditorMode == mqlEditorMode {
				queryInterface = &cloudMonitoringTimeSeriesQuery{
					RefID:       query.RefID,
					ProjectName: q.MetricQuery.ProjectName,
					Query:       q.MetricQuery.Query,
					IntervalMS:  query.IntervalMS,
					AliasBy:     q.MetricQuery.AliasBy,
					timeRange:   *tsdbQuery.TimeRange,
				}
			} else {
				cmtsf.AliasBy = q.MetricQuery.AliasBy
				cmtsf.ProjectName = q.MetricQuery.ProjectName
				cmtsf.GroupBys = append(cmtsf.GroupBys, q.MetricQuery.GroupBys...)
				if q.MetricQuery.View == "" {
					q.MetricQuery.View = "FULL"
				}
				params.Add("filter", buildFilterString(q.MetricQuery.MetricType, q.MetricQuery.Filters))
				params.Add("view", q.MetricQuery.View)
				setMetricAggParams(&params, &q.MetricQuery, durationSeconds, query.IntervalMS)
				queryInterface = cmtsf
			}
		case sloQueryType:
			cmtsf.AliasBy = q.SloQuery.AliasBy
			cmtsf.ProjectName = q.SloQuery.ProjectName
			cmtsf.Selector = q.SloQuery.SelectorName
			cmtsf.Service = q.SloQuery.ServiceId
			cmtsf.Slo = q.SloQuery.SloId
			params.Add("filter", buildSLOFilterExpression(q.SloQuery))
			setSloAggParams(&params, &q.SloQuery, durationSeconds, query.IntervalMS)
			queryInterface = cmtsf
		default:
			panic(fmt.Sprintf("Unrecognized query type %q", q.QueryType))
		}

		target = params.Encode()
		cmtsf.Target = target
		cmtsf.Params = params

		if setting.Env == setting.Dev {
			slog.Debug("CloudMonitoring request", "params", params)
		}

		cloudMonitoringQueryExecutors = append(cloudMonitoringQueryExecutors, queryInterface)
	}

	return cloudMonitoringQueryExecutors, nil
}

func migrateLegacyQueryModel(query *plugins.DataSubQuery) {
	mq := query.Model.Get("metricQuery").MustMap()
	if mq == nil {
		migratedModel := simplejson.NewFromAny(map[string]interface{}{
			"queryType":   metricQueryType,
			"metricQuery": query.Model.MustMap(),
		})
		query.Model = migratedModel
	}
}

func reverse(s string) string {
	chars := []rune(s)
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func interpolateFilterWildcards(value string) string {
	matches := strings.Count(value, "*")
	switch {
	case matches == 2 && strings.HasSuffix(value, "*") && strings.HasPrefix(value, "*"):
		value = strings.ReplaceAll(value, "*", "")
		value = fmt.Sprintf(`has_substring("%s")`, value)
	case matches == 1 && strings.HasPrefix(value, "*"):
		value = strings.Replace(value, "*", "", 1)
		value = fmt.Sprintf(`ends_with("%s")`, value)
	case matches == 1 && strings.HasSuffix(value, "*"):
		value = reverse(strings.Replace(reverse(value), "*", "", 1))
		value = fmt.Sprintf(`starts_with("%s")`, value)
	case matches != 0:
		value = string(wildcardRegexRe.ReplaceAllFunc([]byte(value), func(in []byte) []byte {
			return []byte(strings.Replace(string(in), string(in), `\\`+string(in), 1))
		}))
		value = strings.ReplaceAll(value, "*", ".*")
		value = strings.ReplaceAll(value, `"`, `\\"`)
		value = fmt.Sprintf(`monitoring.regex.full_match("^%s$")`, value)
	}

	return value
}

func buildFilterString(metricType string, filterParts []string) string {
	filterString := ""
	for i, part := range filterParts {
		mod := i % 4
		switch {
		case part == "AND":
			filterString += " "
		case mod == 2:
			operator := filterParts[i-1]
			switch {
			case operator == "=~" || operator == "!=~":
				filterString = reverse(strings.Replace(reverse(filterString), "~", "", 1))
				filterString += fmt.Sprintf(`monitoring.regex.full_match("%s")`, part)
			case strings.Contains(part, "*"):
				filterString += interpolateFilterWildcards(part)
			default:
				filterString += fmt.Sprintf(`"%s"`, part)
			}
		default:
			filterString += part
		}
	}

	return strings.Trim(fmt.Sprintf(`metric.type="%s" %s`, metricType, filterString), " ")
}

func buildSLOFilterExpression(q sloQuery) string {
	return fmt.Sprintf(`%s("projects/%s/services/%s/serviceLevelObjectives/%s")`, q.SelectorName, q.ProjectName, q.ServiceId, q.SloId)
}

func setMetricAggParams(params *url.Values, query *metricQuery, durationSeconds int, intervalMs int64) {
	if query.CrossSeriesReducer == "" {
		query.CrossSeriesReducer = crossSeriesReducerDefault
	}

	if query.PerSeriesAligner == "" {
		query.PerSeriesAligner = perSeriesAlignerDefault
	}

	alignmentPeriod := calculateAlignmentPeriod(query.AlignmentPeriod, intervalMs, durationSeconds)

	// In case a preprocessor is defined, the preprocessor becomes the primary aggregation
	// and the aggregation that is specified in the UI becomes the secondary aggregation
	// Rules are specified in this issue: https://github.com/grafana/grafana/issues/30866
	if query.PreprocessorType != PreprocessorTypeNone {
		params.Add("secondaryAggregation.alignmentPeriod", alignmentPeriod)
		params.Add("secondaryAggregation.crossSeriesReducer", query.CrossSeriesReducer)
		params.Add("secondaryAggregation.perSeriesAligner", query.PerSeriesAligner)

		primaryCrossSeriesReducer := crossSeriesReducerDefault
		if len(query.GroupBys) > 0 {
			primaryCrossSeriesReducer = query.CrossSeriesReducer
		}
		params.Add("aggregation.crossSeriesReducer", primaryCrossSeriesReducer)

		aligner := "ALIGN_RATE"
		if query.PreprocessorType == PreprocessorTypeDelta {
			aligner = "ALIGN_DELTA"
		}
		params.Add("aggregation.perSeriesAligner", aligner)

		for _, groupBy := range query.GroupBys {
			params.Add("secondaryAggregation.groupByFields", groupBy)
		}
	} else {
		params.Add("aggregation.crossSeriesReducer", query.CrossSeriesReducer)
		params.Add("aggregation.perSeriesAligner", query.PerSeriesAligner)
	}

	params.Add("aggregation.alignmentPeriod", alignmentPeriod)

	for _, groupBy := range query.GroupBys {
		params.Add("aggregation.groupByFields", groupBy)
	}
}

func setSloAggParams(params *url.Values, query *sloQuery, durationSeconds int, intervalMs int64) {
	params.Add("aggregation.alignmentPeriod", calculateAlignmentPeriod(query.AlignmentPeriod, intervalMs, durationSeconds))
	if query.SelectorName == "select_slo_health" {
		params.Add("aggregation.perSeriesAligner", "ALIGN_MEAN")
	} else {
		params.Add("aggregation.perSeriesAligner", "ALIGN_NEXT_OLDER")
	}
}

func calculateAlignmentPeriod(alignmentPeriod string, intervalMs int64, durationSeconds int) string {
	if alignmentPeriod == "grafana-auto" || alignmentPeriod == "" {
		alignmentPeriodValue := int(math.Max(float64(intervalMs)/1000, 60.0))
		alignmentPeriod = "+" + strconv.Itoa(alignmentPeriodValue) + "s"
	}

	if alignmentPeriod == "cloud-monitoring-auto" || alignmentPeriod == "stackdriver-auto" { // legacy
		alignmentPeriodValue := int(math.Max(float64(durationSeconds), 60.0))
		switch {
		case alignmentPeriodValue < 60*60*23:
			alignmentPeriod = "+60s"
		case alignmentPeriodValue < 60*60*24*6:
			alignmentPeriod = "+300s"
		default:
			alignmentPeriod = "+3600s"
		}
	}

	return alignmentPeriod
}

func toSnakeCase(str string) string {
	return strings.ToLower(matchAllCap.ReplaceAllString(str, "${1}_${2}"))
}

func containsLabel(labels []string, newLabel string) bool {
	for _, val := range labels {
		if val == newLabel {
			return true
		}
	}
	return false
}

func formatLegendKeys(metricType string, defaultMetricName string, labels map[string]string,
	additionalLabels map[string]string, query *cloudMonitoringTimeSeriesFilter) string {
	if query.AliasBy == "" {
		return defaultMetricName
	}

	result := legendKeyFormat.ReplaceAllFunc([]byte(query.AliasBy), func(in []byte) []byte {
		metaPartName := strings.Replace(string(in), "{{", "", 1)
		metaPartName = strings.Replace(metaPartName, "}}", "", 1)
		metaPartName = strings.TrimSpace(metaPartName)

		if metaPartName == "metric.type" {
			return []byte(metricType)
		}

		metricPart := replaceWithMetricPart(metaPartName, metricType)

		if metricPart != nil {
			return metricPart
		}

		if val, exists := labels[metaPartName]; exists {
			return []byte(val)
		}

		if val, exists := additionalLabels[metaPartName]; exists {
			return []byte(val)
		}

		if metaPartName == "project" && query.ProjectName != "" {
			return []byte(query.ProjectName)
		}

		if metaPartName == "service" && query.Service != "" {
			return []byte(query.Service)
		}

		if metaPartName == "slo" && query.Slo != "" {
			return []byte(query.Slo)
		}

		if metaPartName == "selector" && query.Selector != "" {
			return []byte(query.Selector)
		}

		return in
	})

	return string(result)
}

func replaceWithMetricPart(metaPartName string, metricType string) []byte {
	// https://cloud.google.com/monitoring/api/v3/metrics-details#label_names
	shortMatches := metricNameFormat.FindStringSubmatch(metricType)

	if metaPartName == "metric.name" {
		if len(shortMatches) > 2 {
			return []byte(shortMatches[3])
		}
	}

	if metaPartName == "metric.service" {
		if len(shortMatches) > 0 {
			return []byte(shortMatches[1])
		}
	}

	return nil
}

func calcBucketBound(bucketOptions cloudMonitoringBucketOptions, n int) string {
	bucketBound := "0"
	if n == 0 {
		return bucketBound
	}

	switch {
	case bucketOptions.LinearBuckets != nil:
		bucketBound = strconv.FormatInt(bucketOptions.LinearBuckets.Offset+(bucketOptions.LinearBuckets.Width*int64(n-1)), 10)
	case bucketOptions.ExponentialBuckets != nil:
		bucketBound = strconv.FormatInt(int64(bucketOptions.ExponentialBuckets.Scale*math.Pow(bucketOptions.ExponentialBuckets.GrowthFactor, float64(n-1))), 10)
	case bucketOptions.ExplicitBuckets != nil:
		bucketBound = fmt.Sprintf("%g", bucketOptions.ExplicitBuckets.Bounds[n])
	}
	return bucketBound
}

func (e *Executor) createRequest(ctx context.Context, dsInfo *models.DataSource, proxyPass string, body io.Reader) (*http.Request, error) {
	u, err := url.Parse(dsInfo.Url)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "render")

	method := http.MethodGet
	if body != nil {
		method = http.MethodPost
	}
	req, err := http.NewRequest(method, "https://monitoring.googleapis.com/", body)
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// find plugin
	plugin := e.pluginManager.GetDataSource(dsInfo.Type)
	if plugin == nil {
		return nil, errors.New("unable to find datasource plugin CloudMonitoring")
	}

	var cloudMonitoringRoute *plugins.AppPluginRoute
	for _, route := range plugin.Routes {
		if route.Path == "cloudmonitoring" {
			cloudMonitoringRoute = route
			break
		}
	}

	pluginproxy.ApplyRoute(ctx, req, proxyPass, cloudMonitoringRoute, dsInfo, e.cfg)

	return req, nil
}

func (e *Executor) getDefaultProject(ctx context.Context) (string, error) {
	authenticationType := e.dsInfo.JsonData.Get("authenticationType").MustString(jwtAuthentication)
	if authenticationType == gceAuthentication {
		defaultCredentials, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/monitoring.read")
		if err != nil {
			return "", fmt.Errorf("failed to retrieve default project from GCE metadata server: %w", err)
		}
		token, err := defaultCredentials.TokenSource.Token()
		if err != nil {
			return "", fmt.Errorf("failed to retrieve GCP credential token: %w", err)
		}
		if !token.Valid() {
			return "", errors.New("failed to validate GCP credentials")
		}

		return defaultCredentials.ProjectID, nil
	}
	return e.dsInfo.JsonData.Get("defaultProject").MustString(), nil
}

func unmarshalResponse(res *http.Response) (cloudMonitoringResponse, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return cloudMonitoringResponse{}, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Warn("Failed to close response body", "err", err)
		}
	}()

	if res.StatusCode/100 != 2 {
		slog.Error("Request failed", "status", res.Status, "body", string(body))
		return cloudMonitoringResponse{}, fmt.Errorf("query failed: %s", string(body))
	}

	var data cloudMonitoringResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		slog.Error("Failed to unmarshal CloudMonitoring response", "error", err, "status", res.Status, "body", string(body))
		return cloudMonitoringResponse{}, fmt.Errorf("failed to unmarshal query response: %w", err)
	}

	return data, nil
}

func addConfigData(frames data.Frames, dl string, unit string) data.Frames {
	for i := range frames {
		if frames[i].Fields[1].Config == nil {
			frames[i].Fields[1].Config = &data.FieldConfig{}
		}
		deepLink := data.DataLink{
			Title:       "View in Metrics Explorer",
			TargetBlank: true,
			URL:         dl,
		}
		frames[i].Fields[1].Config.Links = append(frames[i].Fields[1].Config.Links, deepLink)
		if len(unit) > 0 {
			if val, ok := cloudMonitoringUnitMappings[unit]; ok {
				frames[i].Fields[1].Config.Unit = val
			}
		}
	}
	return frames
}
