package emulatoranomaly

import "github.com/synaptecltd/emulator/mathfuncs"

// TrendAnomaly modulates data using repeated continuous functions.
type TrendAnomaly struct {
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

// Converts TrendFuncName to a mathematical function and stores the function internally.
func (t *TrendAnomaly) initTrendFunction() {
	trendFunc, err := mathfuncs.GetTrendFunctionFromName(t.TrendFuncName)
	if err != nil {
		panic(err)
	}
	t.trendFunction = trendFunc
}

// Returns the change in signal caused by the trend anomaly this timestep.
// Manages internal indices to track the progress of trend cycles, and delays between trend cycles.
// Ts is the sampling period of the data.
func (t *TrendAnomaly) stepTrendDelta(Ts float64) float64 {
	if !t.isTrendsAnomalyValid() {
		return 0.0
	}
	// Check if the trend anomaly is active this timestep
	t.TrendAnomalyActive = t.isTrendsAnomalyActive(Ts)
	if !t.isTrendsAnomalyActive(Ts) {
		t.TrendStartIndex += 1 // only increment if inactive to keep track of the delay between trend cycles
		return 0.0
	}

	// How long this trend cycle has been active in seconds
	elapsedTrendTime := float64(t.TrendAnomalyIndex) * Ts

	trendAnomalyMagnitude := t.trendFunction(elapsedTrendTime, t.TrendAnomalyMagnitude, t.TrendAnomalyDuration)
	trendAnomalyDelta := t.getTrendAnomalySign() * trendAnomalyMagnitude
	t.TrendAnomalyIndex += 1

	// If the trend anomaly is complete, reset the index and increment the repeat counter
	if t.TrendAnomalyIndex == int(t.TrendAnomalyDuration/Ts) {
		t.TrendAnomalyIndex = 0
		t.TrendStartIndex = 0
		t.trendRepeats += 1
	}

	return trendAnomalyDelta
}

// Returns whether trend anomalies should be active this timestep. This is true if:
//  1. Enough time has elapsed for the trend anomaly to start, and;
//  2. The trend anomaly has not yet completed all repetitions.
func (t *TrendAnomaly) isTrendsAnomalyActive(Ts float64) bool {
	moreRepeatsAllowed := t.trendRepeats < t.TrendRepetition || t.TrendRepetition == 0 // 0 means infinite repetitions
	if !moreRepeatsAllowed {
		return false
	}

	hasTrendStarted := t.TrendStartIndex >= int(t.TrendStartDelay/Ts)-1
	return hasTrendStarted
}

// Returns whether a trend anomaly is valid based on the following criteria:
//  1. IsTrendAnomaly == true, and;
//  2. TrendAnomalyDuration != 0;
func (t *TrendAnomaly) isTrendsAnomalyValid() bool {
	hasNonZeroDuration := t.TrendAnomalyDuration != 0.0
	return t.IsTrendAnomaly && hasNonZeroDuration
}

// Returns +1.0 if RisingTrendAnomaly is true, or -1.0 if false.
func (t *TrendAnomaly) getTrendAnomalySign() float64 {
	if t.IsRisingTrendAnomaly {
		return 1.0
	}
	return -1.0
}
