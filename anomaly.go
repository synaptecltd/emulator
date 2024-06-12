package emulator

import (
	"math/rand/v2"

	"github.com/synaptecltd/emulator/mathfuncs"
)

// Anomaly provides combinations of instantaneous and trend anomalies.
//   - InstantaneousAnomaly produces spikes in the data that occur at each timestep based on a probability factor.
//   - TrendAnomaly modulates data using repeated continuous functions.
type Anomaly struct {
	// Instantaneous anomalies
	InstantaneousAnomalyProbability        float64 `yaml:"InstantaneousAnomalyProbability"`        // probability of instantaneous anomaly in each time step
	InstantaneousAnomalyMagnitude          float64 `yaml:"InstantaneousAnomalyMagnitude"`          // magnitude of instantaneous anomalies
	InstantaneousAnomalyMagnitudeVariation bool    `yaml:"InstantaneousAnomalyMagnitudeVariation"` // whether to vary the magnitude of instantaneous anomaly spikes, default false
	InstantaneousAnomalyActive             bool    // indicates whether instantaneous anomaly spike is active in this time step

	// Trend anomalies
	IsTrendAnomaly        bool    `yaml:"IsTrendAnomaly"`        // true: trend anomalies activated, false: deactivated
	IsRisingTrendAnomaly  bool    `yaml:"IsRisingTrendAnomaly"`  // true: positive slope, false: negative slope
	TrendAnomalyDuration  float64 `yaml:"TrendAnomalyDuration"`  // duration of each trend anomaly in seconds
	TrendStartDelay       float64 `yaml:"TrendStartDelay"`       // start time of trend anomalies in seconds
	TrendAnomalyMagnitude float64 `yaml:"TrendAnomalyMagnitude"` // magnitude of trend anomaly
	TrendRepetition       int     `yaml:"TrendRepetition"`       // number of times the trend anomaly repeats, default 0 for infinite
	TrendFuncName         string  `yaml:"TrendFunction"`         // name of function to use for the trend, defaults to linear ramp if empty
	TrendAnomalyActive    bool    // whether this trend anomaly is modulating the waveform in this time step

	TrendStartIndex   int `yaml:"TrendStartIndex"`   // TrendStartDelay converted to time steps, used to track delay period between trend repeats
	TrendAnomalyIndex int `yaml:"TrendAnomalyIndex"` // number of time steps since start of active trend anomaly, used to track the progress of trend anomaly

	// internal state
	trendRepeats  int                     // counter for number of times trend anomaly has repeated
	trendFunction mathfuncs.TrendFunction // returns trend anomaly magnitude for a given elapsed time, magntiude and period; set internally from TrendFuncName
}

// A collection of named anomalies.
type AnomalyContainer map[string]*Anomaly

// Steps all anomalies and returns the sum of their effects.
func stepAllAnomalies(anomalies AnomalyContainer, r *rand.Rand, Ts float64) float64 {
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

// Converts TrendFuncName to a mathematical function and stores the function internally.
func (a *Anomaly) initTrendFunction() {
	trendFunc, err := mathfuncs.GetTrendFunctionFromName(a.TrendFuncName)
	if err != nil {
		panic(err)
	}
	a.trendFunction = trendFunc
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

// Returns the change in signal caused by the trend anomaly this timestep.
// Manages internal indices to track the progress of trend cycles, and delays between trend cycles.
// Ts is the sampling period of the data.
func (a *Anomaly) stepTrendDelta(Ts float64) float64 {
	if !a.isTrendsAnomalyValid() {
		return 0.0
	}
	// Check if the trend anomaly is active this timestep
	a.TrendAnomalyActive = a.isTrendsAnomalyActive(Ts)
	if !a.isTrendsAnomalyActive(Ts) {
		a.TrendStartIndex += 1 // only increment if inactive to keep track of the delay between trend cycles
		return 0.0
	}

	// How long this trend cycle has been active in seconds
	elapsedTrendTime := float64(a.TrendAnomalyIndex) * Ts

	trendAnomalyMagnitude := a.trendFunction(elapsedTrendTime, a.TrendAnomalyMagnitude, a.TrendAnomalyDuration)
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

// Returns whether trend anomalies should be active this timestep. This is true if:
//  1. Enough time has elapsed for the trend anomaly to start, and;
//  2. The trend anomaly has not yet completed all repetitions.
func (a *Anomaly) isTrendsAnomalyActive(Ts float64) bool {
	moreRepeatsAllowed := a.trendRepeats < a.TrendRepetition || a.TrendRepetition == 0 // 0 means infinite repetitions
	if !moreRepeatsAllowed {
		return false
	}

	hasTrendStarted := a.TrendStartIndex >= int(a.TrendStartDelay/Ts)-1
	return hasTrendStarted
}

// Returns whether a trend anomaly is valid based on the following criteria:
//  1. IsTrendAnomaly == true, and;
//  2. TrendAnomalyDuration != 0;
func (a *Anomaly) isTrendsAnomalyValid() bool {
	hasNonZeroDuration := a.TrendAnomalyDuration != 0.0
	return a.IsTrendAnomaly && hasNonZeroDuration
}

// Returns +1.0 if RisingTrendAnomaly is true, or -1.0 if false.
func (a *Anomaly) getTrendAnomalySign() float64 {
	if a.IsRisingTrendAnomaly {
		return 1.0
	}
	return -1.0
}
