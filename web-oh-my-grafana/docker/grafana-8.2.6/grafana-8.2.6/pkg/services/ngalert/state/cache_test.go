package state

import (
	"errors"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/services/ngalert/eval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptr "github.com/xorcare/pointer"
)

func TestTemplateCaptureValueStringer(t *testing.T) {
	cases := []struct {
		name     string
		value    templateCaptureValue
		expected string
	}{{
		name:     "0 is returned as integer value",
		value:    templateCaptureValue{Value: 0},
		expected: "0",
	}, {
		name:     "1.0 is returned as integer value",
		value:    templateCaptureValue{Value: 1.0},
		expected: "1",
	}, {
		name:     "1.1 is returned as decimal value",
		value:    templateCaptureValue{Value: 1.1},
		expected: "1.1",
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.value.String())
		})
	}
}

func TestExpandTemplate(t *testing.T) {
	cases := []struct {
		name          string
		text          string
		alertInstance eval.Result
		labels        data.Labels
		expected      string
		expectedError error
	}{{
		name:     "labels are expanded into $labels",
		text:     "{{ $labels.instance }} is down",
		labels:   data.Labels{"instance": "foo"},
		expected: "foo is down",
	}, {
		name:          "missing label in $labels returns error",
		text:          "{{ $labels.instance }} is down",
		labels:        data.Labels{},
		expectedError: errors.New("error executing template __alert_test: template: __alert_test:1:86: executing \"__alert_test\" at <$labels.instance>: map has no entry for key \"instance\""),
	}, {
		name: "values are expanded into $values",
		text: "{{ $values.A.Labels.instance }} has value {{ $values.A }}",
		alertInstance: eval.Result{
			Values: map[string]eval.NumberValueCapture{
				"A": {
					Var:    "A",
					Labels: data.Labels{"instance": "foo"},
					Value:  ptr.Float64(1),
				},
			},
		},
		expected: "foo has value 1",
	}, {
		name: "values can be passed to template functions such as printf",
		text: "{{ $values.A.Labels.instance }} has value {{ $values.A.Value | printf \"%.1f\" }}",
		alertInstance: eval.Result{
			Values: map[string]eval.NumberValueCapture{
				"A": {
					Var:    "A",
					Labels: data.Labels{"instance": "foo"},
					Value:  ptr.Float64(1.1),
				},
			},
		},
		expected: "foo has value 1.1",
	}, {
		name: "missing label in $values returns error",
		text: "{{ $values.A.Labels.instance }} has value {{ $values.A }}",
		alertInstance: eval.Result{
			Values: map[string]eval.NumberValueCapture{
				"A": {
					Var:    "A",
					Labels: data.Labels{},
					Value:  ptr.Float64(1),
				},
			},
		},
		expectedError: errors.New("error executing template __alert_test: template: __alert_test:1:86: executing \"__alert_test\" at <$values.A.Labels.instance>: map has no entry for key \"instance\""),
	}, {
		name: "missing value in $values is returned as NaN",
		text: "{{ $values.A.Labels.instance }} has value {{ $values.A }}",
		alertInstance: eval.Result{
			Values: map[string]eval.NumberValueCapture{
				"A": {
					Var:    "A",
					Labels: data.Labels{"instance": "foo"},
					Value:  nil,
				},
			},
		},
		expected: "foo has value NaN",
	}, {
		name: "assert value string is expanded into $value",
		text: "{{ $value }}",
		alertInstance: eval.Result{
			EvaluationString: "[ var='A' labels={instance=foo} value=10 ]",
		},
		expected: "[ var='A' labels={instance=foo} value=10 ]",
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v, err := expandTemplate("test", c.text, c.labels, c.alertInstance)
			require.Equal(t, c.expectedError, err)
			require.Equal(t, c.expected, v)
		})
	}
}
