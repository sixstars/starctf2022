package conditions

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/services/validations"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/null"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/services/alerting"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
	"github.com/xorcare/pointer"
)

func newTimeSeriesPointsFromArgs(values ...float64) plugins.DataTimeSeriesPoints {
	points := make(plugins.DataTimeSeriesPoints, 0)

	for i := 0; i < len(values); i += 2 {
		points = append(points, plugins.DataTimePoint{null.FloatFrom(values[i]), null.FloatFrom(values[i+1])})
	}

	return points
}

func TestQueryCondition(t *testing.T) {
	Convey("when evaluating query condition", t, func() {
		queryConditionScenario("Given avg() and > 100", func(ctx *queryConditionTestContext) {
			ctx.reducer = `{"type": "avg"}`
			ctx.evaluator = `{"type": "gt", "params": [100]}`

			Convey("Can read query condition from json model", func() {
				_, err := ctx.exec()
				So(err, ShouldBeNil)

				So(ctx.condition.Query.From, ShouldEqual, "5m")
				So(ctx.condition.Query.To, ShouldEqual, "now")
				So(ctx.condition.Query.DatasourceID, ShouldEqual, 1)

				Convey("Can read query reducer", func() {
					reducer := ctx.condition.Reducer
					So(reducer.Type, ShouldEqual, "avg")
				})

				Convey("Can read evaluator", func() {
					evaluator, ok := ctx.condition.Evaluator.(*thresholdEvaluator)
					So(ok, ShouldBeTrue)
					So(evaluator.Type, ShouldEqual, "gt")
				})
			})

			Convey("should fire when avg is above 100", func() {
				points := newTimeSeriesPointsFromArgs(120, 0)
				ctx.series = plugins.DataTimeSeriesSlice{plugins.DataTimeSeries{Name: "test1", Points: points}}
				cr, err := ctx.exec()

				So(err, ShouldBeNil)
				So(cr.Firing, ShouldBeTrue)
			})

			Convey("should fire when avg is above 100 on dataframe", func() {
				ctx.frame = data.NewFrame("",
					data.NewField("time", nil, []time.Time{time.Now(), time.Now()}),
					data.NewField("val", nil, []int64{120, 150}),
				)
				cr, err := ctx.exec()

				So(err, ShouldBeNil)
				So(cr.Firing, ShouldBeTrue)
			})

			Convey("Should not fire when avg is below 100", func() {
				points := newTimeSeriesPointsFromArgs(90, 0)
				ctx.series = plugins.DataTimeSeriesSlice{plugins.DataTimeSeries{Name: "test1", Points: points}}
				cr, err := ctx.exec()

				So(err, ShouldBeNil)
				So(cr.Firing, ShouldBeFalse)
			})

			Convey("Should not fire when avg is below 100 on dataframe", func() {
				ctx.frame = data.NewFrame("",
					data.NewField("time", nil, []time.Time{time.Now(), time.Now()}),
					data.NewField("val", nil, []int64{12, 47}),
				)
				cr, err := ctx.exec()

				So(err, ShouldBeNil)
				So(cr.Firing, ShouldBeFalse)
			})

			Convey("Should fire if only first series matches", func() {
				ctx.series = plugins.DataTimeSeriesSlice{
					plugins.DataTimeSeries{Name: "test1", Points: newTimeSeriesPointsFromArgs(120, 0)},
					plugins.DataTimeSeries{Name: "test2", Points: newTimeSeriesPointsFromArgs(0, 0)},
				}
				cr, err := ctx.exec()

				So(err, ShouldBeNil)
				So(cr.Firing, ShouldBeTrue)
			})

			Convey("No series", func() {
				Convey("Should set NoDataFound when condition is gt", func() {
					ctx.series = plugins.DataTimeSeriesSlice{}
					cr, err := ctx.exec()

					So(err, ShouldBeNil)
					So(cr.Firing, ShouldBeFalse)
					So(cr.NoDataFound, ShouldBeTrue)
				})

				Convey("Should be firing when condition is no_value", func() {
					ctx.evaluator = `{"type": "no_value", "params": []}`
					ctx.series = plugins.DataTimeSeriesSlice{}
					cr, err := ctx.exec()

					So(err, ShouldBeNil)
					So(cr.Firing, ShouldBeTrue)
				})
			})

			Convey("Empty series", func() {
				Convey("Should set Firing if eval match", func() {
					ctx.evaluator = `{"type": "no_value", "params": []}`
					ctx.series = plugins.DataTimeSeriesSlice{
						plugins.DataTimeSeries{Name: "test1", Points: newTimeSeriesPointsFromArgs()},
					}
					cr, err := ctx.exec()

					So(err, ShouldBeNil)
					So(cr.Firing, ShouldBeTrue)
				})

				Convey("Should set NoDataFound both series are empty", func() {
					ctx.series = plugins.DataTimeSeriesSlice{
						plugins.DataTimeSeries{Name: "test1", Points: newTimeSeriesPointsFromArgs()},
						plugins.DataTimeSeries{Name: "test2", Points: newTimeSeriesPointsFromArgs()},
					}
					cr, err := ctx.exec()

					So(err, ShouldBeNil)
					So(cr.NoDataFound, ShouldBeTrue)
				})

				Convey("Should set NoDataFound both series contains null", func() {
					ctx.series = plugins.DataTimeSeriesSlice{
						plugins.DataTimeSeries{Name: "test1", Points: plugins.DataTimeSeriesPoints{plugins.DataTimePoint{null.FloatFromPtr(nil), null.FloatFrom(0)}}},
						plugins.DataTimeSeries{Name: "test2", Points: plugins.DataTimeSeriesPoints{plugins.DataTimePoint{null.FloatFromPtr(nil), null.FloatFrom(0)}}},
					}
					cr, err := ctx.exec()

					So(err, ShouldBeNil)
					So(cr.NoDataFound, ShouldBeTrue)
				})

				Convey("Should not set NoDataFound if one series is empty", func() {
					ctx.series = plugins.DataTimeSeriesSlice{
						plugins.DataTimeSeries{Name: "test1", Points: newTimeSeriesPointsFromArgs()},
						plugins.DataTimeSeries{Name: "test2", Points: newTimeSeriesPointsFromArgs(120, 0)},
					}
					cr, err := ctx.exec()

					So(err, ShouldBeNil)
					So(cr.NoDataFound, ShouldBeFalse)
				})
			})
		})
	})
}

