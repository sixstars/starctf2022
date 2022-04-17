// Package eval executes the condition for an alert definition, evaluates the condition results, and
// returns the alert instance states.
package eval

import (
	"context"
	"fmt"
	"runtime/debug"
	"sort"
	"time"

	"github.com/grafana/grafana/pkg/expr/classic"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/ngalert/models"

	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/tsdb"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/expr"
)

type Evaluator struct {
	Cfg *setting.Cfg
	Log log.Logger
}

// invalidEvalResultFormatError is an error for invalid format of the alert definition evaluation results.
type invalidEvalResultFormatError struct {
	refID  string
	reason string
	err    error
}

func (e *invalidEvalResultFormatError) Error() string {
	s := fmt.Sprintf("invalid format of evaluation results for the alert definition %s: %s", e.refID, e.reason)
	if e.err != nil {
		s = fmt.Sprintf("%s: %s", s, e.err.Error())
	}
	return s
}

func (e *invalidEvalResultFormatError) Unwrap() error {
	return e.err
}

// ExecutionResults contains the unevaluated results from executing
// a condition.
type ExecutionResults struct {
	Error error

	Results data.Frames
}

// Results is a slice of evaluated alert instances states.
type Results []Result

// Result contains the evaluated State of an alert instance
// identified by its labels.
type Result struct {
	Instance data.Labels
	State    State // Enum
	// Error message for Error state. should be nil if State != Error.
	Error              error
	EvaluatedAt        time.Time
	EvaluationDuration time.Duration

	// EvaluationString is a string representation of evaluation data such
	// as EvalMatches (from "classic condition"), and in the future from operations
	// like SSE "math".
	EvaluationString string

	// Values contains the RefID and value of reduce and math expressions.
	// It does not contain values for classic conditions as the values
	// in classic conditions do not have a RefID.
	Values map[string]NumberValueCapture
}

// State is an enum of the evaluation State for an alert instance.
type State int

const (
	// Normal is the eval state for an alert instance condition
	// that evaluated to false.
	Normal State = iota

	// Alerting is the eval state for an alert instance condition
	// that evaluated to true (Alerting).
	Alerting

	// Pending is the eval state for an alert instance condition
	// that evaluated to true (Alerting) but has not yet met
	// the For duration defined in AlertRule.
	Pending

	// NoData is the eval state for an alert rule condition
	// that evaluated to NoData.
	NoData

	// Error is the eval state for an alert rule condition
	// that evaluated to Error.
	Error
)

func (s State) String() string {
	return [...]string{"Normal", "Alerting", "Pending", "NoData", "Error"}[s]
}

// AlertExecCtx is the context provided for executing an alert condition.
type AlertExecCtx struct {
	OrgID              int64
	ExpressionsEnabled bool
	Log                log.Logger

	Ctx context.Context
}

// GetExprRequest validates the condition and creates a expr.Request from it.
func GetExprRequest(ctx AlertExecCtx, data []models.AlertQuery, now time.Time) (*expr.Request, error) {
	req := &expr.Request{
		OrgId: ctx.OrgID,
		Headers: map[string]string{
			// Some data sources check this in query method as sometimes alerting needs special considerations.
			"FromAlert":    "true",
			"X-Cache-Skip": "true",
		},
	}

	for i := range data {
		q := data[i]
		model, err := q.GetModel()
		if err != nil {
			return nil, fmt.Errorf("failed to get query model: %w", err)
		}
		interval, err := q.GetIntervalDuration()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve intervalMs from the model: %w", err)
		}

		maxDatapoints, err := q.GetMaxDatapoints()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve maxDatapoints from the model: %w", err)
		}

		req.Queries = append(req.Queries, expr.Query{
			TimeRange: expr.TimeRange{
				From: q.RelativeTimeRange.ToTimeRange(now).From,
				To:   q.RelativeTimeRange.ToTimeRange(now).To,
			},
			DatasourceUID: q.DatasourceUID,
			JSON:          model,
			Interval:      interval,
			RefID:         q.RefID,
			MaxDataPoints: maxDatapoints,
			QueryType:     q.QueryType,
		})
	}
	return req, nil
}

type NumberValueCapture struct {
	Var    string // RefID
	Labels data.Labels
	Value  *float64
}

