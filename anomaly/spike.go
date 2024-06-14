package anomaly

import (
	"errors"
	"math/rand/v2"

	"github.com/synaptecltd/emulator/mathfuncs"
)

// SpikeAnomaly produces spikes in the data that occur at each timestep based on a probability factor.
type SpikeAnomaly struct {
	probability   float64 // probability of spike in each time step, default 0
	Magnitude     float64 // magnitude of spikes, default 0
	VaryMagnitude bool    // whether to vary the magnitude of spikes, default false
	SpikeSign     float64 // positive value only allows positive spikes, negative only allows negative, default 0 allows both

	duration   float64 // duration of each burst of spike anomalies in seconds, negative values mean continuous burst, default -1
	startDelay float64 // start time for spike anomalies to start occuring in seconds, default 0
	Repeats    uint64  // number of times bursts of spike anomalies repeat, default 0 for infinite
	Off        bool    // true: spike anomaly deactivated, false: activated (default)

	funcName string // name of the function used to vary the magnitude of the spikes, empty defaults to no functional modulation

	elapsedActivatedIndex int     // number of time steps since start of active burst of spike anomaly, used to track the progress of bursts
	elapsedActivatedTime  float64 // as above but in seconds
	isAnomalyActive       bool    // indicates whether a spike anomaly (not burst, but all) is active in this time step
	startDelayIndex       int     // startDelay converted to time steps, used to track delay period between instantaneous anomaly bursts
	countRepeats          uint64

	spikeFunction mathfuncs.TrendFunction // returns spike anomaly magnitude for a given elapsed time, magntiude and period; set internally from FuncName

	// TODO vary anomaly probability using trends
}

// Parameters used for spike anomaly
type SpikeParams struct {
	Probability   float64 `yaml:"probability"`
	Magnitude     float64 `yaml:"magnitude"`
	VaryMagnitude bool    `yaml:"vary_magnitude"`
	Duration      float64 `yaml:"duration"`
	StartDelay    float64 `yaml:"start_delay"`
	Repeats       uint64  `yaml:"repeat"`
	Off           bool    `yaml:"off"`
	FuncName      string  `yaml:"func_name"`
	SpikeSign     float64 `yaml:"spike_sign"`
}

func (ia *SpikeAnomaly) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain SpikeAnomaly
	if err := unmarshal((*plain)(ia)); err != nil {
		return err
	}

	// Add any additional logic here if needed
	// TODO : Checks for valid values are actually needed. Probability needs to be private as does instantAnomalyActive

	return nil
}

func NewSpikeAnomaly(params SpikeParams) (*SpikeAnomaly, error) {
	spikeAnomaly := &SpikeAnomaly{}

	if err := spikeAnomaly.SetStartDelay(params.StartDelay); err != nil {
		return nil, err
	}

	if err := spikeAnomaly.SetProbability(params.Probability); err != nil {
		return nil, err
	}

	if err := spikeAnomaly.SetFunctionByName(params.FuncName); err != nil {
		return nil, err
	}

	spikeAnomaly.SetDuration(params.Duration)

	spikeAnomaly.Magnitude = params.Magnitude
	spikeAnomaly.VaryMagnitude = params.VaryMagnitude
	spikeAnomaly.Repeats = params.Repeats
	spikeAnomaly.Off = params.Off
	spikeAnomaly.SpikeSign = params.SpikeSign

	return spikeAnomaly, nil
}

func (s *SpikeAnomaly) SetFunctionByName(name string) error {
	if name == "" {
		s.funcName = name
		s.spikeFunction = nil
		return nil
	}

	trendFunc, err := mathfuncs.GetTrendFunctionFromName(name)
	if err != nil {
		return err
	}
	s.spikeFunction = trendFunc
	s.funcName = name
	return nil
}

func (s *SpikeAnomaly) SetDuration(duration float64) error {
	if duration == 0 {
		if s.spikeFunction != nil {
			return errors.New("duration must be greater than 0 when using a functional dependence for magntiude")
		}
		duration = -1.0 // continuous burst
	}
	s.duration = duration
	return nil
}

