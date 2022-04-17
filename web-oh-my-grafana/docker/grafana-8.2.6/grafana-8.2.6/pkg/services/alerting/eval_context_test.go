package alerting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/services/validations"

	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestStateIsUpdatedWhenNeeded(t *testing.T) {
	ctx := NewEvalContext(context.TODO(), &Rule{Conditions: []Condition{&conditionStub{firing: true}}}, &validations.OSSPluginRequestValidator{})

	t.Run("ok -> alerting", func(t *testing.T) {
		ctx.PrevAlertState = models.AlertStateOK
		ctx.Rule.State = models.AlertStateAlerting

		if !ctx.shouldUpdateAlertState() {
			t.Fatalf("expected should updated to be true")
		}
	})

	t.Run("ok -> ok", func(t *testing.T) {
		ctx.PrevAlertState = models.AlertStateOK
		ctx.Rule.State = models.AlertStateOK

		if ctx.shouldUpdateAlertState() {
			t.Fatalf("expected should updated to be false")
		}
	})
}

func TestGetStateFromEvalContext(t *testing.T) {
	tcs := []struct {
		name     string
		expected models.AlertStateType
		applyFn  func(ec *EvalContext)
	}{
		{
			name:     "ok -> alerting",
			expected: models.AlertStateAlerting,
			applyFn: func(ec *EvalContext) {
				ec.Firing = true
				ec.PrevAlertState = models.AlertStateOK
			},
		},
		{
			name:     "ok -> error(alerting)",
			expected: models.AlertStateAlerting,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStateOK
				ec.Error = errors.New("test error")
				ec.Rule.ExecutionErrorState = models.ExecutionErrorSetAlerting
			},
		},
		{
			name:     "ok -> pending. since its been firing for less than FOR",
			expected: models.AlertStatePending,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStateOK
				ec.Firing = true
				ec.Rule.LastStateChange = time.Now().Add(-time.Minute * 2)
				ec.Rule.For = time.Minute * 5
			},
		},
		{
			name:     "ok -> pending. since it has to be pending longer than FOR and prev state is ok",
			expected: models.AlertStatePending,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStateOK
				ec.Firing = true
				ec.Rule.LastStateChange = time.Now().Add(-(time.Hour * 5))
				ec.Rule.For = time.Minute * 2
			},
		},
		{
			name:     "pending -> alerting. since its been firing for more than FOR and prev state is pending",
			expected: models.AlertStateAlerting,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStatePending
				ec.Firing = true
				ec.Rule.LastStateChange = time.Now().Add(-(time.Hour * 5))
				ec.Rule.For = time.Minute * 2
			},
		},
		{
			name:     "alerting -> alerting. should not update regardless of FOR",
			expected: models.AlertStateAlerting,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStateAlerting
				ec.Firing = true
				ec.Rule.LastStateChange = time.Now().Add(-time.Minute * 5)
				ec.Rule.For = time.Minute * 2
			},
		},
		{
			name:     "ok -> ok. should not update regardless of FOR",
			expected: models.AlertStateOK,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStateOK
				ec.Rule.LastStateChange = time.Now().Add(-time.Minute * 5)
				ec.Rule.For = time.Minute * 2
			},
		},
		{
			name:     "ok -> error(keep_last)",
			expected: models.AlertStateOK,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStateOK
				ec.Error = errors.New("test error")
				ec.Rule.ExecutionErrorState = models.ExecutionErrorKeepState
			},
		},
		{
			name:     "pending -> error(keep_last)",
			expected: models.AlertStatePending,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStatePending
				ec.Error = errors.New("test error")
				ec.Rule.ExecutionErrorState = models.ExecutionErrorKeepState
			},
		},
		{
			name:     "ok -> no_data(alerting)",
			expected: models.AlertStateAlerting,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStateOK
				ec.Rule.NoDataState = models.NoDataSetAlerting
				ec.NoDataFound = true
			},
		},
		{
			name:     "ok -> no_data(keep_last)",
			expected: models.AlertStateOK,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStateOK
				ec.Rule.NoDataState = models.NoDataKeepState
				ec.NoDataFound = true
			},
		},
		{
			name:     "pending -> no_data(keep_last)",
			expected: models.AlertStatePending,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStatePending
				ec.Rule.NoDataState = models.NoDataKeepState
				ec.NoDataFound = true
			},
		},
		{
			name:     "pending -> no_data(alerting) with for duration have not passed",
			expected: models.AlertStatePending,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStatePending
				ec.Rule.NoDataState = models.NoDataSetAlerting
				ec.NoDataFound = true
				ec.Rule.For = time.Minute * 5
				ec.Rule.LastStateChange = time.Now().Add(-time.Minute * 2)
			},
		},
		{
			name:     "pending -> no_data(alerting) should set alerting since time passed FOR",
			expected: models.AlertStateAlerting,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStatePending
				ec.Rule.NoDataState = models.NoDataSetAlerting
				ec.NoDataFound = true
				ec.Rule.For = time.Minute * 2
				ec.Rule.LastStateChange = time.Now().Add(-time.Minute * 5)
			},
		},
		{
			name:     "pending -> error(alerting) with for duration have not passed ",
			expected: models.AlertStatePending,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStatePending
				ec.Rule.ExecutionErrorState = models.ExecutionErrorSetAlerting
				ec.Error = errors.New("test error")
				ec.Rule.For = time.Minute * 5
				ec.Rule.LastStateChange = time.Now().Add(-time.Minute * 2)
			},
		},
		{
			name:     "pending -> error(alerting) should set alerting since time passed FOR",
			expected: models.AlertStateAlerting,
			applyFn: func(ec *EvalContext) {
				ec.PrevAlertState = models.AlertStatePending
				ec.Rule.ExecutionErrorState = models.ExecutionErrorSetAlerting
				ec.Error = errors.New("test error")
				ec.Rule.For = time.Minute * 2
				ec.Rule.LastStateChange = time.Now().Add(-time.Minute * 5)
			},
		},
	}

	for _, tc := range tcs {
		evalContext := NewEvalContext(context.Background(), &Rule{Conditions: []Condition{&conditionStub{firing: true}}}, &validations.OSSPluginRequestValidator{})

		tc.applyFn(evalContext)
		newState := evalContext.GetNewState()
		assert.Equal(t, tc.expected, newState, "failed: %s \n expected '%s' have '%s'\n", tc.name, tc.expected, string(newState))
	}
}

