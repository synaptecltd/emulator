package anomaly

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
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
