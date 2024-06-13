package emulatoranomaly

import "math/rand/v2"

// InstantaneousAnomaly produces spikes in the data that occur at each timestep based on a probability factor.
type InstantaneousAnomaly struct {
	InstantaneousAnomalyProbability        float64 `yaml:"InstantaneousAnomalyProbability"`        // probability of instantaneous anomaly in each time step
	InstantaneousAnomalyMagnitude          float64 `yaml:"InstantaneousAnomalyMagnitude"`          // magnitude of instantaneous anomalies
	InstantaneousAnomalyMagnitudeVariation bool    `yaml:"InstantaneousAnomalyMagnitudeVariation"` // whether to vary the magnitude of instantaneous anomaly spikes, default false
	InstantaneousAnomalyActive             bool    // indicates whether instantaneous anomaly spike is active in this time step
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
	if r.Float64() > ia.InstantaneousAnomalyProbability {
		ia.InstantaneousAnomalyActive = false
		return 0.0
	}

	ia.InstantaneousAnomalyActive = true
	if ia.InstantaneousAnomalyMagnitudeVariation {
		return ia.InstantaneousAnomalyMagnitude * r.NormFloat64()
	}
	return ia.InstantaneousAnomalyMagnitude
}

func (ia *InstantaneousAnomaly) typeAsString() string {
	return "instantaneous"
}
