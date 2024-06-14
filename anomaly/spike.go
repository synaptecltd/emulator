package anomaly

import (
	"errors"
	"math"
	"math/rand/v2"

	"github.com/synaptecltd/emulator/mathfuncs"
)

// SpikeAnomaly produces spikes in the data that occur at each timestep based on a probability factor.
type spikeAnomaly struct {
	AnomalyBase

	// Private fields have setters for invalid value checking

	Magnitude     float64 // magnitude of spikes, default 0
	magFuncName   string  // name of the function used to vary the magnitude of the spikes, empty defaults to no functional modulation
	VaryMagnitude bool    // whether apply Gaussian variation to magnitude of spikes, default false
	spikeSign     float64 // the probability of spikes being positive or negative. default 0 (equally likely +/-). negative numbers favour negative spikes, positive numbers favour positive spikes

	probability  float64 // magnitude of probability of spike in each time step, default 0
	probFuncName string  // name of the function used to vary the probability of the spikes, empty defaults to constant =probability

	// internal state
	magFunction  mathfuncs.MathsFunction // returns spike anomaly magnitude for a given elapsed time, magntiude and period; set internally from magFuncName
	probFunction mathfuncs.MathsFunction // returns spike anomaly probability for a given elapsed time, magntiude and period; set internally from probFuncName
}

// Parameters used to request a spike anomaly. These map onto the fields of spikeAnomaly.
type SpikeParams struct {
	// Defined in AnomalyBase

	Repeats    uint64  `yaml:"repeat"`      // the number of times spike bursts repeat, 0 for infinite
	Off        bool    `yaml:"off"`         // true: anomaly deactivated, false: activated
	StartDelay float64 `yaml:"start_delay"` // the delay before spike bursts begin (and time between bursts) in seconds
	Duration   float64 `yaml:"duration"`    // the duration of burst of spikes in seconds, 0 for continuous

	// Defined in spikeAnomaly

	Magnitude     float64 `yaml:"magnitude"`      // magnitude of spikes, default 0
	MagFuncName   string  `yaml:"mag_func"`       // name of the function used to vary the magnitude of the spikes, empty defaults to no functional modulation
	VaryMagnitude bool    `yaml:"vary_magnitude"` // whether apply Gaussian variation to magnitude of spikes, default false
	SpikeSign     float64 `yaml:"spike_sign"`     // the probability of spikes being positive or negative. default 0 (equally likely +/-). negative numbers favour negative spikes, positive numbers favour positive spikes

	Probability  float64 `yaml:"probability"` // magnitude of probability of spike in each time step, default 0
	ProbFuncName string  `yaml:"prob_func"`   // name of the function used to vary the probability of the spikes, empty defaults to constant =probability
}

// Initialise the internal fields of SpikeAnomaly when it is unmarshalled from yaml.
func (s *spikeAnomaly) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var params SpikeParams
	if err := unmarshal(&params); err != nil {
		return err
	}

	// This performs checking for invalid values
	spikeAnomaly, err := NewSpikeAnomaly(params)
	if err != nil {
		return err
	}

	// Copy fields to s
	*s = *spikeAnomaly

	return nil
}

// Returns a spikeAnomaly pointer with the requested parameters, checking for invalid values.
func NewSpikeAnomaly(params SpikeParams) (*spikeAnomaly, error) {
	spikeAnomaly := &spikeAnomaly{}

	// Invalid values checked by setters
	if err := spikeAnomaly.SetStartDelay(params.StartDelay); err != nil {
		return nil, err
	}
	if err := spikeAnomaly.SetProbability(params.Probability); err != nil {
		return nil, err
	}
	if err := spikeAnomaly.SetMagFunctionByName(params.MagFuncName); err != nil {
		return nil, err
	}
	if err := spikeAnomaly.SetProbFunctionByName(params.ProbFuncName); err != nil {
		return nil, err
	}
	if err := spikeAnomaly.SetSpikeSign(params.SpikeSign); err != nil {
		return nil, err
	}
	if err := spikeAnomaly.SetDuration(params.Duration); err != nil {
		return nil, err
	}

	// Fields that can never be invalid set directly
	spikeAnomaly.typeName = "spike"
	spikeAnomaly.Magnitude = params.Magnitude
	spikeAnomaly.VaryMagnitude = params.VaryMagnitude
	spikeAnomaly.Repeats = params.Repeats
	spikeAnomaly.Off = params.Off

	return spikeAnomaly, nil
}

