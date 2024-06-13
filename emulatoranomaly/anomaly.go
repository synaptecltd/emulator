package emulatoranomaly

import (
	"math/rand/v2"
)

// Anomaly provides combinations of instantaneous and trend anomalies.
type Anomaly struct {
	InstantaneousAnomaly
	TrendAnomaly
}

// A collection of named anomalies.
type Container map[string]*Anomaly

// Steps all anomalies and returns the sum of their effects.
func (anomalies Container) StepAll(r *rand.Rand, Ts float64) float64 {
	value := 0.0
	for key := range anomalies {
		value += anomalies[key].stepAnomaly(r, Ts)
	}
	return value
}

// Returns the change in signal from instantaneous and trend anomalies this timestep.
// Ts is the sampling period of the data in seconds, and r is a random number generator.
func (a *Anomaly) stepAnomaly(r *rand.Rand, Ts float64) float64 {
	// Set internal state: get the function to use for the trend anomaly if not set
	if a.trendFunction == nil {
		a.initTrendFunction()
	}

	instantaneousAnomalyDelta := a.getInstantaneousDelta(r)
	trendAnomalyDelta := a.stepTrendDelta(Ts)

	return instantaneousAnomalyDelta + trendAnomalyDelta
}
