package classic

import (
	"fmt"

	"github.com/grafana/grafana/pkg/expr/mathexp"
)

type evaluator interface {
	Eval(mathexp.Number) bool
}

type noValueEvaluator struct{}

type thresholdEvaluator struct {
	Type      string
	Threshold float64
}

type rangedEvaluator struct {
	Type  string
	Lower float64
	Upper float64
}

// newAlertEvaluator is a factory function for returning
// an AlertEvaluator depending on evaluation operator.
func newAlertEvaluator(model ConditionEvalJSON) (evaluator, error) {
	switch model.Type {
	case "gt", "lt":
		return newThresholdEvaluator(model)
	case "within_range", "outside_range":
		return newRangedEvaluator(model)
	case "no_value":
		return &noValueEvaluator{}, nil
	}

	return nil, fmt.Errorf("evaluator invalid evaluator type: %s", model.Type)
}

func (e *thresholdEvaluator) Eval(reducedValue mathexp.Number) bool {
	fv := reducedValue.GetFloat64Value()
	if fv == nil {
		return false
	}

	switch e.Type {
	case "gt":
		return *fv > e.Threshold
	case "lt":
		return *fv < e.Threshold
	}

	return false
}

func newThresholdEvaluator(model ConditionEvalJSON) (*thresholdEvaluator, error) {
	if len(model.Params) == 0 {
		return nil, fmt.Errorf("evaluator '%v' is missing the threshold parameter", model.Type)
	}

	return &thresholdEvaluator{
		Type:      model.Type,
		Threshold: model.Params[0],
	}, nil
}

func (e *noValueEvaluator) Eval(reducedValue mathexp.Number) bool {
	return reducedValue.GetFloat64Value() == nil
}

func newRangedEvaluator(model ConditionEvalJSON) (*rangedEvaluator, error) {
	if len(model.Params) != 2 {
		return nil, fmt.Errorf("ranged evaluator requires 2 parameters")
	}

	return &rangedEvaluator{
		Type:  model.Type,
		Lower: model.Params[0],
		Upper: model.Params[1],
	}, nil
}

func (e *rangedEvaluator) Eval(reducedValue mathexp.Number) bool {
	fv := reducedValue.GetFloat64Value()
	if fv == nil {
		return false
	}

	switch e.Type {
	case "within_range":
		return (e.Lower < *fv && e.Upper > *fv) || (e.Upper < *fv && e.Lower > *fv)
	case "outside_range":
		return (e.Upper < *fv && e.Lower < *fv) || (e.Upper > *fv && e.Lower > *fv)
	}

	return false
}
