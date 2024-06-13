package emulatoranomaly

import (
	"errors"
	"math/rand/v2"

	"github.com/synaptecltd/emulator/mathfuncs"
)

// trendAnomaly modulates data using repeated continuous functions.
type trendAnomaly struct {
	StartDelay  float64 `yaml:"start_delay"`    // start time of trend anomalies in seconds
	Duration    float64 `yaml:"duration"`       // duration of each trend anomaly in seconds, >= 0
	Magnitude   float64 `yaml:"magnitude"`      // magnitude of trend anomaly
	Repeats     uint64  `yaml:"repeat"`         // number of times the trend anomaly repeats, default 0 for infinite
	FuncName    string  `yaml:"trend_function"` // name of function to use for the trend, defaults to "linear" if empty
	InvertTrend bool    `yaml:"invert"`         // true inverts the trend function (multiplies by -1.0), default false (no inverting)
	Off         bool    `yaml:"off"`            // true: trend anomaly deactivated, false: activated (default)

	// Internal state
	isAnomalyActive   bool                    // whether the trend anomaly is modulating the waveform in this time step
	startDelayIndex   int                     // startDelay converted to time steps, used to track delay period between trend repeats
	elapsedTrendTime  float64                 // time elapsed since the start of the active trend anomaly
	elapsedTrendIndex int                     // number of time steps since start of active trend anomaly, used to track the progress of trend anomaly
	countRepeats      int                     // counter for number of times trend anomaly has repeated
	trendFunction     mathfuncs.TrendFunction // returns trend anomaly magnitude for a given elapsed time, magntiude and period; set internally from TrendFuncName
}

// Initialise the internal fields of TrendAnomaly when it is unmarshalled from yaml.
func (t *trendAnomaly) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Define a secondary type that doesn't inherit trendAnomaly methods. This prevents recursive calls to UnmarshalYAML.
	type trendAnomalyDuplicate trendAnomaly

	var d trendAnomalyDuplicate
	err := unmarshal(d)
	if err != nil {
		return err
	}

	// Use the duplicate's fields to initialise a new TrendAnomaly
	trendAnomaly, err := NewTrendAnomaly(d.StartDelay, d.Duration, d.Magnitude, d.Repeats, d.FuncName, d.InvertTrend, d.Off)
	if err != nil {
		return err
	}

	// Need checkers for startDelay, duration, funcName

	// Copy fields from newTrend to t
	*t = *trendAnomaly

	return nil
}

// Returns a TrendAnomaly with the specified parameters, checking for invalid values.
func NewTrendAnomaly(startDelay float64, duration float64, magnitude float64, repetition uint64, funcName string, invert bool, off bool) (*trendAnomaly, error) {
	// Public fields can never be invalid so set directly
	trendAnomaly := &trendAnomaly{
		Magnitude:   magnitude,
		Repeats:     repetition,
		InvertTrend: invert,
		Off:         off,
	}

	if err := trendAnomaly.SetDuration(duration); err != nil {
		return nil, err
	}

	if err := trendAnomaly.SetStartDelay(startDelay); err != nil {
		return nil, err
	}

	if err := trendAnomaly.SetFunctionByName(funcName); err != nil {
		return nil, err
	}

	return trendAnomaly, nil
}

// Sets the duration of each trend anomaly in seconds if duration > 0.
// If duration=0, the trend anomaly is deactivated.
func (t *trendAnomaly) SetDuration(duration float64) error {
	if duration < 0 {
		return errors.New("duration must be positive value")
	}
	if duration == 0 {
		t.Off = false
	}
	t.Duration = duration
	return nil
}

// Sets the start time of trend anomalies in seconds if delay >= 0.
func (t *trendAnomaly) SetStartDelay(startDelay float64) error {
	if startDelay < 0 {
		return errors.New("startDelay must be greater than or equal to 0")
	}

	t.StartDelay = startDelay
	return nil
}

