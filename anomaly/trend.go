package anomaly

import (
	"errors"
	"math/rand/v2"

	"github.com/synaptecltd/emulator/mathfuncs"
)

// trendAnomaly modulates data using repeated continuous functions.
// The type is private so that it can only be created using NewTrendAnomaly or by
// unmarshalling from yaml.
type trendAnomaly struct {
	AnomalyBase

	Magnitude   float64 // magnitude of trend anomaly
	magFuncName string  // name of function to use to vary the trend magnitude, defaults to "linear" if empty
	InvertTrend bool    // true inverts the trend function (multiplies by -1.0), default false (no inverting)

	// internal state
	magFunction mathfuncs.TrendFunction // returns trend anomaly magnitude for a given elapsed time, magntiude and period; set internally from TrendFuncName
}

// Parameters to use for the trend anomaly. All can be accessed publicly and used to define trendAnomaly.
type TrendParams struct {
	StartDelay  float64 `yaml:"start_delay"`
	Duration    float64 `yaml:"duration"`
	Magnitude   float64 `yaml:"magnitude"`
	Repeats     uint64  `yaml:"repeat"`
	FuncName    string  `yaml:"trend_function"`
	InvertTrend bool    `yaml:"invert"`
	Off         bool    `yaml:"off"`
}

// Initialise the internal fields of TrendAnomaly when it is unmarshalled from yaml.
func (t *trendAnomaly) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var params TrendParams
	err := unmarshal(&params)
	if err != nil {
		return err
	}

	// Use the duplicate's fields to initialise a new TrendAnomaly
	trendAnomaly, err := NewTrendAnomaly(params)
	if err != nil {
		return err
	}

	// Copy fields from newTrend to t
	*t = *trendAnomaly

	return nil
}

// Returns a TrendAnomaly with the specified parameters, checking for invalid values.
func NewTrendAnomaly(params TrendParams) (*trendAnomaly, error) {
	trendAnomaly := &trendAnomaly{}

	if err := trendAnomaly.SetDuration(params.Duration); err != nil {
		return nil, err
	}

	if err := trendAnomaly.SetStartDelay(params.StartDelay); err != nil {
		return nil, err
	}

	if err := trendAnomaly.SetMagFunctionByName(params.FuncName); err != nil {
		return nil, err
	}

	// Fields that can never be invalid are set directly
	trendAnomaly.typeName = "trend"
	trendAnomaly.Magnitude = params.Magnitude
	trendAnomaly.Repeats = params.Repeats
	trendAnomaly.InvertTrend = params.InvertTrend
	trendAnomaly.Off = params.Off

	return trendAnomaly, nil
}

// Sets the duration of each trend anomaly in seconds if duration > 0.
// If duration=0, the trend anomaly is deactivated.
func (t *trendAnomaly) SetDuration(duration float64) error {
	if duration < 0 {
		return errors.New("duration must be positive value")
	}
	if duration == 0 {
		t.Off = true
	}
	t.duration = duration
	return nil
}

// Sets the function of the trend anomaly using the name of the function if the
// name is valid.
func (t *trendAnomaly) SetMagFunctionByName(name string) error {
	return t.SetFunctionByName(name, mathfuncs.GetTrendFunctionFromName, &t.magFuncName, &t.magFunction)
}

// Returns the name of the trend function.
func (t *trendAnomaly) GetTrendFuncName() string {
	return t.magFuncName
}

// Returns the trend function used by the trend anomaly.
func (t *trendAnomaly) GetTrendFunction() mathfuncs.TrendFunction {
	return t.magFunction
}

// Returns the change in signal caused by the trend anomaly this timestep.
// Manages internal indices to track the progress of trend cycles, and delays between trend cycles.
// Ts is the sampling period of the data.
func (t *trendAnomaly) stepAnomaly(_ *rand.Rand, Ts float64) float64 {
	if t.Off {
		return 0.0
	}
	// Check if the trend anomaly is active this timestep
	t.isAnomalyActive = t.CheckAnomalyActive(Ts)
	if !t.isAnomalyActive {
		t.startDelayIndex += 1 // only increment if inactive to keep track of the delay between trend cycles
		return 0.0
	}

	t.elapsedActivatedTime = float64(t.elapsedActivatedIndex) * Ts

	trendAnomalyMagnitude := t.magFunction(t.elapsedActivatedTime, t.Magnitude, t.duration)
	trendAnomalyDelta := t.getTrendAnomalySign() * trendAnomalyMagnitude
	t.elapsedActivatedIndex += 1

	// If the trend anomaly is complete, reset the index and increment the repeat counter
	if t.elapsedActivatedIndex == int(t.duration/Ts) {
		t.elapsedActivatedIndex = 0
		t.startDelayIndex = 0
		t.countRepeats += 1
	}

	return trendAnomalyDelta
}

// Returns -1.0 if InvertTrend is true, or +1.0 if false.
func (t *trendAnomaly) getTrendAnomalySign() float64 {
	if t.InvertTrend {
		return -1.0
	}
	return 1.0
}
