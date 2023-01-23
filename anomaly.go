package emulator

import "math/rand"

type Anomaly struct {
	// instantaneous anomalies
	isInstantaneousAnomaly          bool // private
	InstantaneousAnomalyProbability float64
	InstantaneousAnomalyMagnitude   float64

	// trend anomalies
	IsTrendAnomaly        bool
	IsRisingTrendAnomaly  bool
	TrendAnomalyDuration  int // duration in seconds
	TrendAnomalyIndex     int
	TrendAnomalyMagnitude float64
}

func (anomaly *Anomaly) stepAnomaly(r *rand.Rand, Ts float64) float64 {
	trendAnomalyDelta := 0.0
	trendAnomalyStep := (anomaly.TrendAnomalyMagnitude / float64(anomaly.TrendAnomalyDuration)) * Ts

	if anomaly.IsTrendAnomaly {
		if anomaly.IsRisingTrendAnomaly {
			trendAnomalyDelta = float64(anomaly.TrendAnomalyIndex) * trendAnomalyStep
		} else {
			trendAnomalyDelta = float64(anomaly.TrendAnomalyIndex) * trendAnomalyStep * (-1.0)
		}

		if anomaly.TrendAnomalyIndex == int(float64(anomaly.TrendAnomalyDuration)/Ts)-1 {
			anomaly.TrendAnomalyIndex = 0
		} else {
			anomaly.TrendAnomalyIndex += 1
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
