package mathfuncs_test

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/synaptecltd/emulator/mathfuncs"
)

// Tests for non-random trend functions
func TestDeterministicTrendFunctions(t *testing.T) {
	testCases := []struct {
		name     string // name of the function, defined in the TrendFunctions map
		function func(float64, float64, float64) float64
		t        float64 // time in seconds
		A        float64 // amplitude
		T        float64 // period of the trend in seconds
		expected float64 // expected value of the function at time t
		isError  bool    // true if an error is expected
	}{
		{
			name:     "linear",
			t:        2.0,
			A:        3.0,
			T:        3.0,
			expected: 2.0, // y = (3/3)*2 = 2
			isError:  false,
		},
		{
			name:     "sine",
			t:        0.5,
			A:        5.0,
			T:        2,
			expected: 5.0, // A*sin(2*pi*0.5/2) = 5
			isError:  false,
		},
		{
			name:     "cosine",
			t:        0.5,
			A:        5.0,
			T:        2.0,
			expected: 0.0, // A*cos(2*pi*0.5/2) = 0
			isError:  false,
		},
		{
			name:     "exponential",
			t:        1.0,
			A:        2.0,
			T:        1.0,
			expected: 2*math.Exp(1) - 2, // A*exp(1) - A = 2*exp(1) - 2
			isError:  false,
		},
		{
			name:     "parabolic",
			t:        1.0,
			A:        4.0,
			T:        2.0,
			expected: 1.0, // A*(1/2)^2 = 4*(1/4) = 1
			isError:  false,
		},
		{
			name:     "step",
			t:        1.5,
			A:        6.0,
			T:        2.0,
			expected: 6.0, // positive value for t > T/2
			isError:  false,
		},
		{
			name:     "step",
			t:        0.0,
			A:        6.0,
			T:        2.0,
			expected: 0.0, // zero value for t < T/2
			isError:  false,
		},
		{
			name:     "square",
			t:        0.0,
			A:        6.0,
			T:        2.0,
			expected: 6.0, // positive value for sin >= 0
			isError:  false,
		},
		{
			name:     "square",
			t:        1.5,
			A:        6.0,
			T:        2.0,
			expected: -6.0, // negative value for sin < 0
			isError:  false,
		},
		{
			name:     "sawtooth",
			t:        5.0,
			A:        10.0,
			T:        1.0,
			expected: 0.0, // middle of sawtooth wave
			isError:  false,
		},
		{
			name:     "sawtooth",
			t:        2.5,
			A:        10.0,
			T:        10.0,
			expected: 5.0, // half way up the sawtooth wave A/2
			isError:  false,
		},
		{
			name:     "impulse",
			t:        0.5,
			A:        100.0,
			T:        10.0,
			expected: 0.0, // no impulse at t=0.5
			isError:  false,
		},
		{
			name:     "impulse",
			t:        10.0,
			A:        100.0,
			T:        10.0,
			expected: 100.0, // impulse at t=10
			isError:  false,
		},
		// Add more test cases for other trend functions
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// get the function from the name
			testFunction, err := mathfuncs.GetTrendFunctionFromName(tc.name)

			if tc.isError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			result := testFunction(tc.t, tc.A, tc.T)
			assert.InDelta(t, tc.expected, result, 1e-6)
		})
	}
}

// Tests for non-deteministic trend functions
func TestNoiseFunctions(t *testing.T) {
	A := 1.0 + rand.Float64()*9.0 // ampltiude of noise (between 1 and 10)

	type TestCase struct {
		name            string  // name of the function, defined in the TrendFunctions map
		numSamples      int     // number of samples of noise to generate, generate at least 1e6 samples if checking statistics
		checkStatistics bool    // whether test should check the mean and standard deviation of the noise
		expectedMean    float64 // expected mean of the noise
		expectedStdDev  float64 // expected standard deviation of the noise
		checkBounds     bool    // whether to check the bounds of the noise
		lowerBound      float64 // lower bound of the noise
		upperBound      float64 // upper bound of the noise
		checkStepSize   bool    // whether to check the change in noise from one sample to the next
		maxStepSize     float64 // maximum step size allowed
	}

	testCases := []TestCase{
		{
			name:            "random_noise",
			numSamples:      1e6,
			checkStatistics: true,
			expectedMean:    0,
			expectedStdDev:  A / math.Sqrt(3),
			checkBounds:     true,
			lowerBound:      -A,
			upperBound:      A,
			checkStepSize:   false,
		},
		{
			name:            "gaussian_noise",
			numSamples:      1e6,
			checkStatistics: true,
			expectedMean:    0,
			expectedStdDev:  A,
			checkBounds:     false,
			checkStepSize:   false,
		},
		{
			name:            "exponential_noise",
			numSamples:      1e6,
			checkStatistics: true,
			expectedMean:    A,
			expectedStdDev:  A,
			checkBounds:     true,
			lowerBound:      0,
			upperBound:      math.Inf(1),
			checkStepSize:   false,
		},
		{
			name:            "random_walk",
			numSamples:      100, // statistics not being checked so fewer samples required
			checkStatistics: false,
			checkBounds:     true,
			lowerBound:      -A,
			upperBound:      A,
			checkStepSize:   true,
			maxStepSize:     A / 20.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFunction, err := mathfuncs.GetTrendFunctionFromName(tc.name)
			assert.NoError(t, err)

			var sum, sumSq float64
			var prevValue float64
			for i := 0; i < tc.numSamples; i++ {
				x := testFunction(float64(i), A, 0)
				if tc.checkBounds {
					assert.True(t, x >= tc.lowerBound && x <= tc.upperBound, "value out of bounds")
				}
				if tc.checkStepSize {
					assert.True(t, math.Abs(x-prevValue) <= tc.maxStepSize, "step size larger than max step size")
				}
				sum += x
				sumSq += x * x
				prevValue = x
			}

			if tc.checkStatistics {
				mean := sum / float64(tc.numSamples)
				variance := sumSq/float64(tc.numSamples) - mean*mean
				stddev := math.Sqrt(variance)
				// Low value of 0.01 used for the delta: non-exact values due to small sample sizes
				assert.InDelta(t, tc.expectedMean, mean, 0.01)
				assert.InDelta(t, tc.expectedStdDev, stddev, 0.01)
			}
		})
	}
}