// Returns the change in signal caused by the instantaneous anomaly this timestep.
func (s *spikeAnomaly) stepAnomaly(r *rand.Rand, Ts float64) float64 {
	if s.Off {
		return 0.0
	}

	// Check if the spike anomaly is active this timestep
	s.isAnomalyActive = s.CheckAnomalyActive(Ts)
	if !s.isAnomalyActive {
		s.startDelayIndex += 1
		return 0.0
	}

	s.elapsedActivatedTime = float64(s.elapsedActivatedIndex) * Ts

	// No anomaly if probability is not met
	probThisStep := s.probability
	if s.probFunction != nil {
		probThisStep = s.probFunction(s.elapsedActivatedTime, s.probability, s.duration)
		// take positive values only
		probThisStep = math.Abs(probThisStep)
	}

	if r.Float64() > probThisStep {
		s.isAnomalyActive = false
		s.elapsedActivatedIndex += 1 // still increment to keep the bursts spaced out
		return 0.0
	}

	s.isAnomalyActive = true
	returnval := s.Magnitude * s.getSign(r)
	if s.VaryMagnitude {
		returnval = returnval * r.NormFloat64()
	}

	// if duration is negative the spike anomaly is continuous and there is no need to worry about
	// repeats or elapsedActivatedIndex or functions for magnitude
	if s.duration < 0 {
		return returnval
	}

	// If a function is set, use it to vary the magnitude of the spikes
	if s.magFunction != nil {
		returnval = s.magFunction(s.elapsedActivatedTime, s.Magnitude, s.duration) * s.getSign(r)
	}
	if s.VaryMagnitude {
		returnval = returnval * r.NormFloat64()
	}

	s.elapsedActivatedIndex += 1

	// If the spike anomaly is complete, reset the index and increment the repeat counter
	if s.elapsedActivatedIndex >= int(s.duration/Ts)-1 {
		s.elapsedActivatedIndex = 0
		s.startDelayIndex = 0
		s.countRepeats += 1
	}

	return returnval
}

// Returns -1.0 or +1.0 with a probability based on the spikeSign parameter.
// If SpikeSign is 0, -1.0 and +1.0 are returned with equal probability.
func (s *spikeAnomaly) getSign(r *rand.Rand) float64 {
	if r.Float64()*2-1 > s.spikeSign {
		return -1.0
	} else {
		return 1.0
	}
}

// Setters

// Sets the duration of each spike anomaly in seconds. If duration=0, the spike anomaly
// defined as is continuous (duration=-1.0).
func (s *spikeAnomaly) SetDuration(duration float64) error {
	if duration == 0 {
		if s.magFunction != nil {
			return errors.New("duration must be greater than 0 when using a functional dependence for magntiude")
		}
		duration = -1.0 // continuous burst
	}
	s.duration = duration
	return nil
}

// Set probability of spike anomalies occurring each timestep if probability >= 0.
func (s *spikeAnomaly) SetProbability(probability float64) error {
	if probability < 0 {
		return errors.New("probability must be greater than or equal to 0")
	}

	s.probability = probability
	return nil
}

func (s *spikeAnomaly) SetSpikeSign(spikeSign float64) error {
	if spikeSign < -1.0 || spikeSign > 1.0 {
		return errors.New("spike sign must be between -1 and 1")
	}
	s.spikeSign = spikeSign
	return nil
}

// Sets the field magFunction to the function with the given name.
func (s *spikeAnomaly) SetMagFunctionByName(name string) error {
	return s.SetFunctionByName(name, mathfuncs.GetTrendFunctionFromName, &s.magFuncName, &s.magFunction)
}

// Sets the field probFunction to the function with the given name.
func (s *spikeAnomaly) SetProbFunctionByName(name string) error {
	return s.SetFunctionByName(name, mathfuncs.GetTrendFunctionFromName, &s.probFuncName, &s.probFunction)
}

// Getters

func (s *spikeAnomaly) GetProbability() float64 {
	return s.probability
}

func (s *spikeAnomaly) GetSpikeSign() float64 {
	return s.spikeSign
}

func (s *spikeAnomaly) GetMagFunctionName() mathfuncs.MathsFunction {
	return s.magFunction
}

func (s *spikeAnomaly) GetProbFunctionName() string {
	return s.probFuncName
}

func (s *spikeAnomaly) GetMagFunction() mathfuncs.MathsFunction {
	return s.magFunction
}

func (s *spikeAnomaly) GetProbFunction() mathfuncs.MathsFunction {
	return s.probFunction
}
