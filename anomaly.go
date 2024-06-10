package emulator

import "math/rand/v2"

type Anomaly struct {
	InstantaneousAnomaly
	TrendAnomaly
}

// Calculates the anomaly delta for the current timestep based on all active anomalies. Ts is the sampling
// period of the data in seconds, and r is a random number generator.
func (anomaly *Anomaly) stepAnomaly(r *rand.Rand, Ts float64) float64 {
	instantaneousAnomalyDelta := anomaly.InstantaneousAnomaly.getDelta(r)
	trendAnomalyDelta := anomaly.TrendAnomaly.getDelta(Ts)

	totalAnomalyDelta := trendAnomalyDelta + instantaneousAnomalyDelta

	return totalAnomalyDelta
}

// InstantaneousAnomaly provides a spike in the temperature based on a probability factor.
type InstantaneousAnomaly struct {
	InstantaneousAnomalyProbability float64 `yaml:"InstantaneousAnomalyProbability"` // probability of instantaneous anomaly in each time step
	InstantaneousAnomalyMagnitude   float64 `yaml:"InstantaneousAnomalyMagnitude"`   // magnitude of instantaneous anomalies
	InstantaneousAnomalyActive      bool    // whether an instantaneous anomaly is active in this time step
}

// Returns the instantaneous anomaly delta. The anomaly is activated based
// by comparing a random number with the probability of an anomaly occuring.
func (a *InstantaneousAnomaly) getDelta(r *rand.Rand) float64 {
	instantaneousAnomalyDelta := 0.0
	a.InstantaneousAnomalyActive = false
	if a.InstantaneousAnomalyProbability > r.Float64() {
		instantaneousAnomalyDelta = a.InstantaneousAnomalyMagnitude
		a.InstantaneousAnomalyActive = true
	}
	return instantaneousAnomalyDelta
}

// TrendAnomaly provides periodic positive or negative slopes of given magnitude and duration
type TrendAnomaly struct {
	IsTrendAnomaly        bool    `yaml:"IsTrendAnomaly"`        // true to turn trend anomalies on, false to deactivate
	IsRisingTrendAnomaly  bool    `yaml:"IsRisingTrendAnomaly"`  // true for positive slope, false for negative slope
	TrendAnomalyDuration  float64 `yaml:"TrendAnomalyDuration"`  // duration of trend anomaly in seconds
	TrendStartDelay       float64 `yaml:"TrendStartDelay"`       // start time of trend anomaly in seconds
	TrendStartIndex       int     `yaml:"TrendStartIndex"`       // number of time step ticks
	TrendAnomalyIndex     int     `yaml:"TrendAnomalyIndex"`     // number of time step ticks
	TrendAnomalyMagnitude float64 `yaml:"TrendAnomalyMagnitude"` // magnitude of trend anomaly (the maximum height of the slope)
	TrendRepetition       int     `yaml:"TrendRepetition"`       // number of times the trend anomaly is repeated, default 0 for infinite
	TrendAnomalyActive    bool    // whether a trend anomaly is active in this time step

	trendRepeats int // internal counter for number of times the trend anomaly has been repeated
}

// Returns the trend anomaly delta and increments the internal index of the anomaly to
// track the progress of the anomaly. Ts is the sampling period of the data.
// The duration and progress of the anomaly is tracked internally.
func (t *TrendAnomaly) getDelta(Ts float64) float64 {
	// Check if the trend anomaly should be active
	t.TrendAnomalyActive = t.isTrendsAnomalyActive(Ts)
	if !t.isTrendsAnomalyActive(Ts) {
		return 0.0
	}

	// Slope of the trend anomaly in units of magnitude change per second
	trendSlope := t.TrendAnomalyMagnitude / t.TrendAnomalyDuration

	// The duration that we are through the existing trend anomaly in seconds
	elapsedTrendTime := float64(t.TrendAnomalyIndex) * Ts

	// The sign of the trend anomaly based on whether it is a rising or falling trend
	trendAnomalySign := t.getTrendAnomalySign()

	trendAnomalyDelta := elapsedTrendTime * trendSlope * trendAnomalySign

	t.TrendAnomalyIndex += 1

	// If the trend anomaly is complete, reset the index and increment the repeat counter
	if t.TrendAnomalyIndex == int(t.TrendAnomalyDuration/Ts) {
		t.TrendAnomalyIndex = 0
		t.TrendStartIndex = 0
		t.trendRepeats += 1
	}

	return trendAnomalyDelta
}

// Returns whether the trend anomaly should be active based on it meeting
// all of the following validity criteria:
//  1. IsTrendAnomaly is true;
//  2. TrendAnomalyDuration is not 0;
//  3. The number of repetitions of the trend has not exceeded the limit, and;
//  4. The index is within the scope of the anomaly.
//
// If the anomaly is active and has not yet started, the index is incremented.
func (t *TrendAnomaly) isTrendsAnomalyActive(Ts float64) bool {
	trendsActive := t.IsTrendAnomaly
	hasNonZeroDuration := t.TrendAnomalyDuration != 0.0
	repetitionsWithinScope := t.trendRepeats < t.TrendRepetition || t.TrendRepetition == 0

	validTrends := trendsActive && hasNonZeroDuration && repetitionsWithinScope

	if !validTrends {
		return false
	}

	// Increment the index if the anomaly has not yet started
	indexWithinScope := t.TrendStartIndex >= int(t.TrendStartDelay/Ts)-1
	if !indexWithinScope {
		t.TrendStartIndex += 1
		return false
	}

	return true
}

// Returns the sign of the trend anomaly based on whether it is a rising or falling trend.
func (t *TrendAnomaly) getTrendAnomalySign() float64 {
	if t.IsRisingTrendAnomaly {
		return 1.0
	}
	return -1.0
}
