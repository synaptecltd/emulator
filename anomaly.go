package emulator

import (
	"math/rand/v2"

	"github.com/synaptecltd/emulator/mathfuncs"
)

// Anomaly provides combinations of instantaneous and trend anomalies.
//   - InstantaneousAnomaly produces spikes in the data that occur at each timestep based on a probability factor.
//   - TrendAnomaly provides periodic positive or negative slopes of given magnitude and duration
type Anomaly struct {
	// Instantaneous anomalies
	InstantaneousAnomalyProbability        float64 `yaml:"InstantaneousAnomalyProbability"`        // probability of instantaneous anomaly in each time step
	InstantaneousAnomalyMagnitude          float64 `yaml:"InstantaneousAnomalyMagnitude"`          // magnitude of instantaneous anomalies
	InstantaneousAnomalyMagnitudeVariation bool    `yaml:"InstantaneousAnomalyMagnitudeVariation"` // whether to vary the magnitude of the instantaneous anomaly, default false
	InstantaneousAnomalyActive             bool    // indicates whether instantaneous anomaly is active in this time step

	// Trend anomalies
	IsTrendAnomaly        bool    `yaml:"IsTrendAnomaly"`        // true: trend anomalies activated, false: deactivated
	IsRisingTrendAnomaly  bool    `yaml:"IsRisingTrendAnomaly"`  // true: positive slope, false: negative slope
	TrendAnomalyDuration  float64 `yaml:"TrendAnomalyDuration"`  // duration of each trend anomaly in seconds
	TrendStartDelay       float64 `yaml:"TrendStartDelay"`       // start time of trend anomalies in seconds
	TrendAnomalyMagnitude float64 `yaml:"TrendAnomalyMagnitude"` // magnitude of trend anomaly
	TrendRepetition       int     `yaml:"TrendRepetition"`       // number of times the trend anomaly repeats, default 0 for infinite
	TrendFuncName         string  `yaml:"TrendFunction"`         // name of function to use for the trend anomaly, defaults to linear ramp if empty
	TrendAnomalyActive    bool    // indicates whether trend anomaly is active in this time step

	TrendStartIndex   int `yaml:"TrendStartIndex"`   // TrendStartDelay converted to number of time steps
	TrendAnomalyIndex int `yaml:"TrendAnomalyIndex"` // number of time steps since the start of the last trend anomaly

	// internal state
	trendRepeats int                                     // counter for number of times trend anomaly has repeated
	trendFunc    func(float64, float64, float64) float64 // function to use for the trend anomaly
}

// A collection of named anomalies.
type AnomalyContainer map[string]*Anomaly

// Steps all anomalies and returns the sum of their effects.
func stepAllAnomalies(anomalies AnomalyContainer, r *rand.Rand, Ts float64) float64 {
	value := 0.0
	// Must use indexing so that each anomaly internal state is updated
	for key := range anomalies {
		value += anomalies[key].stepAnomaly(r, Ts)
	}
	return value
}

// Returns the change in signal from instantaneous and trend anomalies this timestep.
// Ts is the sampling period of the data in seconds, and r is a random number generator.
func (a *Anomaly) stepAnomaly(r *rand.Rand, Ts float64) float64 {
	// Set internal state: get the function to use for the trend anomaly if not set
	if a.trendFunc == nil {
		a.initTrendFunc()
	}

	instantaneousAnomalyDelta := a.getInstantaneousDelta(r)
	trendAnomalyDelta := a.getTrendDelta(Ts)

	return instantaneousAnomalyDelta + trendAnomalyDelta
}

// Initializes the trend function to use for the trend anomaly.
func (a *Anomaly) initTrendFunc() {
	trendFunc, err := mathfuncs.GetTrendFunctionFromName(a.TrendFuncName)
	if err != nil {
		panic(err)
	}
	a.trendFunc = trendFunc
}

// Returns the change in signal caused by the instantaneous anomaly this timestep.
func (a *Anomaly) getInstantaneousDelta(r *rand.Rand) float64 {
	// No anomaly if probability is not met
	if r.Float64() > a.InstantaneousAnomalyProbability {
		a.InstantaneousAnomalyActive = false
		return 0.0
	}

	a.InstantaneousAnomalyActive = true
	if a.InstantaneousAnomalyMagnitudeVariation {
		return a.InstantaneousAnomalyMagnitude * r.NormFloat64()
	}
	return a.InstantaneousAnomalyMagnitude
}

// Returns the change in signal caused by the trend anomaly this timestep (the delta),
// and increments trendAnomalyIndex to track the progress of the trend internally.
// Ts is the sampling period of the data.
func (a *Anomaly) getTrendDelta(Ts float64) float64 {
	// Check if the trend anomaly should be active
	a.TrendAnomalyActive = a.isTrendsAnomalyActive(Ts)
	if !a.isTrendsAnomalyActive(Ts) {
		return 0.0
	}

	// Current duration into the trend anomaly in seconds
	elapsedTrendTime := float64(a.TrendAnomalyIndex) * Ts

	trendAnomalyMagnitude := a.trendFunc(elapsedTrendTime, a.TrendAnomalyMagnitude, a.TrendAnomalyDuration)
	trendAnomalyDelta := a.getTrendAnomalySign() * trendAnomalyMagnitude
	a.TrendAnomalyIndex += 1

	// If the trend anomaly is complete, reset the index and increment the repeat counter
	if a.TrendAnomalyIndex == int(a.TrendAnomalyDuration/Ts) {
		a.TrendAnomalyIndex = 0
		a.TrendStartIndex = 0
		a.trendRepeats += 1
	}

	return trendAnomalyDelta
}

// Returns whether trend anomalies should be active this timestep based on meeting
// all of the following criteria:
//  1. IsTrendAnomaly == true;
//  2. TrendAnomalyDuration != 0;
//  3. The number of repetitions of the trend has not exceeded TrendRepetition, and;
//  4. TrendStartDelay has elapsed.
//
// If the anomaly has not yet started, then TrendStartIndex is incremented.
func (a *Anomaly) isTrendsAnomalyActive(Ts float64) bool {
	// Start with validity checks
	isActivated := a.IsTrendAnomaly
	hasNonZeroDuration := a.TrendAnomalyDuration != 0.0
	moreRepeatsAllowed := a.trendRepeats < a.TrendRepetition || a.TrendRepetition == 0

	isValid := isActivated && hasNonZeroDuration && moreRepeatsAllowed

	if !isValid {
		return false
	}

	// Increment TrendStartIndex if the anomaly has not yet started
	indexWithinScope := a.TrendStartIndex >= int(a.TrendStartDelay/Ts)-1
	if !indexWithinScope {
		a.TrendStartIndex += 1
		return false
	}

	return true
}

// Returns +1.0 if RisingTrendAnomaly is true, or -1.0 if false.
func (a *Anomaly) getTrendAnomalySign() float64 {
	if a.IsRisingTrendAnomaly {
		return 1.0
	}
	return -1.0
}
