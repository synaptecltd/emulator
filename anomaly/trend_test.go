package anomaly

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/synaptecltd/emulator/mathfuncs"
)

func TestReverseInvertTrend(t *testing.T) {
	type testcase struct {
		InvertTrend  bool
		ReverseTrend bool
		Expected     []float64
	}

	testcases := []testcase{
		{InvertTrend: false, ReverseTrend: false, Expected: []float64{0.0, 2.0, 4.0, 6.0, 8.0}},
		{InvertTrend: true, ReverseTrend: false, Expected: []float64{0.0, -2.0, -4.0, -6.0, -8.0}},
		{InvertTrend: false, ReverseTrend: true, Expected: []float64{10.0, 8.0, 6.0, 4.0, 2.0}},
		{InvertTrend: true, ReverseTrend: true, Expected: []float64{-10.0, -8.0, -6.0, -4.0, -2.0}},
	}

	Ts := 1.0 // 1 second time step
	rng := rand.New(rand.NewPCG(42, 0))

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("InvertTrend:%v-ReverseTrend:%v", tc.InvertTrend, tc.ReverseTrend), func(t *testing.T) {
			params := TrendParams{
				Name:         "ReverseInvertTrendTest",
				StartDelay:   0.0,
				Duration:     5.0,
				Magnitude:    10.0,
				InvertTrend:  tc.InvertTrend,
				ReverseTrend: tc.ReverseTrend,
				Off:          false,
			}

			trendAnomaly, err := NewTrendAnomaly(params)
			assert.NoError(t, err)

			for i, expected := range tc.Expected {
				value := trendAnomaly.stepAnomaly(rng, Ts)
				assert.InDelta(t, expected, value, 1e-6, "At step %d with InvertTrend=%v ReverseTrend=%v", i, tc.InvertTrend, tc.ReverseTrend)
			}
			for i, expected := range tc.Expected {
				value := trendAnomaly.stepAnomaly(rng, Ts)
				assert.InDelta(t, expected, value, 1e-6, "Post-duration step %d with InvertTrend=%v ReverseTrend=%v", i, tc.InvertTrend, tc.ReverseTrend)
			}
		})
	}
}

