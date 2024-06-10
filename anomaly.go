package emulator

import "math/rand/v2"

type Anomaly struct {
	// instantaneous anomalies, based on probability factor
	InstantaneousAnomalyProbability float64 `yaml:"InstantaneousAnomalyProbability"` // probability of an instantaneous anomaly in each time step
	InstantaneousAnomalyMagnitude   float64 `yaml:"InstantaneousAnomalyMagnitude"`   // magnitude of instantaneous anomalies
	InstantaneousAnomalyActive      bool    // whether an instantaneous anomaly is active in this time step

	// trend anomalies, providing periodic positive or negative slopes of given magnitude and duration
	IsTrendAnomaly        bool    `yaml:"IsTrendAnomaly"`        // true to turn trend anomalies on, false to deactivate
	IsRisingTrendAnomaly  bool    `yaml:"IsRisingTrendAnomaly"`  // true for positive slope, false for negative slope
	TrendAnomalyDuration  float64 `yaml:"TrendAnomalyDuration"`  // duration of trend anomaly in seconds
	TrendStartDelay       float64 `yaml:"TrendStartDelay"`       // start time of trend anomaly in seconds
	TrendStartIndex       int     `yaml:"TrendStartIndex"`       // number of time step ticks
	TrendAnomalyIndex     int     `yaml:"TrendAnomalyIndex"`     // number of time step ticks
	TrendAnomalyMagnitude float64 `yaml:"TrendAnomalyMagnitude"` // magnitude of trend anomaly (the maximum height of the slope)
	TrendRepetition       int     `yaml:"TrendRepetition"`       // number of times the trend anomaly is repeated, default 0 for infinite
	TrendAnomalyActive    bool    // whether a trend anomaly is active in this time step

	trendRepeats int // counter for number of times the trend anomaly has been repeated
}

func (anomaly *Anomaly) stepAnomaly(r *rand.Rand, Ts float64) float64 {
	// Define trend anomalies
	trendAnomalyDelta := 0.0
	anomaly.TrendAnomalyActive = false
	if anomaly.IsTrendAnomaly && anomaly.TrendAnomalyDuration > 0.0 && (anomaly.trendRepeats < anomaly.TrendRepetition || anomaly.TrendRepetition == 0) {
		anomaly.TrendAnomalyActive = true
		trendAnomalyStep := (anomaly.TrendAnomalyMagnitude / anomaly.TrendAnomalyDuration) * Ts

		if anomaly.TrendStartIndex >= int(anomaly.TrendStartDelay/Ts)-1 {
			if anomaly.IsRisingTrendAnomaly {
				trendAnomalyDelta = float64(anomaly.TrendAnomalyIndex) * trendAnomalyStep
			} else {
				trendAnomalyDelta = float64(anomaly.TrendAnomalyIndex) * trendAnomalyStep * (-1.0)
			}

			if anomaly.TrendAnomalyIndex == int(anomaly.TrendAnomalyDuration/Ts)-1 {
				anomaly.TrendAnomalyIndex = 0
				anomaly.TrendStartIndex = 0
				anomaly.trendRepeats += 1
			} else {
				anomaly.TrendAnomalyIndex += 1
			}
		} else {
			anomaly.TrendStartIndex += 1
		}
	}

	// Define instananous anomalies
	instantaneousAnomalyDelta := anomaly.instantaneousAnomalyDelta(r)
	totalAnomalyDelta := trendAnomalyDelta + instantaneousAnomalyDelta

	return totalAnomalyDelta
}

// Returns the instantaneous anomaly delta. The anomaly is activated based
// on the probability of an anomaly occuring.
func (anomaly *Anomaly) instantaneousAnomalyDelta(r *rand.Rand) float64 {
	instantaneousAnomalyDelta := 0.0
	anomaly.InstantaneousAnomalyActive = false
	if anomaly.InstantaneousAnomalyProbability > r.Float64() {
		instantaneousAnomalyDelta = anomaly.InstantaneousAnomalyMagnitude
		anomaly.InstantaneousAnomalyActive = true
	}
	return instantaneousAnomalyDelta
}
