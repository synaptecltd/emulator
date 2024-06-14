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

	// Setters and getters are provided for private fields below to allow for error checking
	probability   float64 // magnitude of probability of spike in each time step, default 0
	Magnitude     float64 // magnitude of spikes, default 0
	VaryMagnitude bool    // whether to vary the magnitude of spikes with Gaussian variation, default false
	spikeSign     float64 // the probability of spikes being positive or negative. default 0 (equally likely +/-). negative numbers favour negative spikes, positive numbers favour positive spikes

	magFuncName  string // name of the function used to vary the magnitude of the spikes, empty defaults to no functional modulation
	probFuncName string // name of the function used to vary the probability of the spikes, empty defaults to no functional modulation

	// internal state
	magFunction  mathfuncs.TrendFunction // returns spike anomaly magnitude for a given elapsed time, magntiude and period; set internally from FuncName
	probFunction mathfuncs.TrendFunction // returns spike anomaly probability for a given elapsed time, magntiude and period; set internally from FuncName
}

// Parameters used for spike anomaly
type SpikeParams struct {
	Probability     float64 `yaml:"probability"`
	Magnitude       float64 `yaml:"magnitude"`
	VaryMagnitude   bool    `yaml:"vary_magnitude"`
	Duration        float64 `yaml:"duration"`
	StartDelay      float64 `yaml:"start_delay"`
	Repeats         uint64  `yaml:"repeat"`
	Off             bool    `yaml:"off"`
	MagnitudeFunc   string  `yaml:"mag_func"`
	ProbabilityFunc string  `yaml:"prob_func"`
	SpikeSign       float64 `yaml:"spike_sign"`
}

func (s *spikeAnomaly) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var params SpikeParams
	err := unmarshal(&params)
	if err != nil {
		return err
	}

	spikeAnomaly, err := NewSpikeAnomaly(params)
	if err != nil {
		return err
	}

	// Copy fields
	*s = *spikeAnomaly

	return nil
}

func NewSpikeAnomaly(params SpikeParams) (*spikeAnomaly, error) {
	spikeAnomaly := &spikeAnomaly{}

	if err := spikeAnomaly.SetStartDelay(params.StartDelay); err != nil {
		return nil, err
	}

	if err := spikeAnomaly.SetProbability(params.Probability); err != nil {
		return nil, err
	}

	if err := spikeAnomaly.SetMagFunctionByName(params.MagnitudeFunc); err != nil {
		return nil, err
	}

	if err := spikeAnomaly.SetProbFunctionByName(params.ProbabilityFunc); err != nil {
		return nil, err
	}

	spikeAnomaly.SetDuration(params.Duration)
	spikeAnomaly.SetSpikeSign(params.SpikeSign)

	spikeAnomaly.typeName = "spike"
	spikeAnomaly.Magnitude = params.Magnitude
	spikeAnomaly.VaryMagnitude = params.VaryMagnitude
	spikeAnomaly.Repeats = params.Repeats
	spikeAnomaly.Off = params.Off

	return spikeAnomaly, nil
}

func (s *spikeAnomaly) SetMagFunctionByName(name string) error {
	return s.SetFunctionByName(name, mathfuncs.GetTrendFunctionFromName, &s.magFuncName, &s.magFunction)
}

func (s *spikeAnomaly) SetProbFunctionByName(name string) error {
	return s.SetFunctionByName(name, mathfuncs.GetTrendFunctionFromName, &s.probFuncName, &s.probFunction)
}

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

// Sets probability of spike anomalies occurring each timestep if probability >= 0.
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

// Returns the change in signal caused by the instantaneous anomaly this timestep.
func (ia *spikeAnomaly) stepAnomaly(r *rand.Rand, Ts float64) float64 {
	if ia.Off {
		return 0.0
	}

	// Check if the spike anomaly is active this timestep
	ia.isAnomalyActive = ia.CheckAnomalyActive(Ts)
	if !ia.isAnomalyActive {
		ia.startDelayIndex += 1
		return 0.0
	}

	ia.elapsedActivatedTime = float64(ia.elapsedActivatedIndex) * Ts

	// No anomaly if probability is not met
	probThisStep := ia.probability
	if ia.probFunction != nil {
		probThisStep = ia.probFunction(ia.elapsedActivatedTime, ia.probability, ia.duration)
		// take positive values only
		probThisStep = math.Abs(probThisStep)
	}

	if r.Float64() > probThisStep {
		ia.isAnomalyActive = false
		ia.elapsedActivatedIndex += 1 // still increment to keep the bursts spaced out
		return 0.0
	}

	ia.isAnomalyActive = true
	returnval := ia.Magnitude * ia.GetSpikeSignFromSpikeSign(r)
	if ia.VaryMagnitude {
		returnval = returnval * r.NormFloat64()
	}

	// if duration is negative the spike anomaly is continuous and there is no need to worry about
	// repeats or elapsedActivatedIndex or functions for magnitude
	if ia.duration < 0 {
		return returnval
	}

	// If a function is set, use it to vary the magnitude of the spikes
	if ia.magFunction != nil {
		returnval = ia.magFunction(ia.elapsedActivatedTime, ia.Magnitude, ia.duration) * ia.GetSpikeSignFromSpikeSign(r)
	}
	if ia.VaryMagnitude {
		returnval = returnval * r.NormFloat64()
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

// Returns the sign of the spike anomaly based on the SpikeSign parameter.
// If SpikeSign is positive, only positive spikes are allowed.
// If SpikeSign is negative, only negative spikes are allowed.
// If SpikeSign is 0, both positive and negative spikes are allowed with equal probability.
// Values in between 0 and 1 will allow positive spikes with a probability equal to the value.
func (ia *spikeAnomaly) GetSpikeSignFromSpikeSign(r *rand.Rand) float64 {
	if r.Float64()*2-1 > ia.spikeSign {
		return -1.0
	} else {
		return 1.0
	}
}

func (s *spikeAnomaly) GetProbability() float64 {
	return s.probability
}

func (s *spikeAnomaly) GetSpikeSign() float64 {
	return s.spikeSign
}

func (s *spikeAnomaly) GetMagFunction() mathfuncs.TrendFunction {
	return s.magFunction
}

func (s *spikeAnomaly) GetProbFunction() mathfuncs.TrendFunction {
	return s.probFunction
}
