package emulatoranomaly

import "math/rand/v2"

// InstantaneousAnomaly produces spikes in the data that occur at each timestep based on a probability factor.
type InstantaneousAnomaly struct {
	InstantaneousAnomalyProbability        float64 `yaml:"InstantaneousAnomalyProbability"`        // probability of instantaneous anomaly in each time step
	InstantaneousAnomalyMagnitude          float64 `yaml:"InstantaneousAnomalyMagnitude"`          // magnitude of instantaneous anomalies
	InstantaneousAnomalyMagnitudeVariation bool    `yaml:"InstantaneousAnomalyMagnitudeVariation"` // whether to vary the magnitude of instantaneous anomaly spikes, default false
	InstantaneousAnomalyActive             bool    // indicates whether instantaneous anomaly spike is active in this time step
}

// Returns the change in signal caused by the instantaneous anomaly this timestep.
func (s *InstantaneousAnomaly) getInstantaneousDelta(r *rand.Rand) float64 {
	// No anomaly if probability is not met
	if r.Float64() > s.InstantaneousAnomalyProbability {
		s.InstantaneousAnomalyActive = false
		return 0.0
	}

	s.InstantaneousAnomalyActive = true
	if s.InstantaneousAnomalyMagnitudeVariation {
		return s.InstantaneousAnomalyMagnitude * r.NormFloat64()
	}
	return s.InstantaneousAnomalyMagnitude
}
