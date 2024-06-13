package anomaly

import "math/rand/v2"

// InstantaneousAnomaly produces spikes in the data that occur at each timestep based on a probability factor.
type InstantaneousAnomaly struct {
	Probability     float64 `yaml:"probability"`    // probability of instantaneous anomaly in each time step
	Magnitude       float64 `yaml:"magnitude"`      // magnitude of instantaneous anomalies
	VaryMagnitude   bool    `yaml:"vary_magnitude"` // whether to vary the magnitude of instantaneous anomaly spikes, default false
	isAnomalyActive bool    // indicates whether instantaneous anomaly spike is active in this time step
}

func (ia *InstantaneousAnomaly) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain InstantaneousAnomaly
	if err := unmarshal((*plain)(ia)); err != nil {
		return err
	}

	// Add any additional logic here if needed

	return nil
}

// Returns the change in signal caused by the instantaneous anomaly this timestep.
func (ia *InstantaneousAnomaly) stepAnomaly(r *rand.Rand, _ float64) float64 {
	// No anomaly if probability is not met
	if r.Float64() > ia.Probability {
		ia.isAnomalyActive = false
		return 0.0
	}

	ia.isAnomalyActive = true
	if ia.VaryMagnitude {
		return ia.Magnitude * r.NormFloat64()
	}
	return ia.Magnitude
}

func (ia *InstantaneousAnomaly) TypeAsString() string {
	return "instantaneous"
}

// Returns whether the instantaneous anomaly is active this timestep.
func (ia *InstantaneousAnomaly) GetIsAnomalyActive() bool {
	return ia.isAnomalyActive
}
