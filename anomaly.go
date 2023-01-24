package emulator

import "math/rand"

type Anomaly struct {
	// instantaneous anomalies, based on probability factor
	isInstantaneousAnomaly          bool // private, activated based on probability
	InstantaneousAnomalyProbability float64
	InstantaneousAnomalyMagnitude   float64

	// trend anomalies, providing positive or negative slope of given magnitude and duration
	IsTrendAnomaly        bool
	IsRisingTrendAnomaly  bool
	TrendAnomalyDuration  float64 // duration in seconds
	TrendStartDelay       float64 // duration in seconds
	TrendStartIndex       int     // number of time step ticks
	TrendAnomalyIndex     int     // number of time step ticks
	TrendAnomalyMagnitude float64
}

func (anomaly *Anomaly) stepAnomaly(r *rand.Rand, Ts float64) float64 {
	trendAnomalyDelta := 0.0
	trendAnomalyStep := (anomaly.TrendAnomalyMagnitude / anomaly.TrendAnomalyDuration) * Ts

	if anomaly.IsTrendAnomaly {
		if anomaly.TrendStartIndex >= int(anomaly.TrendStartDelay/Ts)-1 {
			if anomaly.IsRisingTrendAnomaly {
				trendAnomalyDelta = float64(anomaly.TrendAnomalyIndex) * trendAnomalyStep
			} else {
				trendAnomalyDelta = float64(anomaly.TrendAnomalyIndex) * trendAnomalyStep * (-1.0)
			}

			if anomaly.TrendAnomalyIndex == int(anomaly.TrendAnomalyDuration/Ts)-1 {
				anomaly.TrendAnomalyIndex = 0
				anomaly.TrendStartIndex = 0
			} else {
				anomaly.TrendAnomalyIndex += 1
			}
		} else {
			anomaly.TrendStartIndex += 1
		}
	}

	instantaneousAnomalyDelta := 0.0
	anomaly.isInstantaneousAnomaly = false
	if anomaly.InstantaneousAnomalyProbability > r.Float64() {
		instantaneousAnomalyDelta = anomaly.InstantaneousAnomalyMagnitude
		anomaly.isInstantaneousAnomaly = true
	}

	totalAnomalyDelta := trendAnomalyDelta + instantaneousAnomalyDelta

	return totalAnomalyDelta
}