func TestNewTrendAnomaly(t *testing.T) {
	t.Run("ValidParams", func(t *testing.T) {
		params := TrendParams{
			Name:         "test_trend",
			Repeats:      5,
			Off:          false,
			StartDelay:   1.0,
			Duration:     10.0,
			Magnitude:    2.5,
			MagFuncName:  "linear",
			InvertTrend:  false,
			ReverseTrend: false,
		}

		trend, err := NewTrendAnomaly(params)
		require.NoError(t, err)
		assert.Equal(t, "test_trend", trend.name)
		assert.Equal(t, "trend", trend.typeName)
		assert.Equal(t, uint64(5), trend.Repeats)
		assert.Equal(t, false, trend.Off)
		assert.Equal(t, 1.0, trend.startDelay)
		assert.Equal(t, 10.0, trend.duration)
		assert.Equal(t, 2.5, trend.Magnitude)
		assert.Equal(t, "linear", trend.magFuncName)
		assert.Equal(t, false, trend.InvertTrend)
		assert.Equal(t, false, trend.ReverseTrend)
		assert.NotNil(t, trend.magFunction)
	})

	t.Run("DefaultMagFunc", func(t *testing.T) {
		params := TrendParams{
			Name:        "test_trend",
			Duration:    5.0,
			MagFuncName: "", // empty should default to linear
		}

		trend, err := NewTrendAnomaly(params)
		require.NoError(t, err)
		assert.Equal(t, "linear", trend.magFuncName)
	})

	t.Run("InvalidDuration", func(t *testing.T) {
		params := TrendParams{
			Name:     "test_trend",
			Duration: -1.0,
		}

		_, err := NewTrendAnomaly(params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duration must be positive value")
	})

	t.Run("InvalidStartDelay", func(t *testing.T) {
		params := TrendParams{
			Name:       "test_trend",
			Duration:   5.0,
			StartDelay: -1.0,
		}

		_, err := NewTrendAnomaly(params)
		assert.Error(t, err)
	})

	t.Run("InvalidMagFuncName", func(t *testing.T) {
		params := TrendParams{
			Name:        "test_trend",
			Duration:    5.0,
			MagFuncName: "invalid_function",
		}

		_, err := NewTrendAnomaly(params)
		assert.Error(t, err)
	})

	t.Run("ZeroDurationTurnsOff", func(t *testing.T) {
		params := TrendParams{
			Name:     "test_trend",
			Duration: 0.0,
			Off:      false, // This will be overridden by SetDuration when duration is 0
		}

		trend, err := NewTrendAnomaly(params)
		require.NoError(t, err)
		// SetDuration(0.0) should set Off = true regardless of the initial Off value
		assert.True(t, trend.Off)
	})
}

func TestTrendAnomalyUnmarshalYAML(t *testing.T) {
	t.Run("ValidYAML", func(t *testing.T) {
		yamlData := `
Name: "test_trend"
Repeats: 3
Off: false
StartDelay: 2.0
Duration: 8.0
Magnitude: 1.5
MagFunc: "sine"
Invert: true
Reverse: false
`
		var trend trendAnomaly
		err := yaml.Unmarshal([]byte(yamlData), &trend)
		require.NoError(t, err)
		assert.Equal(t, "test_trend", trend.name)
		assert.Equal(t, uint64(3), trend.Repeats)
		assert.Equal(t, 1.5, trend.Magnitude)
		assert.Equal(t, "sine", trend.magFuncName)
		assert.True(t, trend.InvertTrend)
		assert.False(t, trend.ReverseTrend)
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		yamlData := `
Name: "test_trend"
Duration: -5.0
`
		var trend trendAnomaly
		err := yaml.Unmarshal([]byte(yamlData), &trend)
		assert.Error(t, err)
	})
}

func TestTrendAnomalyStepAnomaly(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 0))
	Ts := 0.1

	t.Run("OffState", func(t *testing.T) {
		params := TrendParams{
			Name:     "test_trend",
			Duration: 5.0,
			Off:      true,
		}
		trend, err := NewTrendAnomaly(params)
		require.NoError(t, err)

		result := trend.stepAnomaly(rng, Ts)
		assert.Equal(t, 0.0, result)
	})

	t.Run("LinearTrend", func(t *testing.T) {
		params := TrendParams{
			Name:        "test_trend",
			Duration:    1.0,
			StartDelay:  0.0, // No delay to ensure immediate activation
			Magnitude:   2.0,
			MagFuncName: "linear",
		}
		trend, err := NewTrendAnomaly(params)
		require.NoError(t, err)

		// First step should be active
		var result1 float64
		for range 2 {
			result1 = trend.stepAnomaly(rng, Ts)
		}
		assert.Greater(t, result1, 0.0)

		// Continue stepping through the duration
		for i := 2; i < int(1.0/Ts); i++ {
			result := trend.stepAnomaly(rng, Ts)
			assert.Greater(t, result, 0.0)
		}

		// After duration completes, should reset
		assert.Equal(t, 0, trend.elapsedActivatedIndex)
		assert.Equal(t, uint64(1), trend.countRepeats)
	})

	t.Run("WithStartDelay", func(t *testing.T) {
		params := TrendParams{
			Name:       "test_trend",
			Duration:   1.0,
			StartDelay: 0.5,
			Magnitude:  2.0,
		}
		trend, err := NewTrendAnomaly(params)
		require.NoError(t, err)

		// First few steps should return 0 due to start delay
		for i := 0; i < int(0.5/Ts); i++ {
			result := trend.stepAnomaly(rng, Ts)
			assert.Equal(t, 0.0, result)
		}

		// After delay, should start producing values
		result := trend.stepAnomaly(rng, Ts)
		assert.Greater(t, result, 0.0)
	})
}

func TestTrendAnomalyAllMagFunctions(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 0))
	Ts := 0.1
	for _, funcName := range mathfuncs.GetMathsFunctionNames() {
		t.Run(funcName, func(t *testing.T) {
			params := TrendParams{
				Name:        "test_trend",
				Duration:    1.0,
				StartDelay:  0.0,
				Magnitude:   1.0,
				MagFuncName: funcName,
			}
			trend, err := NewTrendAnomaly(params)
			require.NoError(t, err)

			// Step through the trend anomaly duration
			for i := 0; i < int(1.0/Ts); i++ {
				result := trend.stepAnomaly(rng, Ts)
				// Just ensure we get a float64 result without error
				assert.IsType(t, float64(0.0), result)
				assert.False(t, math.IsNaN(result))
				assert.False(t, math.IsInf(result, 0))
			}
		})
	}
}

