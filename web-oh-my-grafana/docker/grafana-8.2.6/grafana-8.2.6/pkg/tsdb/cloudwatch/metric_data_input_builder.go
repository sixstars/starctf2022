package cloudwatch

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func (e *cloudWatchExecutor) buildMetricDataInput(startTime time.Time, endTime time.Time,
	queries []*cloudWatchQuery) (*cloudwatch.GetMetricDataInput, error) {
	metricDataInput := &cloudwatch.GetMetricDataInput{
		StartTime: aws.Time(startTime),
		EndTime:   aws.Time(endTime),
		ScanBy:    aws.String("TimestampAscending"),
	}
	for _, query := range queries {
		metricDataQuery, err := e.buildMetricDataQuery(query)
		if err != nil {
			return nil, &queryError{err, query.RefId}
		}
		metricDataInput.MetricDataQueries = append(metricDataInput.MetricDataQueries, metricDataQuery)
	}

	return metricDataInput, nil
}
