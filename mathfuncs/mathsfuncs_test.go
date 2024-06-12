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
	M := 1.0 + rand.Float64()*99.0 // ampltiude (between 1 and 100)
	x := 1.0 + rand.Float64()*99.0 // time (between 1 and 100)

	testCases := []struct {
		name     string  // name of the function, defined in the TrendFunctions map
		t        float64 // time in seconds
		A        float64 // amplitude
		T        float64 // period of the trend in seconds
		expected float64 // expected value of the function at time t
		isError  bool    // true if an error is expected
	}{
		{
			name:    "not_a_function",
			isError: true,
		},
		{
			name:     "linear",
			t:        x,
			A:        M,
			T:        M,
			expected: x, // y = (A/A)*x = x
			isError:  false,
		},
		{
			name:     "sine",
			t:        x,
			A:        M,
			T:        4 * x,
			expected: M, // M*sin(2*pi*(x/4x)) = M*sin(pi/2) = M
			isError:  false,
		},
		{
			name:     "cosine",
			t:        x,
			A:        M,
			T:        4 * x,
			expected: 0.0, // M*cos(pi/2) = 0
			isError:  false,
		},
		{
			name:     "exponential",
			t:        x,
			A:        M,
			T:        x,
			expected: M*math.Exp(1) - M, // because M*exp(t/T) = M*exp(1)
			isError:  false,
		},
		{
			name:     "parabolic",
			t:        x,
			A:        M,
			T:        2 * x,
			expected: M / 4, // M*(x/2x)^2 = M*(1/2)^2 = M/4
			isError:  false,
		},
		{
			name:     "step",
			t:        1.5 * x,
			A:        M,
			T:        2.0 * x,
			expected: M, // positive value for t > T/2
			isError:  false,
		},
		{
			name:     "step",
			t:        0.0,
			A:        M,
			T:        x,
			expected: 0.0, // zero value for t < T/2
			isError:  false,
		},
		{
			name:     "square",
			t:        0.0,
			A:        M,
			T:        x,
			expected: M, // positive value for t=0
			isError:  false,
		},
		{
			name:     "square",
			t:        1.5 * x,
			A:        M,
			T:        2.0 * x,
			expected: -M, // negative value for t > T/2
			isError:  false,
		},
		{
			name:     "sawtooth",
			t:        3.0 * x,
			A:        M,
			T:        x,
			expected: 0.0, // odd numbered factors of time period = middle of sawtooth wave
			isError:  false,
		},
		{
			name:     "sawtooth",
			t:        x,
			A:        M,
			T:        4 * x,
			expected: M / 2, // quarter of time period = half way up the sawtooth wave
			isError:  false,
		},
		{
			name:     "impulse",
			t:        x / 2.0,
			A:        M,
			T:        x,
			expected: 0.0, // no impulse when t != T
			isError:  false,
		},
		{
			name:     "impulse",
			t:        x,
			A:        M,
			T:        x,
			expected: M, // impulse at t==T
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
		checkNoiseDelta bool    // whether to check the change in noise from one sample to the next
		maxDelta        float64 // maximum change in noise allowed between samples
	}

	testCases := []TestCase{
		{
			name:            "random_noise",
			numSamples:      1e6,
			checkStatistics: true,
			expectedMean:    0,                // by definition of uniform distribution
			expectedStdDev:  A / math.Sqrt(3), // by definition of uniform distribution
			checkBounds:     true,
			lowerBound:      -A,
			upperBound:      A,
			checkNoiseDelta: false,
		},
		{
			name:            "gaussian_noise",
			numSamples:      1e6,
			checkStatistics: true,
			expectedMean:    0, // by definition of normal distribution
			expectedStdDev:  A, // by definition of normal distribution
			checkBounds:     false,
			checkNoiseDelta: false,
		},
		{
			name:            "exponential_noise",
			numSamples:      1e6,
			checkStatistics: true,
			expectedMean:    A, // by definition of exponential distribution
			expectedStdDev:  A, // by definition of exponential distribution
			checkBounds:     true,
			lowerBound:      0,           // exponential distribution always non-negative values
			upperBound:      math.Inf(1), // exponential distribution is unbounded at the upper end
			checkNoiseDelta: false,
		},
		{
			name:            "random_walk",
			numSamples:      100, // statistics not being checked so fewer samples required
			checkStatistics: false,
			checkBounds:     true,
			lowerBound:      -A, // bounds are defined within mathfuncs.randomWalk
			upperBound:      A,
			checkNoiseDelta: true,
			maxDelta:        A / 20.0, // maximum step size is defined within mathfuncs.randomWalk
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
				if tc.checkNoiseDelta {
					assert.True(t, math.Abs(x-prevValue) <= tc.maxDelta, "step size larger than max step size")
				}
				sum += x
				sumSq += x * x
				prevValue = x
			}

			if tc.checkStatistics {
				mean := sum / float64(tc.numSamples)
				variance := sumSq/float64(tc.numSamples) - mean*mean
				stddev := math.Sqrt(variance)
				// Low value of 0.1 used for the delta: non-exact values due to small sample sizes
				assert.InDelta(t, tc.expectedMean, mean, 0.1)
				assert.InDelta(t, tc.expectedStdDev, stddev, 0.1)
			}
		})
	}
}