func TestBuildTemplateDataMap(t *testing.T) {
	tcs := []struct {
		name     string
		matches  []*EvalMatch
		expected map[string]string
	}{
		{
			name: "single match",
			matches: []*EvalMatch{
				{
					Tags: map[string]string{
						"InstanceId": "i-123456789",
						"Percentile": "0.999",
					},
				},
			},
			expected: map[string]string{
				"InstanceId": "i-123456789",
				"Percentile": "0.999",
			},
		},
		{
			name: "matches with duplicate keys",
			matches: []*EvalMatch{
				{
					Tags: map[string]string{
						"InstanceId": "i-123456789",
					},
				},
				{
					Tags: map[string]string{
						"InstanceId": "i-987654321",
						"Percentile": "0.999",
					},
				},
			},
			expected: map[string]string{
				"InstanceId": "i-123456789, i-987654321",
				"Percentile": "0.999",
			},
		},
		{
			name: "matches with duplicate keys and values",
			matches: []*EvalMatch{
				{
					Tags: map[string]string{
						"InstanceId": "i-123456789",
						"Percentile": "0.999",
					},
				},
				{
					Tags: map[string]string{
						"InstanceId": "i-987654321",
						"Percentile": "0.995",
					},
				},
				{
					Tags: map[string]string{
						"InstanceId": "i-987654321",
						"Percentile": "0.999",
					},
				},
			},
			expected: map[string]string{
				"InstanceId": "i-123456789, i-987654321",
				"Percentile": "0.999, 0.995",
			},
		},
		{
			name: "a value and its substring for same key",
			matches: []*EvalMatch{
				{
					Tags: map[string]string{
						"Percentile": "0.9990",
					},
				},
				{
					Tags: map[string]string{
						"Percentile": "0.999",
					},
				},
			},
			expected: map[string]string{
				"Percentile": "0.9990, 0.999",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result, err := buildTemplateDataMap(tc.matches)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result, "failed: %s \n expected '%s' have '%s'\n", tc.name, tc.expected, result)
		})
	}
}

func TestEvaluateTemplate(t *testing.T) {
	tcs := []struct {
		name     string
		message  string
		data     map[string]string
		expected string
	}{
		{
			name:    "matching terms",
			message: "Degraded ${percentile} latency on ${instance}",
			data: map[string]string{
				"instance":   "i-123456789",
				"percentile": "0.95",
			},
			expected: "Degraded 0.95 latency on i-123456789",
		},
		{
			name:    "non-matching terms",
			message: "Degraded $percentile latency for endpoint ${ endpoint } on ${instance}",
			data: map[string]string{
				"INSTANCE":   "i-123456789",
				"percentile": "0.95",
				"endpoint":   "/api/dashboard/123",
			},
			expected: "Degraded $percentile latency for endpoint ${ endpoint } on ${instance}",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result, err := evaluateTemplate(tc.message, tc.data)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result, "failed: %s \n expected '%s' have '%s'\n", tc.name, tc.expected, result)
		})
	}
}