type queryConditionTestContext struct {
	reducer   string
	evaluator string
	series    plugins.DataTimeSeriesSlice
	frame     *data.Frame
	result    *alerting.EvalContext
	condition *QueryCondition
}

type queryConditionScenarioFunc func(c *queryConditionTestContext)

//nolint: staticcheck // plugins.DataPlugin deprecated
func (ctx *queryConditionTestContext) exec() (*alerting.ConditionResult, error) {
	jsonModel, err := simplejson.NewJson([]byte(`{
            "type": "query",
            "query":  {
              "params": ["A", "5m", "now"],
              "datasourceId": 1,
              "model": {"target": "aliasByNode(statsd.fakesite.counters.session_start.mobile.count, 4)"}
            },
            "reducer":` + ctx.reducer + `,
            "evaluator":` + ctx.evaluator + `
          }`))
	So(err, ShouldBeNil)

	condition, err := newQueryCondition(jsonModel, 0)
	So(err, ShouldBeNil)

	ctx.condition = condition

	qr := plugins.DataQueryResult{
		Series: ctx.series,
	}

	if ctx.frame != nil {
		qr = plugins.DataQueryResult{
			Dataframes: plugins.NewDecodedDataFrames(data.Frames{ctx.frame}),
		}
	}
	reqHandler := fakeReqHandler{
		response: plugins.DataResponse{
			Results: map[string]plugins.DataQueryResult{
				"A": qr,
			},
		},
	}

	return condition.Eval(ctx.result, reqHandler)
}

type fakeReqHandler struct {
	//nolint: staticcheck // plugins.DataPlugin deprecated
	response plugins.DataResponse
}

//nolint: staticcheck // plugins.DataPlugin deprecated
func (rh fakeReqHandler) HandleRequest(context.Context, *models.DataSource, plugins.DataQuery) (
	plugins.DataResponse, error) {
	return rh.response, nil
}

func queryConditionScenario(desc string, fn queryConditionScenarioFunc) {
	Convey(desc, func() {
		bus.AddHandler("test", func(query *models.GetDataSourceQuery) error {
			query.Result = &models.DataSource{Id: 1, Type: "graphite"}
			return nil
		})

		ctx := &queryConditionTestContext{}
		ctx.result = &alerting.EvalContext{
			Rule:             &alerting.Rule{},
			RequestValidator: &validations.OSSPluginRequestValidator{},
		}

		fn(ctx)
	})
}