func executeCondition(ctx AlertExecCtx, c *models.Condition, now time.Time, dataService *tsdb.Service) ExecutionResults {
	result := ExecutionResults{}

	execResp, err := executeQueriesAndExpressions(ctx, c.Data, now, dataService)

	if err != nil {
		return ExecutionResults{Error: err}
	}

	// eval captures for the '__value_string__' annotation and the Value property of the API response.
	captures := make([]NumberValueCapture, 0, len(execResp.Responses))

	captureVal := func(refID string, labels data.Labels, value *float64) {
		captures = append(captures, NumberValueCapture{
			Var:    refID,
			Value:  value,
			Labels: labels.Copy(),
		})
	}

	for refID, res := range execResp.Responses {
		// for each frame within each response, the response can contain several data types including time-series data.
		// For now, we favour simplicity and only care about single scalar values.
		for _, frame := range res.Frames {
			if len(frame.Fields) != 1 || frame.Fields[0].Type() != data.FieldTypeNullableFloat64 {
				continue
			}
			var v *float64
			if frame.Fields[0].Len() == 1 {
				v = frame.At(0, 0).(*float64) // type checked above
			}
			captureVal(frame.RefID, frame.Fields[0].Labels, v)
		}

		if refID == c.Condition {
			result.Results = res.Frames
		}
	}

	// add capture values as data frame metadata to each result (frame) that has matching labels.
	for _, frame := range result.Results {
		// classic conditions already have metadata set and only have one value, there's no need to add anything in this case.
		if frame.Meta != nil && frame.Meta.Custom != nil {
			if _, ok := frame.Meta.Custom.([]classic.EvalMatch); ok {
				continue // do not overwrite EvalMatch from classic condition.
			}
		}

		frame.SetMeta(&data.FrameMeta{}) // overwrite metadata

		if len(frame.Fields) == 1 {
			theseLabels := frame.Fields[0].Labels
			for _, cap := range captures {
				// matching labels are equal labels, or when one set of labels includes the labels of the other.
				if theseLabels.Equals(cap.Labels) || theseLabels.Contains(cap.Labels) || cap.Labels.Contains(theseLabels) {
					if frame.Meta.Custom == nil {
						frame.Meta.Custom = []NumberValueCapture{}
					}
					frame.Meta.Custom = append(frame.Meta.Custom.([]NumberValueCapture), cap)
				}
			}
		}
	}

	return result
}

func executeQueriesAndExpressions(ctx AlertExecCtx, data []models.AlertQuery, now time.Time, dataService *tsdb.Service) (resp *backend.QueryDataResponse, err error) {
	defer func() {
		if e := recover(); e != nil {
			ctx.Log.Error("alert rule panic", "error", e, "stack", string(debug.Stack()))
			panicErr := fmt.Errorf("alert rule panic; please check the logs for the full stack")
			if err != nil {
				err = fmt.Errorf("queries and expressions execution failed: %w; %v", err, panicErr.Error())
			} else {
				err = panicErr
			}
		}
	}()

	queryDataReq, err := GetExprRequest(ctx, data, now)
	if err != nil {
		return nil, err
	}

	exprService := expr.Service{
		Cfg:         &setting.Cfg{ExpressionsEnabled: ctx.ExpressionsEnabled},
		DataService: dataService,
	}
	return exprService.TransformData(ctx.Ctx, queryDataReq)
}

// evaluateExecutionResult takes the ExecutionResult which includes data.Frames returned
// from SSE (Server Side Expressions). It will create Results (slice of Result) with a State
// extracted from each Frame.
//
// If the ExecutionResults error property is not nil, a single Error result will be returned.
// If there is no error and no results then a single NoData state Result will be returned.
//
// Each non-empty Frame must be a single Field of type []*float64 and of length 1.
// Also, each Frame must be uniquely identified by its Field.Labels or a single Error result will be returned.
//
// Per Frame, data becomes a State based on the following rules:
//  - Empty or zero length Frames result in NoData.
//  - If a value:
//    - 0 results in Normal.
//    - Nonzero (e.g 1.2, NaN) results in Alerting.
//    - nil results in noData.
//    - unsupported Frame schemas results in Error.
func evaluateExecutionResult(execResults ExecutionResults, ts time.Time) Results {
	evalResults := make([]Result, 0)

	appendErrRes := func(e error) {
		evalResults = append(evalResults, Result{
			State:              Error,
			Error:              e,
			EvaluatedAt:        ts,
			EvaluationDuration: time.Since(ts),
		})
	}

	appendNoData := func(l data.Labels) {
		evalResults = append(evalResults, Result{
			State:              NoData,
			Instance:           l,
			EvaluatedAt:        ts,
			EvaluationDuration: time.Since(ts),
		})
	}

	if execResults.Error != nil {
		appendErrRes(execResults.Error)
		return evalResults
	}

	if len(execResults.Results) == 0 {
		appendNoData(nil)
		return evalResults
	}

	for _, f := range execResults.Results {
		rowLen, err := f.RowLen()
		if err != nil {
			appendErrRes(&invalidEvalResultFormatError{refID: f.RefID, reason: "unable to get frame row length", err: err})
			continue
		}

		if len(f.TypeIndices(data.FieldTypeTime, data.FieldTypeNullableTime)) > 0 {
			appendErrRes(&invalidEvalResultFormatError{refID: f.RefID, reason: "looks like time series data, only reduced data can be alerted on."})
			continue
		}

		if rowLen == 0 {
			if len(f.Fields) == 0 {
				appendNoData(nil)
				continue
			}
			if len(f.Fields) == 1 {
				appendNoData(f.Fields[0].Labels)
				continue
			}
		}

		if rowLen > 1 {
			appendErrRes(&invalidEvalResultFormatError{refID: f.RefID, reason: fmt.Sprintf("unexpected row length: %d instead of 0 or 1", rowLen)})
			continue
		}

		if len(f.Fields) > 1 {
			appendErrRes(&invalidEvalResultFormatError{refID: f.RefID, reason: fmt.Sprintf("unexpected field length: %d instead of 1", len(f.Fields))})
			continue
		}

		if f.Fields[0].Type() != data.FieldTypeNullableFloat64 {
			appendErrRes(&invalidEvalResultFormatError{refID: f.RefID, reason: fmt.Sprintf("invalid field type: %s", f.Fields[0].Type())})
			continue
		}

		val := f.Fields[0].At(0).(*float64) // type checked by data.FieldTypeNullableFloat64 above

		r := Result{
			Instance:           f.Fields[0].Labels,
			EvaluatedAt:        ts,
			EvaluationDuration: time.Since(ts),
			EvaluationString:   extractEvalString(f),
			Values:             extractValues(f),
		}

		switch {
		case val == nil:
			r.State = NoData
		case *val == 0:
			r.State = Normal
		default:
			r.State = Alerting
		}

		evalResults = append(evalResults, r)
	}

	seenLabels := make(map[string]bool)
	for _, res := range evalResults {
		labelsStr := res.Instance.String()
		_, ok := seenLabels[labelsStr]
		if ok {
			return Results{
				Result{
					State:              Error,
					Instance:           res.Instance,
					EvaluatedAt:        ts,
					EvaluationDuration: time.Since(ts),
					Error:              &invalidEvalResultFormatError{reason: fmt.Sprintf("frame cannot uniquely be identified by its labels: has duplicate results with labels {%s}", labelsStr)},
				},
			}
		}
		seenLabels[labelsStr] = true
	}

	return evalResults
}

