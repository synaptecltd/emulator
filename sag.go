package emulator

import (
	"math/rand"
	"time"
)

type SagEmulation struct {
	MeanStrain                float64
	MeanSag                   float64
	MeanCalculatedTemperature float64

	// outputs
	TotalStrain           float64
	Sag                   float64
	CalculatedTemperature float64
}

func (e *SagEmulation) stepSag(r *rand.Rand) {
	r.Seed(time.Now().UnixNano())
	e.TotalStrain = e.MeanStrain * r.Float64()
	e.Sag = e.MeanSag * r.Float64()
	e.CalculatedTemperature = e.MeanCalculatedTemperature * r.Float64()
}
