package emulator

import (
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/synaptecltd/emulator/anomaly"
)

type TemperatureEmulation struct {
	MeanTemperature float64           `yaml:"MeanTemperature"` // mean temperature
	NoiseMag        float64           `yaml:"NoiseMag"`        // magnitude of Gaussian noise
	Anomaly         anomaly.Container `yaml:"Anomaly"`         // anomalies
	T               float64           `yaml:"-"`               // present value of temperature
}

// Steps the temperature emulation forward by one time step. The new temperature is
// calculated as the mean temperature + Gaussian noise + anomalies (if present).
func (t *TemperatureEmulation) stepTemperature(r *rand.Rand, Ts float64) {
	t.T = t.MeanTemperature + r.NormFloat64()*t.NoiseMag*t.MeanTemperature

	anomalyValues := t.Anomaly.StepAll(r, Ts)
	t.T += anomalyValues
}

// Add an anomaly to the temperature emulation, returning the UUID of the added anomaly.
func (t *TemperatureEmulation) AddAnomaly(anom anomaly.AnomalyInterface) uuid.UUID {
	return t.Anomaly.AddAnomaly(anom)
}
