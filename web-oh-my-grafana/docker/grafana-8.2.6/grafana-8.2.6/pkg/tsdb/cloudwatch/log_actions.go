package cloudwatch

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"golang.org/x/sync/errgroup"
)

func (e *cloudWatchExecutor) executeLogActions(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	resultChan := make(chan backend.Responses, len(req.Queries))
	eg, ectx := errgroup.WithContext(ctx)

	for _, query := range req.Queries {
		model, err := simplejson.NewJson(query.JSON)
		if err != nil {
			return nil, err
		}

		query := query
		eg.Go(func() error {
			dataframe, err := e.executeLogAction(ectx, model, query, req.PluginContext)
			if err != nil {
				return err
			}

			groupedFrames, err := groupResponseFrame(dataframe, model.Get("statsGroups").MustStringArray())
			if err != nil {
				return err
			}
			resultChan <- backend.Responses{
				query.RefID: backend.DataResponse{Frames: groupedFrames},
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	close(resultChan)

	for result := range resultChan {
		for refID, response := range result {
			respD := resp.Responses[refID]
			respD.Frames = response.Frames
			resp.Responses[refID] = respD
		}
	}

	return resp, nil
}

func (e *cloudWatchExecutor) executeLogAction(ctx context.Context, model *simplejson.Json, query backend.DataQuery, pluginCtx backend.PluginContext) (*data.Frame, error) {
	subType := model.Get("subtype").MustString()

	dsInfo, err := e.getDSInfo(pluginCtx)
	if err != nil {
		return nil, err
	}

	defaultRegion := dsInfo.region

	region := model.Get("region").MustString(defaultRegion)
	logsClient, err := e.getCWLogsClient(region, pluginCtx)
	if err != nil {
		return nil, err
	}

	var data *data.Frame = nil

	switch subType {
	case "DescribeLogGroups":
		data, err = e.handleDescribeLogGroups(ctx, logsClient, model)
	case "GetLogGroupFields":
		data, err = e.handleGetLogGroupFields(ctx, logsClient, model, query.RefID)
	case "StartQuery":
		data, err = e.handleStartQuery(ctx, logsClient, model, query.TimeRange, query.RefID)
	case "StopQuery":
		data, err = e.handleStopQuery(ctx, logsClient, model)
	case "GetQueryResults":
		data, err = e.handleGetQueryResults(ctx, logsClient, model, query.RefID)
	case "GetLogEvents":
		data, err = e.handleGetLogEvents(ctx, logsClient, model)
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (e *cloudWatchExecutor) handleGetLogEvents(ctx context.Context, logsClient cloudwatchlogsiface.CloudWatchLogsAPI,
	parameters *simplejson.Json) (*data.Frame, error) {
	queryRequest := &cloudwatchlogs.GetLogEventsInput{
		Limit:         aws.Int64(parameters.Get("limit").MustInt64(10)),
		StartFromHead: aws.Bool(parameters.Get("startFromHead").MustBool(false)),
	}

	logGroupName, err := parameters.Get("logGroupName").String()
	if err != nil {
		return nil, fmt.Errorf("Error: Parameter 'logGroupName' is required")
	}
	queryRequest.SetLogGroupName(logGroupName)

	logStreamName, err := parameters.Get("logStreamName").String()
	if err != nil {
		return nil, fmt.Errorf("Error: Parameter 'logStream' is required")
	}
	queryRequest.SetLogStreamName(logStreamName)

	if startTime, err := parameters.Get("startTime").Int64(); err == nil {
		queryRequest.SetStartTime(startTime)
	}

	if endTime, err := parameters.Get("endTime").Int64(); err == nil {
		queryRequest.SetEndTime(endTime)
	}

	logEvents, err := logsClient.GetLogEventsWithContext(ctx, queryRequest)
	if err != nil {
		return nil, err
	}

	messages := make([]*string, 0)
	timestamps := make([]*int64, 0)

	sort.Slice(logEvents.Events, func(i, j int) bool {
		return *(logEvents.Events[i].Timestamp) > *(logEvents.Events[j].Timestamp)
	})

	for _, event := range logEvents.Events {
		messages = append(messages, event.Message)
		timestamps = append(timestamps, event.Timestamp)
	}

	timestampField := data.NewField("ts", nil, timestamps)
	timestampField.SetConfig(&data.FieldConfig{DisplayName: "Time"})

	messageField := data.NewField("line", nil, messages)

	return data.NewFrame("logEvents", timestampField, messageField), nil
}

func (e *cloudWatchExecutor) handleDescribeLogGroups(ctx context.Context,
	logsClient cloudwatchlogsiface.CloudWatchLogsAPI, parameters *simplejson.Json) (*data.Frame, error) {
	logGroupNamePrefix := parameters.Get("logGroupNamePrefix").MustString("")

	var response *cloudwatchlogs.DescribeLogGroupsOutput = nil
	var err error
	if len(logGroupNamePrefix) == 0 {
		response, err = logsClient.DescribeLogGroupsWithContext(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
			Limit: aws.Int64(parameters.Get("limit").MustInt64(50)),
		})
	} else {
		response, err = logsClient.DescribeLogGroupsWithContext(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
			Limit:              aws.Int64(parameters.Get("limit").MustInt64(50)),
			LogGroupNamePrefix: aws.String(logGroupNamePrefix),
		})
	}
	if err != nil || response == nil {
		return nil, err
	}

	logGroupNames := make([]*string, 0)
	for _, logGroup := range response.LogGroups {
		logGroupNames = append(logGroupNames, logGroup.LogGroupName)
	}

	groupNamesField := data.NewField("logGroupName", nil, logGroupNames)
	frame := data.NewFrame("logGroups", groupNamesField)

	return frame, nil
}

func (e *cloudWatchExecutor) executeStartQuery(ctx context.Context, logsClient cloudwatchlogsiface.CloudWatchLogsAPI,
	parameters *simplejson.Json, timeRange backend.TimeRange) (*cloudwatchlogs.StartQueryOutput, error) {
	startTime := timeRange.From
	endTime := timeRange.To

	if !startTime.Before(endTime) {
		return nil, fmt.Errorf("invalid time range: start time must be before end time")
	}

	// The fields @log and @logStream are always included in the results of a user's query
	// so that a row's context can be retrieved later if necessary.
	// The usage of ltrim around the @log/@logStream fields is a necessary workaround, as without it,
	// CloudWatch wouldn't consider a query using a non-alised @log/@logStream valid.
	modifiedQueryString := "fields @timestamp,ltrim(@log) as " + logIdentifierInternal + ",ltrim(@logStream) as " +
		logStreamIdentifierInternal + "|" + parameters.Get("queryString").MustString("")

	startQueryInput := &cloudwatchlogs.StartQueryInput{
		StartTime:     aws.Int64(startTime.Unix()),
		EndTime:       aws.Int64(endTime.Unix()),
		LogGroupNames: aws.StringSlice(parameters.Get("logGroupNames").MustStringArray()),
		QueryString:   aws.String(modifiedQueryString),
	}

	if resultsLimit, err := parameters.Get("limit").Int64(); err == nil {
		startQueryInput.Limit = aws.Int64(resultsLimit)
	}

	return logsClient.StartQueryWithContext(ctx, startQueryInput)
}

func (e *cloudWatchExecutor) handleStartQuery(ctx context.Context, logsClient cloudwatchlogsiface.CloudWatchLogsAPI,
	model *simplejson.Json, timeRange backend.TimeRange, refID string) (*data.Frame, error) {
	startQueryResponse, err := e.executeStartQuery(ctx, logsClient, model, timeRange)
	if err != nil {
		return nil, err
	}

	dataFrame := data.NewFrame(refID, data.NewField("queryId", nil, []string{*startQueryResponse.QueryId}))
	dataFrame.RefID = refID

	clientRegion := model.Get("region").MustString("default")

	dataFrame.Meta = &data.FrameMeta{
		Custom: map[string]interface{}{
			"Region": clientRegion,
		},
	}

	return dataFrame, nil
}

func (e *cloudWatchExecutor) executeStopQuery(ctx context.Context, logsClient cloudwatchlogsiface.CloudWatchLogsAPI,
	parameters *simplejson.Json) (*cloudwatchlogs.StopQueryOutput, error) {
	queryInput := &cloudwatchlogs.StopQueryInput{
		QueryId: aws.String(parameters.Get("queryId").MustString()),
	}

	response, err := logsClient.StopQueryWithContext(ctx, queryInput)
	if err != nil {
		// If the query has already stopped by the time CloudWatch receives the stop query request,
		// an "InvalidParameterException" error is returned. For our purposes though the query has been
		// stopped, so we ignore the error.
		var awsErr awserr.Error
		if errors.As(err, &awsErr) && awsErr.Code() == "InvalidParameterException" {
			response = &cloudwatchlogs.StopQueryOutput{Success: aws.Bool(false)}
			err = nil
		}
	}

	return response, err
}

func (e *cloudWatchExecutor) handleStopQuery(ctx context.Context, logsClient cloudwatchlogsiface.CloudWatchLogsAPI,
	parameters *simplejson.Json) (*data.Frame, error) {
	response, err := e.executeStopQuery(ctx, logsClient, parameters)
	if err != nil {
		return nil, err
	}

	dataFrame := data.NewFrame("StopQueryResponse", data.NewField("success", nil, []bool{*response.Success}))
	return dataFrame, nil
}

func (e *cloudWatchExecutor) executeGetQueryResults(ctx context.Context, logsClient cloudwatchlogsiface.CloudWatchLogsAPI,
	parameters *simplejson.Json) (*cloudwatchlogs.GetQueryResultsOutput, error) {
	queryInput := &cloudwatchlogs.GetQueryResultsInput{
		QueryId: aws.String(parameters.Get("queryId").MustString()),
	}

	return logsClient.GetQueryResultsWithContext(ctx, queryInput)
}

func (e *cloudWatchExecutor) handleGetQueryResults(ctx context.Context, logsClient cloudwatchlogsiface.CloudWatchLogsAPI,
	parameters *simplejson.Json, refID string) (*data.Frame, error) {
	getQueryResultsOutput, err := e.executeGetQueryResults(ctx, logsClient, parameters)
	if err != nil {
		return nil, err
	}

	dataFrame, err := logsResultsToDataframes(getQueryResultsOutput)
	if err != nil {
		return nil, err
	}

	dataFrame.Name = refID
	dataFrame.RefID = refID

	return dataFrame, nil
}

func (e *cloudWatchExecutor) handleGetLogGroupFields(ctx context.Context, logsClient cloudwatchlogsiface.CloudWatchLogsAPI,
	parameters *simplejson.Json, refID string) (*data.Frame, error) {
	queryInput := &cloudwatchlogs.GetLogGroupFieldsInput{
		LogGroupName: aws.String(parameters.Get("logGroupName").MustString()),
		Time:         aws.Int64(parameters.Get("time").MustInt64()),
	}

	getLogGroupFieldsOutput, err := logsClient.GetLogGroupFieldsWithContext(ctx, queryInput)
	if err != nil {
		return nil, err
	}

	fieldNames := make([]*string, 0)
	fieldPercentages := make([]*int64, 0)

	for _, logGroupField := range getLogGroupFieldsOutput.LogGroupFields {
		fieldNames = append(fieldNames, logGroupField.Name)
		fieldPercentages = append(fieldPercentages, logGroupField.Percent)
	}

	dataFrame := data.NewFrame(
		refID,
		data.NewField("name", nil, fieldNames),
		data.NewField("percent", nil, fieldPercentages),
	)

	dataFrame.RefID = refID

	return dataFrame, nil
}