func TestTrendAnomalySetDuration(t *testing.T) {
	trend := &trendAnomaly{}

	t.Run("ValidDuration", func(t *testing.T) {
		err := trend.SetDuration(5.0)
		assert.NoError(t, err)
		assert.Equal(t, 5.0, trend.duration)
		assert.False(t, trend.Off)
	})

	t.Run("ZeroDuration", func(t *testing.T) {
		trend.Off = false // Reset
		err := trend.SetDuration(0.0)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, trend.duration)
		assert.True(t, trend.Off)
	})

	t.Run("NegativeDuration", func(t *testing.T) {
		err := trend.SetDuration(-1.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duration must be positive")
	})
}

func TestTrendAnomalySetMagFunctionByName(t *testing.T) {
	trend := &trendAnomaly{}

	for _, funcName := range mathfuncs.GetMathsFunctionNames() {
		t.Run(funcName, func(t *testing.T) {
			err := trend.SetMagFunctionByName(funcName)
			assert.NoError(t, err)
			assert.Equal(t, funcName, trend.magFuncName)
			assert.NotNil(t, trend.magFunction)
		})
	}

	t.Run("EmptyName", func(t *testing.T) {
		err := trend.SetMagFunctionByName("")
		assert.NoError(t, err)
		assert.Equal(t, "linear", trend.magFuncName) // Should default to linear
	})

	t.Run("InvalidFunction", func(t *testing.T) {
		err := trend.SetMagFunctionByName("invalid_function")
		assert.Error(t, err)
	})
}

func TestTrendAnomalyGetters(t *testing.T) {
	params := TrendParams{
		Name:        "test_trend",
		Duration:    5.0,
		MagFuncName: "cosine",
	}
	trend, err := NewTrendAnomaly(params)
	require.NoError(t, err)

	t.Run("GetMagFuncName", func(t *testing.T) {
		assert.Equal(t, "cosine", trend.GetMagFuncName())
	})

	t.Run("GetMagFunction", func(t *testing.T) {
		magFunc := trend.GetMagFunction()
		assert.NotNil(t, magFunc)
		// Test that the function works
		result := magFunc(0.0, 1.0, 1.0)
		assert.InDelta(t, 1.0, result, 1e-6) // cos(0) = 1
	})
}

func TestTrendAnomalyIntegration(t *testing.T) {
	t.Run("CompleteLifecycle", func(t *testing.T) {
		params := TrendParams{
			Name:        "integration_test",
			Repeats:     2,
			StartDelay:  0.2,
			Duration:    0.5,
			Magnitude:   1.0,
			MagFuncName: "linear",
		}
		trend, err := NewTrendAnomaly(params)
		require.NoError(t, err)

		rng := rand.New(rand.NewPCG(42, 0))
		Ts := 0.1
		results := []float64{}

		// Run for enough steps to complete 2 repeats
		totalSteps := int((0.2+0.5)*2/Ts) + 10
		for range totalSteps {
			result := trend.stepAnomaly(rng, Ts)
			results = append(results, result)
		}

		// Should have completed 2 repeats
		assert.Equal(t, uint64(2), trend.countRepeats)

		// Check that we have periods of zero (delays) and non-zero (active) values
		hasZeros := false
		hasNonZeros := false
		for _, result := range results {
			if result == 0.0 {
				hasZeros = true
			} else {
				hasNonZeros = true
			}
		}
		assert.True(t, hasZeros, "Should have zero values during delays")
		assert.True(t, hasNonZeros, "Should have non-zero values during active periods")
	})
}

func BenchmarkTrendAnomalyStepAnomaly(b *testing.B) {
	params := TrendParams{
		Name:        "benchmark_trend",
		Duration:    10.0,
		Magnitude:   2.0,
		MagFuncName: "linear",
	}
	trend, err := NewTrendAnomaly(params)
	require.NoError(b, err)

	rng := rand.New(rand.NewPCG(42, 0))
	Ts := 0.01

	b.ResetTimer()
	for b.Loop() {
		trend.stepAnomaly(rng, Ts)
	}
}
