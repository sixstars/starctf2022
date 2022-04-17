package expr

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServicebuildPipeLine(t *testing.T) {
	var tests = []struct {
		name              string
		req               *Request
		expectedOrder     []string
		expectErrContains string
	}{
		{
			name: "simple: a requires b",
			req: &Request{
				Queries: []Query{
					{
						RefID:         "A",
						DatasourceUID: DatasourceUID,
						JSON: json.RawMessage(`{
							"expression": "B",
							"reducer": "mean",
							"type": "reduce"
						}`),
					},
					{
						RefID:         "B",
						DatasourceUID: "Fake",
					},
				},
			},
			expectedOrder: []string{"B", "A"},
		},
		{
			name: "cycle will error",
			req: &Request{
				Queries: []Query{
					{
						RefID:         "A",
						DatasourceUID: DatasourceUID,
						JSON: json.RawMessage(`{
								"expression": "$B",
								"type": "math"
							}`),
					},
					{
						RefID:         "B",
						DatasourceUID: DatasourceUID,
						JSON: json.RawMessage(`{
								"expression": "$A",
								"type": "math"
							}`),
					},
				},
			},
			expectErrContains: "cyclic components",
		},
		{
			name: "self reference will error",
			req: &Request{
				Queries: []Query{
					{
						RefID:         "A",
						DatasourceUID: DatasourceUID,
						JSON: json.RawMessage(`{
								"expression": "$A",
								"type": "math"
							}`),
					},
				},
			},
			expectErrContains: "self referencing node",
		},
		{
			name: "missing dependency will error",
			req: &Request{
				Queries: []Query{
					{
						RefID:         "A",
						DatasourceUID: DatasourceUID,
						JSON: json.RawMessage(`{
								"expression": "$B",
								"type": "math"
							}`),
					},
				},
			},
			expectErrContains: "find dependent",
		},
		{
			name: "classic can not take input from another expression",
			req: &Request{
				Queries: []Query{
					{
						RefID:         "A",
						DatasourceUID: DatasourceUID,
						JSON: json.RawMessage(`{
							"type": "classic_conditions",
							"conditions": [
								{
									"evaluator": {
									"params": [
										2,
										3
									],
									"type": "within_range"
									},
									"operator": {
									"type": "or"
									},
									"query": {
									"params": [
										"B"
									]
									},
									"reducer": {
									"params": [],
									"type": "diff"
									},
									"type": "query"
								}
							]
						}`),
					},
					{
						RefID:         "B",
						DatasourceUID: DatasourceUID,
						JSON: json.RawMessage(`{
							"expression": "C",
							"reducer": "mean",
							"type": "reduce"
						}`),
					},
					{
						RefID:         "C",
						DatasourceUID: "Fake",
					},
				},
			},
			expectErrContains: "only data source queries may be inputs to a classic condition",
		},
		{
			name: "classic can not output to another expression",
			req: &Request{
				Queries: []Query{
					{
						RefID:         "A",
						DatasourceUID: DatasourceUID,
						JSON: json.RawMessage(`{
							"type": "classic_conditions",
							"conditions": [
								{
									"evaluator": {
									"params": [
										2,
										3
									],
									"type": "within_range"
									},
									"operator": {
									"type": "or"
									},
									"query": {
									"params": [
										"C"
									]
									},
									"reducer": {
									"params": [],
									"type": "diff"
									},
									"type": "query"
								}
							]
						}`),
					},
					{
						RefID:         "B",
						DatasourceUID: DatasourceUID,
						JSON: json.RawMessage(`{
							"expression": "A",
							"reducer": "mean",
							"type": "reduce"
						}`),
					},
					{
						RefID:         "C",
						DatasourceUID: "Fake",
					},
				},
			},
			expectErrContains: "classic conditions may not be the input for other expressions",
		},
	}
	s := Service{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := s.buildPipeline(tt.req)
			if tt.expectErrContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectErrContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedOrder, getRefIDOrder(nodes))
			}
		})
	}
}

func getRefIDOrder(nodes []Node) []string {
	ids := make([]string, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, n.RefID())
	}
	return ids
}