// AsDataFrame forms the EvalResults in Frame suitable for displaying in the table panel of the front end.
// It displays one row per alert instance, with a column for each label and one for the alerting state.
func (evalResults Results) AsDataFrame() data.Frame {
	fieldLen := len(evalResults)

	uniqueLabelKeys := make(map[string]struct{})

	for _, evalResult := range evalResults {
		for k := range evalResult.Instance {
			uniqueLabelKeys[k] = struct{}{}
		}
	}

	labelColumns := make([]string, 0, len(uniqueLabelKeys))
	for k := range uniqueLabelKeys {
		labelColumns = append(labelColumns, k)
	}

	labelColumns = sort.StringSlice(labelColumns)

	frame := data.NewFrame("evaluation results")
	for _, lKey := range labelColumns {
		frame.Fields = append(frame.Fields, data.NewField(lKey, nil, make([]string, fieldLen)))
	}
	frame.Fields = append(frame.Fields, data.NewField("State", nil, make([]string, fieldLen)))
	frame.Fields = append(frame.Fields, data.NewField("Info", nil, make([]string, fieldLen)))

	for evalIdx, evalResult := range evalResults {
		for lIdx, v := range labelColumns {
			frame.Set(lIdx, evalIdx, evalResult.Instance[v])
		}

		frame.Set(len(labelColumns), evalIdx, evalResult.State.String())

		switch {
		case evalResult.Error != nil:
			frame.Set(len(labelColumns)+1, evalIdx, evalResult.Error.Error())
		case evalResult.EvaluationString != "":
			frame.Set(len(labelColumns)+1, evalIdx, evalResult.EvaluationString)
		}
	}
	return *frame
}

// ConditionEval executes conditions and evaluates the result.
func (e *Evaluator) ConditionEval(condition *models.Condition, now time.Time, dataService *tsdb.Service) (Results, error) {
	alertCtx, cancelFn := context.WithTimeout(context.Background(), e.Cfg.UnifiedAlerting.EvaluationTimeout)
	defer cancelFn()

	alertExecCtx := AlertExecCtx{OrgID: condition.OrgID, Ctx: alertCtx, ExpressionsEnabled: e.Cfg.ExpressionsEnabled, Log: e.Log}

	execResult := executeCondition(alertExecCtx, condition, now, dataService)

	evalResults := evaluateExecutionResult(execResult, now)
	return evalResults, nil
}

// QueriesAndExpressionsEval executes queries and expressions and returns the result.
func (e *Evaluator) QueriesAndExpressionsEval(orgID int64, data []models.AlertQuery, now time.Time, dataService *tsdb.Service) (*backend.QueryDataResponse, error) {
	alertCtx, cancelFn := context.WithTimeout(context.Background(), e.Cfg.UnifiedAlerting.EvaluationTimeout)
	defer cancelFn()

	alertExecCtx := AlertExecCtx{OrgID: orgID, Ctx: alertCtx, ExpressionsEnabled: e.Cfg.ExpressionsEnabled, Log: e.Log}

	execResult, err := executeQueriesAndExpressions(alertExecCtx, data, now, dataService)
	if err != nil {
		return nil, fmt.Errorf("failed to execute conditions: %w", err)
	}

	return execResult, nil
}
