package emulator

import (
	"math/rand"
	"time"
)

type SagEmulation struct {
	MeanStrain                float64 `yaml:"MeanStrain"`                // Mean strain
	MeanSag                   float64 `yaml:"MeanSag"`                   // Mean sag
	MeanCalculatedTemperature float64 `yaml:"MeanCalculatedTemperature"` // Mean calculated temperature

	// outputs
	TotalStrain           float64 `yaml:"-"` // Total strain
	Sag                   float64 `yaml:"-"` // Sag
	CalculatedTemperature float64 `yaml:"-"` // Calculated temperature
}

func (e *SagEmulation) stepSag(r *rand.Rand) {
	r.Seed(time.Now().UnixNano())
	e.TotalStrain = e.MeanStrain * r.Float64()
	e.Sag = e.MeanSag * r.Float64()
	e.CalculatedTemperature = e.MeanCalculatedTemperature * r.Float64()
}
