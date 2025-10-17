package anomaly

import (
	"errors"
	"math/rand/v2"

	"github.com/synaptecltd/emulator/mathfuncs"
)

// Modulates waveform data using continuous functions.
type trendAnomaly struct {
	AnomalyBase

	Magnitude    float64 // magnitude of trend anomaly, default 0
	magFuncName  string  // name of function to use to vary the trend magnitude, defaults to "linear" if empty
	InvertTrend  bool    // true inverts the trend function (multiplies by -1.0), default false (no inverting)
	ReverseTrend bool    // true subtracts the original value by 'Magnitude' (equivalent of reversing along horizontal axis)

	// internal state
	magFunction mathfuncs.MathsFunction // returns trend anomaly magnitude for a given elapsed time, magntiude and period; set internally from TrendFuncName
}

// Parameters to use for the trend anomaly. All can be accessed publicly and used to define trendAnomaly.
type TrendParams struct {
	// Defined in AnomalyBase

	Name       string  `yaml:"Name"`       // name of the anomaly, used for identification
	Repeats    uint64  `yaml:"Repeats"`    // the number of times the trend anomaly repeats, 0 for infinite
	Off        bool    `yaml:"Off"`        // true: anomaly deactivated, false: activated
	StartDelay float64 `yaml:"StartDelay"` // the delay before trend anomalies begin (and between anomaly repeats) in seconds
	Duration   float64 `yaml:"Duration"`   // the duration of each trend anomaly in seconds, 0 for continuous

	// Defined in trendAnomaly

	Magnitude    float64 `yaml:"Magnitude"` // magnitude of trend anomaly, default 0
	MagFuncName  string  `yaml:"MagFunc"`   // name of the function used to vary the magnitude of the trend anomaly, empty defaults to "linear"
	InvertTrend  bool    `yaml:"Invert"`    // true inverts the trend function (multiplies by -1.0), default false (no inverting)
	ReverseTrend bool    `yaml:"Reverse"`   // true subtracts the original value by 'Magnitude' (equivalent of reversing along horizontal axis)
}

// Initialise the internal fields of TrendAnomaly when it is unmarshalled from yaml.
func (t *trendAnomaly) UnmarshalYAML(unmarshal func(any) error) error {
	var params TrendParams
	if err := unmarshal(&params); err != nil {
		return err
	}

	// This performs checking for invalid values
	trendAnomaly, err := NewTrendAnomaly(params)
	if err != nil {
		return err
	}

	// Copy fields to t
	*t = *trendAnomaly

	return nil
}

// Returns a trendAnomaly pointer with the requested parameters, checking for invalid values.
func NewTrendAnomaly(params TrendParams) (*trendAnomaly, error) {
	trendAnomaly := &trendAnomaly{}

	trendAnomaly.name = params.Name

	// Invalid values checked by setters
	if err := trendAnomaly.SetDuration(params.Duration); err != nil {
		return nil, err
	}
	if err := trendAnomaly.SetStartDelay(params.StartDelay); err != nil {
		return nil, err
	}
	if err := trendAnomaly.SetMagFunctionByName(params.MagFuncName); err != nil {
		return nil, err
	}

	// Fields that can never be invalid set directly
	trendAnomaly.typeName = "trend"
	trendAnomaly.Magnitude = params.Magnitude
	trendAnomaly.Repeats = params.Repeats
	trendAnomaly.InvertTrend = params.InvertTrend
	trendAnomaly.Off = params.Off

	return trendAnomaly, nil
}

// stepAnomaly returns the change in signal caused by the trend anomaly this timestep.
// Manages internal indices to track the progress of trend cycles, and delays between trend cycles.
// Ts is the sampling period of the data.
func (t *trendAnomaly) stepAnomaly(_ *rand.Rand, Ts float64) float64 {
	if t.Off {
		return 0.0
	}
	// Check if the trend anomaly is active this timestep
	t.isAnomalyActive = t.CheckAnomalyActive(Ts)
	if !t.isAnomalyActive {
		t.startDelayIndex += 1 // increment to keep track of the delay between trend repeats
		return 0.0
	}

	// Update the index after logging the current time
	t.elapsedActivatedTime = float64(t.elapsedActivatedIndex) * Ts
	t.elapsedActivatedIndex += 1

	trendAnomalyMagnitude := t.magFunction(t.elapsedActivatedTime, t.Magnitude, t.duration)
	trendAnomalyDelta := t.getSign() * trendAnomalyMagnitude
	if t.ReverseTrend {
		trendAnomalyDelta = t.Magnitude - trendAnomalyMagnitude
	}

	// If the trend anomaly is complete, reset the index and increment the repeat counter
	if t.elapsedActivatedIndex == int(t.duration/Ts) {
		t.elapsedActivatedIndex = 0
		t.startDelayIndex = 0
		t.countRepeats += 1
	}

	return trendAnomalyDelta
}

// Returns -1.0 if InvertTrend is true, or +1.0 if false.
func (t *trendAnomaly) getSign() float64 {
	if t.InvertTrend {
		return -1.0
	}
	return 1.0
}

// Setters

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

func (t *trendAnomaly) SetMagFunctionByName(name string) error {
	if name == "" {
		name = "linear" // default to linear if no name is provided
	}
	return t.SetFunctionByName(name, mathfuncs.GetTrendFunctionFromName, &t.magFuncName, &t.magFunction)
}

// Getters

func (t *trendAnomaly) GetMagFuncName() string {
	return t.magFuncName
}

// Returns the trend function used by the trend anomaly.
func (t *trendAnomaly) GetMagFunction() mathfuncs.MathsFunction {
	return t.magFunction
}
