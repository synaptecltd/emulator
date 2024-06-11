package mathfuncs

import (
	"math"
	"testing"
)

func TestTrendFunctions(t *testing.T) {
	testCases := []struct {
		name     string // name of the function, defined in the TrendFunctions map
		function func(float64, float64, float64) float64
		t        float64 // time in seconds
		A        float64 // amplitude
		T        float64 // period in seconds
		expected float64 // expected value of the function at time t
	}{
		{
			name:     "linear",
			function: TrendFunctions["linear"],
			t:        2.0,
			A:        1.0,
			T:        3.0,
			expected: 5.0,
		},
		{
			name:     "sine",
			function: TrendFunctions["sine"],
			t:        math.Pi / 2,
			A:        0.0,
			T:        1.0,
			expected: 1.0,
		},
		{
			name:     "cosine",
			function: TrendFunctions["cosine"],
			t:        math.Pi / 2,
			A:        0.0,
			T:        1.0,
			expected: 0.0,
		},
		// Add more test cases for other trend functions
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// get the function from the name

			result := tc.function(tc.t, tc.A, tc.T)
			if result != tc.expected {
				t.Errorf("Expected %f, but got %f", tc.expected, result)
			}
		})
	}
}