// Sets the start time of spike anomalies in seconds if delay >= 0.
func (s *SpikeAnomaly) SetStartDelay(startDelay float64) error {
	if startDelay < 0 {
		return errors.New("startDelay must be greater than or equal to 0")
	}

	s.startDelay = startDelay
	return nil
}

// Sets probability of spike anomalies occurring each timestep if probability >= 0.
func (s *SpikeAnomaly) SetProbability(probability float64) error {
	if probability < 0 {
		return errors.New("probability must be greater than or equal to 0")
	}

	s.probability = probability
	return nil
}

// Returns the change in signal caused by the instantaneous anomaly this timestep.
func (ia *SpikeAnomaly) stepAnomaly(r *rand.Rand, Ts float64) float64 {
	if ia.Off {
		return 0.0
	}

	// Check if the spike anomaly is active this timestep
	ia.isAnomalyActive = ia.isSpikeAnomalyActive(Ts)
	if !ia.isAnomalyActive {
		ia.startDelayIndex += 1
		return 0.0
	}

	// No anomaly if probability is not met
	if r.Float64() > ia.probability {
		ia.isAnomalyActive = false
		ia.elapsedActivatedIndex += 1 // still increment to keep the bursts spaced out
		return 0.0
	}

	ia.isAnomalyActive = true
	// TODO something squiffy here
	returnval := ia.Magnitude * ia.GetSpikeSignFromSpikeSign()
	if ia.VaryMagnitude {
		returnval = returnval * r.NormFloat64()
	}

	// if duration is negative the spike anomaly is continuous and there is no need to worry about
	// repeats or elapsedActivatedIndex or functions
	if ia.duration < 0 {
		return returnval
	}

	ia.elapsedActivatedTime = float64(ia.elapsedActivatedIndex) * Ts

	// If a function is set, use it to vary the magnitude of the spikes
	if ia.spikeFunction != nil {
		returnval = ia.spikeFunction(ia.elapsedActivatedTime, ia.Magnitude, ia.duration)
	}

	ia.elapsedActivatedIndex += 1

	// If the spike anomaly is complete, reset the index and increment the repeat counter
	if ia.elapsedActivatedIndex >= int(ia.duration/Ts)-1 {
		ia.elapsedActivatedIndex = 0
		ia.startDelayIndex = 0
		ia.countRepeats += 1
	}

	return returnval
}

func (ia *SpikeAnomaly) GetSpikeSignFromSpikeSign() float64 {
	if ia.SpikeSign < 0 {
		return -1.0
	} else if ia.SpikeSign > 0 {
		return 1.0
	} else {
		if rand.Float64() < 0.5 {
			return -1.0
		}
		return 1.0
	}
}

func (ia *SpikeAnomaly) TypeAsString() string {
	return "instantaneous"
}

// Returns whether the instantaneous anomaly is active this timestep.
func (ia *SpikeAnomaly) GetIsAnomalyActive() bool {
	return ia.isAnomalyActive
}

func (ia *SpikeAnomaly) GetDuration() float64 {
	return ia.duration
}

func (ia *SpikeAnomaly) GetStartDelay() float64 {
	return ia.startDelay
}

// Returns whether spike anomalies should be active this timestep. This is true if:
//  1. Enough time has elapsed for the spike anomaly to start, and;
//  2. The spike anomaly has not yet completed all repetitions.
func (ia *SpikeAnomaly) isSpikeAnomalyActive(Ts float64) bool {
	moreRepeatsAllowed := ia.countRepeats < ia.Repeats || ia.Repeats == 0 // 0 means infinite repetitions

	if !moreRepeatsAllowed {
		ia.Off = true // switch the spike off if all repetitions are complete to save future computation
		return false
	}

	hasSpikeStarted := ia.startDelayIndex >= int(ia.startDelay/Ts)-1
	return hasSpikeStarted
}
