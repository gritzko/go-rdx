package rdx

import (
	"bytes"
	"testing"
)

func TestIntersectNext(t *testing.T) {
	// Helper function to create a multi-element stream
	multiStream := func(elements ...[]byte) []byte {
		var result []byte
		for _, elem := range elements {
			result = append(result, elem...)
		}
		return result
	}

	tests := []struct {
		name        string
		inputs      [][]byte
		totalInputs int
		expected    [][]byte // Expected sequence of outputs from IntersectNext calls
		expectError bool
	}{
		{
			name: "simple intersection - all have common element",
			inputs: [][]byte{
				multiStream(I0(1), I0(2), I0(3)),
				multiStream(I0(2), I0(4), I0(5)),
				multiStream(I0(2), I0(6), I0(7)),
			},
			totalInputs: 3,
			expected:    [][]byte{I0(2)}, // Only 2 appears in all inputs
			expectError: false,
		},
		{
			name: "no intersection",
			inputs: [][]byte{
				multiStream(I0(1), I0(2)),
				multiStream(I0(3), I0(4)),
				multiStream(I0(5), I0(6)),
			},
			totalInputs: 3,
			expected:    [][]byte{}, // No common elements
			expectError: false,
		},
		{
			name: "multiple intersections",
			inputs: [][]byte{
				multiStream(I0(1), I0(2), I0(3), I0(4)),
				multiStream(I0(1), I0(3), I0(5), I0(7)),
				multiStream(I0(1), I0(3), I0(6), I0(8)),
			},
			totalInputs: 3,
			expected:    [][]byte{I0(1), I0(3)}, // 1 and 3 appear in all inputs
			expectError: false,
		},
		{
			name: "single element intersection",
			inputs: [][]byte{
				I0(42),
				I0(42),
				I0(42),
			},
			totalInputs: 3,
			expected:    [][]byte{I0(42)},
			expectError: false,
		},
		{
			name: "different types - no intersection",
			inputs: [][]byte{
				I0(1),
				S0("1"),
				F0(1.0),
			},
			totalInputs: 3,
			expected:    [][]byte{}, // Different types, no intersection
			expectError: false,
		},
		{
			name: "string intersection",
			inputs: [][]byte{
				multiStream(S0("apple"), S0("banana"), S0("cherry")),
				multiStream(S0("banana"), S0("date"), S0("elderberry")),
				multiStream(S0("banana"), S0("fig"), S0("grape")),
			},
			totalInputs: 3,
			expected:    [][]byte{S0("banana")},
			expectError: false,
		},
		{
			name: "empty inputs",
			inputs: [][]byte{
				{},
				{},
				{},
			},
			totalInputs: 3,
			expected:    [][]byte{}, // No elements to intersect
			expectError: false,
		},
		{
			name: "one empty input",
			inputs: [][]byte{
				multiStream(I0(1), I0(2)),
				{}, // Empty input means no intersection possible
				multiStream(I0(1), I0(3)),
			},
			totalInputs: 3,
			expected:    [][]byte{}, // Empty input = no intersection
			expectError: false,
		},
		{
			name: "tuple intersection",
			inputs: [][]byte{
				multiStream(P0(S0("key"), I0(1)), P0(S0("key2"), I0(2))),
				multiStream(P0(S0("key"), I0(1)), P0(S0("key2"), I0(3))),
				multiStream(P0(S0("key"), I0(1)), P0(S0("key2"), I0(4))),
			},
			totalInputs: 3,
			expected:    [][]byte{P0(S0("key"), I0(1)), P0(S0("key2"), I0(2))},
			expectError: false,
		},
		{
			name: "reference intersection",
			inputs: [][]byte{
				multiStream(R0(ID{1, 1}), R0(ID{2, 2}), R0(ID{3, 3})),
				multiStream(R0(ID{2, 2}), R0(ID{4, 4})),
				multiStream(R0(ID{2, 2}), R0(ID{5, 5})),
			},
			totalInputs: 3,
			expected:    [][]byte{R0(ID{2, 2})},
			expectError: false,
		},
		{
			name: "complex intersection with mixed types",
			inputs: [][]byte{
				multiStream(I0(10), S0("common"), I0(20)),
				multiStream(S0("common"), F0(3.14), I0(30)),
				multiStream(I0(40), S0("common"), T0("term")),
			},
			totalInputs: 3,
			expected:    [][]byte{S0("common")}, // Only string "common" appears in all
			expectError: false,
		},
		{
			name: "early termination - first input exhausted",
			inputs: [][]byte{
				I0(1), // Only one element
				multiStream(I0(1), I0(2), I0(3)),
				multiStream(I0(1), I0(4), I0(5)),
			},
			totalInputs: 3,
			expected:    [][]byte{I0(1)}, // After first element, first input is exhausted
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create heap from inputs
			heap, err := Heapize(tt.inputs, CompareEuler)
			if err != nil {
				if !tt.expectError {
					t.Fatalf("Heapize failed: %v", err)
				}
				return
			}

			var result Iter
			var outputs [][]byte
			totalInputs := len(tt.inputs)

			// Keep calling IntersectNext until heap is exhausted or error
			for len(heap) > 0 && len(heap) == totalInputs && err == nil {
				result, err = heap.IntersectNext(CompareEuler)

				if result.HasData() {
					outputs = append(outputs, result.Record())
				}
			}

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check if we got the expected number of outputs
			if len(outputs) != len(tt.expected) {
				t.Errorf("Expected %d outputs, got %d", len(tt.expected), len(outputs))
				return
			}

			// Check each output matches expected
			for i, expected := range tt.expected {
				if !bytes.Equal(outputs[i], expected) {
					t.Errorf("Output %d: expected %v, got %v", i, expected, outputs[i])
				}
			}
		})
	}
}
