package emulator

import (
	"math"
	"math/rand"
)

type TemperatureEmulation struct {
	MeanTemperature float64 `yaml:"MeanTemperature"` // Mean temperature
	NoiseMax        float64 `yaml:"NoseMax"`         // Maximum noise
	ModulationMag   float64 `yaml:"ModulationMag"`   // Magnitude modulation

	Anomaly Anomaly `yaml:"Anomaly"` // Anomaly

	T float64 `yaml:"-"`
}

func (t *TemperatureEmulation) stepTemperature(r *rand.Rand, Ts float64) {
	varyingT := t.MeanTemperature * (1 + t.ModulationMag*math.Cos(1000.0*Ts))

	totalAnomalyDelta := t.Anomaly.stepAnomaly(r, Ts)

	t.T = varyingT + r.NormFloat64()*t.NoiseMax*t.MeanTemperature + totalAnomalyDelta
}
