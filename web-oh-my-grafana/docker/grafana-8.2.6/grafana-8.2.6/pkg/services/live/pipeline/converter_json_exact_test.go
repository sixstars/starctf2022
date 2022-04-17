package pipeline

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/experimental"
	"github.com/stretchr/testify/require"
)

func checkExactConversion(tb testing.TB, file string, fields []Field) *backend.DataResponse {
	tb.Helper()
	content := loadTestJson(tb, file)

	converter := NewExactJsonConverter(ExactJsonConverterConfig{
		Fields: fields,
	})
	converter.nowTimeFunc = func() time.Time {
		return time.Date(2021, 01, 01, 12, 12, 12, 0, time.UTC)
	}
	channelFrames, err := converter.Convert(context.Background(), Vars{}, content)
	require.NoError(tb, err)

	dr := &backend.DataResponse{}
	for _, cf := range channelFrames {
		require.Empty(tb, cf.Channel)
		dr.Frames = append(dr.Frames, cf.Frame)
	}

	err = experimental.CheckGoldenDataResponse(filepath.Join("testdata", file+".golden.txt"), dr, *update)
	require.NoError(tb, err)
	return dr
}

func TestExactJsonConverter_Convert(t *testing.T) {
	checkExactConversion(t, "json_exact", []Field{
		{
			Name:  "time",
			Value: "#{now}",
			Type:  data.FieldTypeTime,
		},
		{
			Name:  "ax",
			Value: "$.ax",
			Type:  data.FieldTypeNullableFloat64,
		},
		{
			Name:  "key1",
			Value: "{x.map_with_floats.key1}",
			Type:  data.FieldTypeNullableFloat64,
			Labels: []Label{
				{
					Name:  "label1",
					Value: "{x.map_with_floats.key2.toString()}",
				},
				{
					Name:  "label2",
					Value: "$.map_with_floats.key2",
				},
			},
		},
	})
}