// Sets the function of the trend anomaly using the name of the function if the
// name is valid.
func (t *trendAnomaly) SetFunctionByName(name string) error {
	trendFunc, err := mathfuncs.GetTrendFunctionFromName(name)
	if err != nil {
		return err
	}
	t.trendFunction = trendFunc
	t.FuncName = name
	return nil
}

// Returns the start delay of the trend anomaly in seconds.
func (t *trendAnomaly) GetStartDelay() float64 {
	return t.StartDelay
}

// Returns the duration of each trend anomaly in seconds.
func (t *trendAnomaly) GetDuration() float64 {
	return t.Duration
}

// Returns the name of the trend function.
func (t *trendAnomaly) GetTrendFuncName() string {
	return t.FuncName
}

// Returns whether the trend anomaly is active this timestep.
func (t *trendAnomaly) GetIsAnomalyActive() bool {
	return t.isAnomalyActive
}

// Returns t.startDelay as a number of time steps.
func (t *trendAnomaly) GetStartDelayIndex() int {
	return t.startDelayIndex
}

// Returns the time elapsed since the start of the active trend anomaly.
func (t *trendAnomaly) GetElapsedTrendTime() float64 {
	return t.elapsedTrendTime
}

// Returns the number of time steps since the start of the active trend anomaly.
func (t *trendAnomaly) GetTrendProgressIndex() int {
	return t.elapsedTrendIndex
}

// Returns the number of times the trend anomaly has repeated so far.
func (t *trendAnomaly) GetCountRepeats() int {
	return t.countRepeats
}

// Returns the trend function used by the trend anomaly.
func (t *trendAnomaly) GetTrendFunction() mathfuncs.TrendFunction {
	return t.trendFunction
}

// Returns the change in signal caused by the trend anomaly this timestep.
// Manages internal indices to track the progress of trend cycles, and delays between trend cycles.
// Ts is the sampling period of the data.
func (t *trendAnomaly) stepAnomaly(_ *rand.Rand, Ts float64) float64 {
	if t.Off {
		return 0.0
	}
	// Check if the trend anomaly is active this timestep
	t.isAnomalyActive = t.isTrendsAnomalyActive(Ts)
	if !t.isTrendsAnomalyActive(Ts) {
		t.startDelayIndex += 1 // only increment if inactive to keep track of the delay between trend cycles
		return 0.0
	}

	t.elapsedTrendTime = float64(t.elapsedTrendIndex) * Ts

	trendAnomalyMagnitude := t.trendFunction(t.elapsedTrendTime, t.Magnitude, t.Duration)
	trendAnomalyDelta := t.getTrendAnomalySign() * trendAnomalyMagnitude
	t.elapsedTrendIndex += 1

	// If the trend anomaly is complete, reset the index and increment the repeat counter
	if t.elapsedTrendIndex == int(t.Duration/Ts) {
		t.elapsedTrendIndex = 0
		t.startDelayIndex = 0
		t.countRepeats += 1
	}

	return trendAnomalyDelta
}

// Returns whether trend anomalies should be active this timestep. This is true if:
//  1. Enough time has elapsed for the trend anomaly to start, and;
//  2. The trend anomaly has not yet completed all repetitions.
func (t *trendAnomaly) isTrendsAnomalyActive(Ts float64) bool {
	moreRepeatsAllowed := t.countRepeats < t.countRepeats || t.Repeats == 0 // 0 means infinite repetitions
	if !moreRepeatsAllowed {
		t.Off = true // switch the trend off if all repetitions are complete to save future computation
		return false
	}

	hasTrendStarted := t.startDelayIndex >= int(t.StartDelay/Ts)-1
	return hasTrendStarted
}

// Returns -1.0 if InvertTrend is true, or +1.0 if false.
func (t *trendAnomaly) getTrendAnomalySign() float64 {
	if t.InvertTrend {
		return -1.0
	}
	return 1.0
}

func (t *trendAnomaly) typeAsString() string {
	return "trend"
}
