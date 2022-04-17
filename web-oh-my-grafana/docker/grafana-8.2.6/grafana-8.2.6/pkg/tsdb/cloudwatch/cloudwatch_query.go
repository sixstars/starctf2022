package cloudwatch

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type cloudWatchQuery struct {
	RefId          string
	Region         string
	Id             string
	Namespace      string
	MetricName     string
	Statistic      string
	Expression     string
	ReturnData     bool
	Dimensions     map[string][]string
	Period         int
	Alias          string
	MatchExact     bool
	UsedExpression string
}

func (q *cloudWatchQuery) isMathExpression() bool {
	return q.Expression != "" && !q.isUserDefinedSearchExpression()
}

func (q *cloudWatchQuery) isSearchExpression() bool {
	return q.isUserDefinedSearchExpression() || q.isInferredSearchExpression()
}

func (q *cloudWatchQuery) isUserDefinedSearchExpression() bool {
	return strings.Contains(q.Expression, "SEARCH(")
}

func (q *cloudWatchQuery) isInferredSearchExpression() bool {
	if len(q.Dimensions) == 0 {
		return !q.MatchExact
	}
	if !q.MatchExact {
		return true
	}

	for _, values := range q.Dimensions {
		if len(values) > 1 {
			return true
		}
		for _, v := range values {
			if v == "*" {
				return true
			}
		}
	}
	return false
}

func (q *cloudWatchQuery) isMultiValuedDimensionExpression() bool {
	for _, values := range q.Dimensions {
		for _, v := range values {
			if v == "*" {
				return false
			}
		}

		if len(values) > 1 {
			return true
		}
	}

	return false
}

func (q *cloudWatchQuery) buildDeepLink(startTime time.Time, endTime time.Time) (string, error) {
	if q.isMathExpression() {
		return "", nil
	}

	link := &cloudWatchLink{
		Title:   q.RefId,
		View:    "timeSeries",
		Stacked: false,
		Region:  q.Region,
		Start:   startTime.UTC().Format(time.RFC3339),
		End:     endTime.UTC().Format(time.RFC3339),
	}

	if q.isSearchExpression() {
		link.Metrics = []interface{}{&metricExpression{Expression: q.UsedExpression}}
	} else {
		metricStat := []interface{}{q.Namespace, q.MetricName}
		for dimensionKey, dimensionValues := range q.Dimensions {
			metricStat = append(metricStat, dimensionKey, dimensionValues[0])
		}
		metricStat = append(metricStat, &metricStatMeta{
			Stat:   q.Statistic,
			Period: q.Period,
		})
		link.Metrics = []interface{}{metricStat}
	}

	linkProps, err := json.Marshal(link)
	if err != nil {
		return "", fmt.Errorf("could not marshal link: %w", err)
	}

	url, err := url.Parse(fmt.Sprintf(`https://%s.console.aws.amazon.com/cloudwatch/deeplink.js`, q.Region))
	if err != nil {
		return "", fmt.Errorf("unable to parse CloudWatch console deep link")
	}

	fragment := url.Query()
	fragment.Set("graph", string(linkProps))

	query := url.Query()
	query.Set("region", q.Region)
	url.RawQuery = query.Encode()

	return fmt.Sprintf(`%s#metricsV2:%s`, url.String(), fragment.Encode()), nil
}
