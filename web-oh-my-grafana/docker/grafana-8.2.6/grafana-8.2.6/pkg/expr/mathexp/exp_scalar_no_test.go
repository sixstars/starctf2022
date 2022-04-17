package mathexp

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScalarExpr(t *testing.T) {
	var tests = []struct {
		name      string
		expr      string
		vars      Vars
		newErrIs  assert.ErrorAssertionFunc
		execErrIs assert.ErrorAssertionFunc
		resultIs  assert.ComparisonAssertionFunc
		Results   Results
	}{
		{
			name:      "a scalar",
			expr:      "1",
			vars:      Vars{},
			newErrIs:  assert.NoError,
			execErrIs: assert.NoError,
			resultIs:  assert.Equal,
			Results:   Results{[]Value{NewScalar("", float64Pointer(1.0))}},
		},
		{
			name:      "unary: scalar",
			expr:      "! 1.2",
			vars:      Vars{},
			newErrIs:  assert.NoError,
			execErrIs: assert.NoError,
			resultIs:  assert.Equal,
			Results:   Results{[]Value{NewScalar("", float64Pointer(0.0))}},
		},
		{
			name:      "binary: scalar Op scalar",
			expr:      "1 + 1",
			vars:      Vars{},
			newErrIs:  assert.NoError,
			execErrIs: assert.NoError,
			resultIs:  assert.Equal,
			Results:   Results{[]Value{NewScalar("", float64Pointer(2.0))}},
		},
		{
			name:      "binary: scalar Op scalar - divide by zero",
			expr:      "1 / 0",
			vars:      Vars{},
			newErrIs:  assert.NoError,
			execErrIs: assert.NoError,
			resultIs:  assert.Equal,
			Results:   Results{[]Value{NewScalar("", float64Pointer(math.Inf(1)))}},
		},
		{
			name:      "binary: scalar Op number",
			expr:      "1 + $A",
			vars:      Vars{"A": Results{[]Value{makeNumber("temp", nil, float64Pointer(2.0))}}},
			newErrIs:  assert.NoError,
			execErrIs: assert.NoError,
			resultIs:  assert.Equal,
			Results:   Results{[]Value{makeNumber("", nil, float64Pointer(3.0))}},
		},
		{
			name:      "binary: number Op Scalar",
			expr:      "$A - 3",
			vars:      Vars{"A": Results{[]Value{makeNumber("temp", nil, float64Pointer(2.0))}}},
			newErrIs:  assert.NoError,
			execErrIs: assert.NoError,
			resultIs:  assert.Equal,
			Results:   Results{[]Value{makeNumber("", nil, float64Pointer(-1))}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := New(tt.expr)
			tt.newErrIs(t, err)
			if e != nil {
				res, err := e.Execute("", tt.vars)
				tt.execErrIs(t, err)
				tt.resultIs(t, tt.Results, res)
			}
		})
	}
}

func TestNumberExpr(t *testing.T) {
	var tests = []struct {
		name      string
		expr      string
		vars      Vars
		newErrIs  assert.ErrorAssertionFunc
		execErrIs assert.ErrorAssertionFunc
		resultIs  assert.ComparisonAssertionFunc
		results   Results
		willPanic bool
	}{
		{
			name:      "binary: number Op Scalar",
			expr:      "$A / $A",
			vars:      Vars{"A": Results{[]Value{makeNumber("temp", nil, float64Pointer(2.0))}}},
			newErrIs:  assert.NoError,
			execErrIs: assert.NoError,
			resultIs:  assert.Equal,
			results:   Results{[]Value{makeNumber("", nil, float64Pointer(1))}},
		},
		{
			name:      "unary: number",
			expr:      "- $A",
			vars:      Vars{"A": Results{[]Value{makeNumber("temp", nil, float64Pointer(2.0))}}},
			newErrIs:  assert.NoError,
			execErrIs: assert.NoError,
			resultIs:  assert.Equal,
			results:   Results{[]Value{makeNumber("", nil, float64Pointer(-2.0))}},
		},
		{
			name:      "binary: Scalar Op Number (Number will nil val) - currently Panics",
			expr:      "1 + $A",
			vars:      Vars{"A": Results{[]Value{makeNumber("", nil, nil)}}},
			willPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testBlock := func() {
				e, err := New(tt.expr)
				tt.newErrIs(t, err)
				if e != nil {
					res, err := e.Execute("", tt.vars)
					tt.execErrIs(t, err)
					tt.resultIs(t, tt.results, res)
				}
			}
			if tt.willPanic {
				assert.Panics(t, testBlock)
			} else {
				assert.NotPanics(t, testBlock)
			}
		})
	}
}
