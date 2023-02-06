package emulator

import (
	"math"
	"math/rand"
)

const TwoPiOverThree = 2 * math.Pi / 3

type ThreePhaseEmulation struct {
	// inputs
	PosSeqMag       float64
	PhaseOffset     float64
	NegSeqMag       float64
	NegSeqAng       float64
	ZeroSeqMag      float64
	ZeroSeqAng      float64
	HarmonicNumbers []float64
	HarmonicMags    []float64 // pu, relative to PosSeqMag
	HarmonicAngs    []float64
	NoiseMax        float64

	// define anomalies
	PosSeqMagAnomaly Anomaly // positive sequence magnitude anomaly
	PosSeqAngAnomaly Anomaly // positive sequence angle anomaly
	PhaseAMagAnomaly Anomaly // phase A magnitude anomaly
	FreqAnomaly      Anomaly // frequency anomaly
	HarmonicsAnomaly Anomaly // harmonics magnitude anomaly

	// event emulation
	FaultPhaseAMag        float64
	FaultPosSeqMag        float64
	FaultRemainingSamples int

	// state change
	PosSeqMagNew      float64
	PosSeqMagRampRate float64

	// internal state
	pAngle float64

	// outputs
	A, B, C float64
}

func (e *ThreePhaseEmulation) stepThreePhase(r *rand.Rand, f float64, Ts float64, smpCnt int) {
	// frequency anomaly
	totalAnomalyDeltaFrequency := e.FreqAnomaly.stepAnomaly(r, Ts)
	freqTotal := f + totalAnomalyDeltaFrequency

	angle := (freqTotal*2*math.Pi*Ts + e.pAngle)
	angle = wrapAngle(angle)
	e.pAngle = angle

	// positive sequence angle anomaly
	totalAnomalyDeltaPosSeqAng := e.PosSeqAngAnomaly.stepAnomaly(r, Ts)

	PosSeqPhase := e.PhaseOffset + e.pAngle + (math.Pi * totalAnomalyDeltaPosSeqAng / 180.0)

	if math.Abs(e.PosSeqMagNew-e.PosSeqMag) >= math.Abs(e.PosSeqMagRampRate) {
		e.PosSeqMag = e.PosSeqMag + e.PosSeqMagRampRate
	}

	posSeqMag := e.PosSeqMag
	// phaseAMag := e.PosSeqMag
	if /*smpCnt > EmulatedFaultStartSamples && */ e.FaultRemainingSamples > 0 {
		posSeqMag = posSeqMag + e.FaultPosSeqMag
		e.FaultRemainingSamples--
	}

	// positive sequence magnitude anomaly
	totalAnomalyDeltaPosSeqMag := e.PosSeqMagAnomaly.stepAnomaly(r, Ts)
	posSeqMag += totalAnomalyDeltaPosSeqMag

	// phase A magnitude anomaly
	anomalyPhaseA := e.PhaseAMagAnomaly.stepAnomaly(r, Ts)

	// positive sequence
	a1 := math.Sin(PosSeqPhase) * (posSeqMag + anomalyPhaseA)
	b1 := math.Sin(PosSeqPhase-TwoPiOverThree) * posSeqMag
	c1 := math.Sin(PosSeqPhase+TwoPiOverThree) * posSeqMag

	// negative sequence
	a2 := math.Sin(PosSeqPhase+e.NegSeqAng) * e.NegSeqMag * e.PosSeqMag
	b2 := math.Sin(PosSeqPhase+TwoPiOverThree+e.NegSeqAng) * e.NegSeqMag * e.PosSeqMag
	c2 := math.Sin(PosSeqPhase-TwoPiOverThree+e.NegSeqAng) * e.NegSeqMag * e.PosSeqMag

	// zero sequence
	abc0 := math.Sin(PosSeqPhase+e.ZeroSeqAng) * e.ZeroSeqMag * e.PosSeqMag

	// harmonics
	ah := 0.0
	bh := 0.0
	ch := 0.0
	if len(e.HarmonicNumbers) > 0 {
		// ensure consistent array sizes have been specified
		if len(e.HarmonicNumbers) == len(e.HarmonicMags) && len(e.HarmonicNumbers) == len(e.HarmonicAngs) {
			for i, n := range e.HarmonicNumbers {
				mag := e.HarmonicMags[i] * e.PosSeqMag
				ang := e.HarmonicAngs[i] // / 180.0 * math.Pi

				ah = ah + math.Sin(n*(PosSeqPhase)+ang)*mag
				bh = bh + math.Sin(n*(PosSeqPhase-TwoPiOverThree)+ang)*mag
				ch = ch + math.Sin(n*(PosSeqPhase+TwoPiOverThree)+ang)*mag
			}
		}
	}

	harmonicsScale := e.HarmonicsAnomaly.stepAnomaly(r, Ts)
	ah = ah * (1 + harmonicsScale)
	bh = bh * (1 + harmonicsScale)
	ch = ch * (1 + harmonicsScale)

	// add noise, ensure worst case where noise is uncorrelated across phases
	ra := r.NormFloat64() * e.NoiseMax * e.PosSeqMag
	rb := r.NormFloat64() * e.NoiseMax * e.PosSeqMag
	rc := r.NormFloat64() * e.NoiseMax * e.PosSeqMag

	// combine the output for each phase
	e.A = a1 + a2 + abc0 + ah + ra
	e.B = b1 + b2 + abc0 + bh + rb
	e.C = c1 + c2 + abc0 + ch + rc
}

func wrapAngle(a float64) float64 {
	if a > math.Pi {
		return a - 2*math.Pi
	}
	return a
}
