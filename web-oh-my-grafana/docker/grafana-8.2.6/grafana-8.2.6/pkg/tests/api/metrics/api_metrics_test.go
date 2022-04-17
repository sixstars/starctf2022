package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/tests/testinfra"
	"github.com/grafana/grafana/pkg/tsdb/cloudwatch"

	cwapi "github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryCloudWatchMetrics(t *testing.T) {
	grafDir, cfgPath := testinfra.CreateGrafDir(t)
	addr, sqlStore := testinfra.StartGrafana(t, grafDir, cfgPath)
	setUpDatabase(t, sqlStore)

	origNewCWClient := cloudwatch.NewCWClient
	t.Cleanup(func() {
		cloudwatch.NewCWClient = origNewCWClient
	})
	var client cloudwatch.FakeCWClient
	cloudwatch.NewCWClient = func(sess *session.Session) cloudwatchiface.CloudWatchAPI {
		return client
	}

	t.Run("Custom metrics", func(t *testing.T) {
		client = cloudwatch.FakeCWClient{
			Metrics: []*cwapi.Metric{
				{
					MetricName: aws.String("Test_MetricName"),
					Dimensions: []*cwapi.Dimension{
						{
							Name: aws.String("Test_DimensionName"),
						},
					},
				},
			},
		}

		req := dtos.MetricRequest{
			Queries: []*simplejson.Json{
				simplejson.NewFromAny(map[string]interface{}{
					"type":         "metricFindQuery",
					"subtype":      "metrics",
					"region":       "us-east-1",
					"namespace":    "custom",
					"datasourceId": 1,
				}),
			},
		}
		result := makeCWRequest(t, req, addr)

		dataFrames := data.Frames{
			&data.Frame{
				RefID: "A",
				Fields: []*data.Field{
					data.NewField("text", nil, []string{"Test_MetricName"}),
					data.NewField("value", nil, []string{"Test_MetricName"}),
				},
				Meta: &data.FrameMeta{
					Custom: map[string]interface{}{
						"rowCount": float64(1),
					},
				},
			},
		}

		expect := backend.NewQueryDataResponse()
		expect.Responses["A"] = backend.DataResponse{
			Frames: dataFrames,
		}
		assert.Equal(t, *expect, result)
	})
}

func TestQueryCloudWatchLogs(t *testing.T) {
	grafDir, cfgPath := testinfra.CreateGrafDir(t)
	addr, store := testinfra.StartGrafana(t, grafDir, cfgPath)
	setUpDatabase(t, store)

	origNewCWLogsClient := cloudwatch.NewCWLogsClient
	t.Cleanup(func() {
		cloudwatch.NewCWLogsClient = origNewCWLogsClient
	})

	var client cloudwatch.FakeCWLogsClient
	cloudwatch.NewCWLogsClient = func(sess *session.Session) cloudwatchlogsiface.CloudWatchLogsAPI {
		return client
	}

	t.Run("Describe log groups", func(t *testing.T) {
		client = cloudwatch.FakeCWLogsClient{}

		req := dtos.MetricRequest{
			Queries: []*simplejson.Json{
				simplejson.NewFromAny(map[string]interface{}{
					"type":         "logAction",
					"subtype":      "DescribeLogGroups",
					"region":       "us-east-1",
					"datasourceId": 1,
				}),
			},
		}
		tr := makeCWRequest(t, req, addr)

		dataFrames := data.Frames{
			&data.Frame{
				Name:  "logGroups",
				RefID: "A",
				Fields: []*data.Field{
					data.NewField("logGroupName", nil, []*string{}),
				},
			},
		}

		expect := backend.NewQueryDataResponse()
		expect.Responses["A"] = backend.DataResponse{
			Frames: dataFrames,
		}
		assert.Equal(t, *expect, tr)
	})
}

func makeCWRequest(t *testing.T, req dtos.MetricRequest, addr string) backend.QueryDataResponse {
	t.Helper()

	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	err := enc.Encode(&req)
	require.NoError(t, err)
	u := fmt.Sprintf("http://%s/api/ds/query", addr)
	t.Logf("Making POST request to %s", u)
	// nolint:gosec
	resp, err := http.Post(u, "application/json", &buf)
	require.NoError(t, err)
	require.NotNil(t, resp)
	t.Cleanup(func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	})

	buf = bytes.Buffer{}
	_, err = io.Copy(&buf, resp.Body)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var tr backend.QueryDataResponse
	err = json.Unmarshal(buf.Bytes(), &tr)
	require.NoError(t, err)

	return tr
}

func setUpDatabase(t *testing.T, store *sqlstore.SQLStore) {
	t.Helper()

	err := store.WithDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		_, err := sess.Insert(&models.DataSource{
			Id: 1,
			// This will be the ID of the main org
			OrgId:   2,
			Name:    "Test",
			Type:    "cloudwatch",
			Created: time.Now(),
			Updated: time.Now(),
		})
		return err
	})
	require.NoError(t, err)

	// Make sure changes are synced with other goroutines
	err = store.Sync()
	require.NoError(t, err)
}
