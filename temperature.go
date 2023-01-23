package emulator

import (
	"math"
	"math/rand"
)

type TemperatureEmulation struct {
	MeanTemperature float64
	NoiseMax        float64
	ModulationMag   float64

	// instantaneous anomalies
	isInstantaneousAnomaly          bool // private
	InstantaneousAnomalyProbability float64
	InstantaneousAnomalyMagnitude   float64

	// trend anomalies
	IsTrendAnomaly        bool
	TrendAnomalyDuration  int // duration in seconds
	TrendAnomalyIndex     int
	TrendAnomalyMagnitude float64

	IsRisingTrendAnomaly bool

	T float64
}

func (t *TemperatureEmulation) stepTemperature(r *rand.Rand, Ts float64) {
	varyingT := t.MeanTemperature * (1 + t.ModulationMag*math.Cos(1000.0*Ts))

	trendAnomalyDelta := 0.0
	trendAnomalyStep := (t.TrendAnomalyMagnitude / float64(t.TrendAnomalyDuration)) * Ts

	if t.IsTrendAnomaly == true {

		if t.IsRisingTrendAnomaly == true {
			trendAnomalyDelta = float64(t.TrendAnomalyIndex) * trendAnomalyStep
		} else {
			trendAnomalyDelta = float64(t.TrendAnomalyIndex) * trendAnomalyStep * (-1.0)
		}

		if t.TrendAnomalyIndex == int(float64(t.TrendAnomalyDuration)/Ts)-1 {
			t.TrendAnomalyIndex = 0
		} else {
			t.TrendAnomalyIndex += 1
		}
	}

	instantaneousAnomalyDelta := 0.0
	t.isInstantaneousAnomaly = false
	if t.InstantaneousAnomalyProbability > rand.Float64() {
		instantaneousAnomalyDelta = t.InstantaneousAnomalyMagnitude
		t.isInstantaneousAnomaly = true
	}

	totalAnomalyDelta := trendAnomalyDelta + instantaneousAnomalyDelta

	t.T = varyingT + r.NormFloat64()*t.NoiseMax*t.MeanTemperature + totalAnomalyDelta
}
