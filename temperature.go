package emulator

import (
	"math/rand/v2"
)

type TemperatureEmulation struct {
	MeanTemperature float64          `yaml:"MeanTemperature"` // mean temperature
	NoiseMax        float64          `yaml:"NoiseMax"`        // magnitude of Gaussian noise
	Anomaly         AnomalyContainer `yaml:"Anomaly"`         // anomalies
	T               float64          `yaml:"-"`               // present value of temperature
}

// Steps the temperature emulation forward by one time step. The new temperature is
// calculated as the mean temperature + Gaussian noise + anomalies (if present).
func (t *TemperatureEmulation) stepTemperature(r *rand.Rand, Ts float64) {
	t.T = t.MeanTemperature + r.NormFloat64()*t.NoiseMax*t.MeanTemperature

	anomalyValues := stepAllAnomalies(t.Anomaly, r, Ts)
	t.T += anomalyValues
}