func TestFrameToSeriesSlice(t *testing.T) {
	tests := []struct {
		name        string
		frame       *data.Frame
		seriesSlice plugins.DataTimeSeriesSlice
		Err         require.ErrorAssertionFunc
	}{
		{
			name: "a wide series",
			frame: data.NewFrame("",
				data.NewField("Time", nil, []time.Time{
					time.Date(2020, 1, 2, 3, 4, 0, 0, time.UTC),
					time.Date(2020, 1, 2, 3, 4, 30, 0, time.UTC),
				}),
				data.NewField(`Values Int64s`, data.Labels{"Animal Factor": "cat"}, []*int64{
					nil,
					pointer.Int64(3),
				}),
				data.NewField(`Values Floats`, data.Labels{"Animal Factor": "sloth"}, []float64{
					2.0,
					4.0,
				})),

			seriesSlice: plugins.DataTimeSeriesSlice{
				plugins.DataTimeSeries{
					Name: "Values Int64s {Animal Factor=cat}",
					Tags: map[string]string{"Animal Factor": "cat"},
					Points: plugins.DataTimeSeriesPoints{
						plugins.DataTimePoint{null.FloatFrom(math.NaN()), null.FloatFrom(1577934240000)},
						plugins.DataTimePoint{null.FloatFrom(3), null.FloatFrom(1577934270000)},
					},
				},
				plugins.DataTimeSeries{
					Name: "Values Floats {Animal Factor=sloth}",
					Tags: map[string]string{"Animal Factor": "sloth"},
					Points: plugins.DataTimeSeriesPoints{
						plugins.DataTimePoint{null.FloatFrom(2), null.FloatFrom(1577934240000)},
						plugins.DataTimePoint{null.FloatFrom(4), null.FloatFrom(1577934270000)},
					},
				},
			},
			Err: require.NoError,
		},
		{
			name: "empty wide series",
			frame: data.NewFrame("",
				data.NewField("Time", nil, []time.Time{}),
				data.NewField(`Values Int64s`, data.Labels{"Animal Factor": "cat"}, []*int64{}),
				data.NewField(`Values Floats`, data.Labels{"Animal Factor": "sloth"}, []float64{})),

			seriesSlice: plugins.DataTimeSeriesSlice{
				plugins.DataTimeSeries{
					Name:   "Values Int64s {Animal Factor=cat}",
					Tags:   map[string]string{"Animal Factor": "cat"},
					Points: plugins.DataTimeSeriesPoints{},
				},
				plugins.DataTimeSeries{
					Name:   "Values Floats {Animal Factor=sloth}",
					Tags:   map[string]string{"Animal Factor": "sloth"},
					Points: plugins.DataTimeSeriesPoints{},
				},
			},
			Err: require.NoError,
		},
		{
			name: "empty labels",
			frame: data.NewFrame("",
				data.NewField("Time", data.Labels{}, []time.Time{}),
				data.NewField(`Values`, data.Labels{}, []float64{})),

			seriesSlice: plugins.DataTimeSeriesSlice{
				plugins.DataTimeSeries{
					Name:   "Values",
					Points: plugins.DataTimeSeriesPoints{},
				},
			},
			Err: require.NoError,
		},
		{
			name: "display name from data source",
			frame: data.NewFrame("",
				data.NewField("Time", data.Labels{}, []time.Time{}),
				data.NewField(`Values`, data.Labels{"Rating": "10"}, []*int64{}).SetConfig(&data.FieldConfig{
					DisplayNameFromDS: "sloth",
				})),

			seriesSlice: plugins.DataTimeSeriesSlice{
				plugins.DataTimeSeries{
					Name:   "sloth",
					Points: plugins.DataTimeSeriesPoints{},
					Tags:   map[string]string{"Rating": "10"},
				},
			},
			Err: require.NoError,
		},
		{
			name: "prefer display name over data source display name",
			frame: data.NewFrame("",
				data.NewField("Time", data.Labels{}, []time.Time{}),
				data.NewField(`Values`, data.Labels{}, []*int64{}).SetConfig(&data.FieldConfig{
					DisplayName:       "sloth #1",
					DisplayNameFromDS: "sloth #2",
				})),

			seriesSlice: plugins.DataTimeSeriesSlice{
				plugins.DataTimeSeries{
					Name:   "sloth #1",
					Points: plugins.DataTimeSeriesPoints{},
				},
			},
			Err: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seriesSlice, err := FrameToSeriesSlice(tt.frame)
			tt.Err(t, err)
			if diff := cmp.Diff(tt.seriesSlice, seriesSlice, cmpopts.EquateNaNs()); diff != "" {
				t.Errorf("Result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
