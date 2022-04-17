package conditions

import (
	"context"
	"testing"

	"github.com/grafana/grafana/pkg/services/validations"
	"github.com/grafana/grafana/pkg/tsdb/interval"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/services/alerting"
	. "github.com/smartystreets/goconvey/convey"
)

func TestQueryInterval(t *testing.T) {
	Convey("When evaluating query condition, regarding the interval value", t, func() {
		Convey("Can handle interval-calculation with no panel-min-interval and no datasource-min-interval", func() {
			// no panel-min-interval in the queryModel
			queryModel := `{"target": "aliasByNode(statsd.fakesite.counters.session_start.mobile.count, 4)"}`

			// no datasource-min-interval
			var dataSourceJson *simplejson.Json = nil

			timeRange := "5m"

			verifier := func(query plugins.DataSubQuery) {
				// 5minutes timerange = 300000milliseconds; default-resolution is 1500pixels,
				// so we should have 300000/1500 = 200milliseconds here
				So(query.IntervalMS, ShouldEqual, 200)
				So(query.MaxDataPoints, ShouldEqual, interval.DefaultRes)
			}

			applyScenario(timeRange, dataSourceJson, queryModel, verifier)
		})
		Convey("Can handle interval-calculation with panel-min-interval and no datasource-min-interval", func() {
			// panel-min-interval in the queryModel
			queryModel := `{"interval":"123s", "target": "aliasByNode(statsd.fakesite.counters.session_start.mobile.count, 4)"}`

			// no datasource-min-interval
			var dataSourceJson *simplejson.Json = nil

			timeRange := "5m"

			verifier := func(query plugins.DataSubQuery) {
				So(query.IntervalMS, ShouldEqual, 123000)
				So(query.MaxDataPoints, ShouldEqual, interval.DefaultRes)
			}

			applyScenario(timeRange, dataSourceJson, queryModel, verifier)
		})
		Convey("Can handle interval-calculation with no panel-min-interval and datasource-min-interval", func() {
			// no panel-min-interval in the queryModel
			queryModel := `{"target": "aliasByNode(statsd.fakesite.counters.session_start.mobile.count, 4)"}`

			// min-interval in datasource-json
			dataSourceJson, err := simplejson.NewJson([]byte(`{
			"timeInterval": "71s"
		}`))
			So(err, ShouldBeNil)

			timeRange := "5m"

			verifier := func(query plugins.DataSubQuery) {
				So(query.IntervalMS, ShouldEqual, 71000)
				So(query.MaxDataPoints, ShouldEqual, interval.DefaultRes)
			}

			applyScenario(timeRange, dataSourceJson, queryModel, verifier)
		})
		Convey("Can handle interval-calculation with both panel-min-interval and datasource-min-interval", func() {
			// panel-min-interval in the queryModel
			queryModel := `{"interval":"19s", "target": "aliasByNode(statsd.fakesite.counters.session_start.mobile.count, 4)"}`

			// min-interval in datasource-json
			dataSourceJson, err := simplejson.NewJson([]byte(`{
			"timeInterval": "71s"
		}`))
			So(err, ShouldBeNil)

			timeRange := "5m"

			verifier := func(query plugins.DataSubQuery) {
				// when both panel-min-interval and datasource-min-interval exists,
				// panel-min-interval is used
				So(query.IntervalMS, ShouldEqual, 19000)
				So(query.MaxDataPoints, ShouldEqual, interval.DefaultRes)
			}

			applyScenario(timeRange, dataSourceJson, queryModel, verifier)
		})

		Convey("Can handle no min-interval, and very small time-ranges, where the default-min-interval=1ms applies", func() {
			// no panel-min-interval in the queryModel
			queryModel := `{"target": "aliasByNode(statsd.fakesite.counters.session_start.mobile.count, 4)"}`

			// no datasource-min-interval
			var dataSourceJson *simplejson.Json = nil

			timeRange := "1s"

			verifier := func(query plugins.DataSubQuery) {
				// no min-interval exists, the default-min-interval will be used,
				// and for such a short time-range this will cause the value to be 1millisecond.
				So(query.IntervalMS, ShouldEqual, 1)
				So(query.MaxDataPoints, ShouldEqual, interval.DefaultRes)
			}

			applyScenario(timeRange, dataSourceJson, queryModel, verifier)
		})
	})
}

type queryIntervalTestContext struct {
	result    *alerting.EvalContext
	condition *QueryCondition
}

type queryIntervalVerifier func(query plugins.DataSubQuery)

type fakeIntervalTestReqHandler struct {
	//nolint: staticcheck // plugins.DataResponse deprecated
	response plugins.DataResponse
	verifier queryIntervalVerifier
}

//nolint: staticcheck // plugins.DataResponse deprecated
func (rh fakeIntervalTestReqHandler) HandleRequest(ctx context.Context, dsInfo *models.DataSource, query plugins.DataQuery) (
	plugins.DataResponse, error) {
	q := query.Queries[0]
	rh.verifier(q)
	return rh.response, nil
}

//nolint: staticcheck // plugins.DataResponse deprecated
func applyScenario(timeRange string, dataSourceJsonData *simplejson.Json, queryModel string, verifier func(query plugins.DataSubQuery)) {
	Convey("desc", func() {
		bus.AddHandler("test", func(query *models.GetDataSourceQuery) error {
			query.Result = &models.DataSource{Id: 1, Type: "graphite", JsonData: dataSourceJsonData}
			return nil
		})

		ctx := &queryIntervalTestContext{}
		ctx.result = &alerting.EvalContext{
			Rule:             &alerting.Rule{},
			RequestValidator: &validations.OSSPluginRequestValidator{},
		}

		jsonModel, err := simplejson.NewJson([]byte(`{
            "type": "query",
            "query":  {
              "params": ["A", "` + timeRange + `", "now"],
              "datasourceId": 1,
              "model": ` + queryModel + `
            },
            "reducer":{"type": "avg"},
					"evaluator":{"type": "gt", "params": [100]}
          }`))
		So(err, ShouldBeNil)

		condition, err := newQueryCondition(jsonModel, 0)
		So(err, ShouldBeNil)

		ctx.condition = condition

		qr := plugins.DataQueryResult{}

		reqHandler := fakeIntervalTestReqHandler{
			response: plugins.DataResponse{
				Results: map[string]plugins.DataQueryResult{
					"A": qr,
				},
			},
			verifier: verifier,
		}

		_, err = condition.Eval(ctx.result, reqHandler)

		So(err, ShouldBeNil)
	})
}
