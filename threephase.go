package emulator

import (
	"math"
	"math/rand/v2"

	"github.com/stevenblair/sigourney/fast"
	"github.com/synaptecltd/emulator/anomaly"
)

const TwoPiOverThree = 2 * math.Pi / 3

type ThreePhaseEmulation struct {
	// inputs
	PosSeqMag       float64   `yaml:"PosSeqMag,omitempty"`            // positive sequence magnitude
	PhaseOffset     float64   `yaml:"PhaseOffset,omitempty"`          // phase offset
	NegSeqMag       float64   `yaml:"NegSeqMag,omitempty"`            // negative sequence magnitude
	NegSeqAng       float64   `yaml:"NegSeqAng,omitempty"`            // negative sequence angle
	ZeroSeqMag      float64   `yaml:"ZeroSeqMag,omitempty"`           // zero sequence magnitude
	ZeroSeqAng      float64   `yaml:"ZeroSeqAng,omitempty"`           // zero sequence angle
	HarmonicNumbers []float64 `yaml:"HarmonicNumbers,flow,omitempty"` // harmonic numbers
	HarmonicMags    []float64 `yaml:"HarmonicMags,flow,omitempty"`    // harmonic magnitudes in pu, relative to PosSeqMag
	HarmonicAngs    []float64 `yaml:"HarmonicAngs,flow,omitempty"`    // harmonic angles
	NoiseMag        float64   `yaml:"NoiseMag,omitempty"`             // magnitude of Gaussian noise

	// define anomalies
	PosSeqMagAnomaly anomaly.Container `yaml:"PosSeqMagAnomaly,omitempty"` // positive sequence magnitude anomalies
	PosSeqAngAnomaly anomaly.Container `yaml:"PosSeqAngAnomaly,omitempty"` // positive sequence angle anomalies
	PhaseAMagAnomaly anomaly.Container `yaml:"PhaseAMagAnomaly,omitempty"` // phase A magnitude anomalies
	PhaseAAngAnomaly anomaly.Container `yaml:"PhaseAMagAnomaly,omitempty"` // phase A magnitude anomalies
	FreqAnomaly      anomaly.Container `yaml:"FreqAnomaly,omitempty"`      // frequency anomalies
	HarmonicsAnomaly anomaly.Container `yaml:"HarmonicsAnomaly,omitempty"` // harmonics anomalies

	// event emulation
	faultPhaseAMag        float64
	faultPosSeqMag        float64
	faultRemainingSamples int

	// internal state, state change
	pAngle            float64
	posSeqMagNew      float64
	posSeqMagRampRate float64

	// outputs
	A, B, C float64 `yaml:"-"`
	AMag    float64 `yaml:"-"`
	BMag    float64 `yaml:"-"`
	CMag    float64 `yaml:"-"`
	AAng    float64 `yaml:"-"`
	BAng    float64 `yaml:"-"`
	CAng    float64 `yaml:"-"`
}

// Steps the three phase emulation forward by one time step. The new values are
// defined based on magntiudes, noise values, anomalies and fault conditions.
func (e *ThreePhaseEmulation) stepThreePhase(r *rand.Rand, f float64, Ts float64) {
	// frequency anomaly
	totalAnomalyDeltaFrequency := e.FreqAnomaly.StepAll(r, Ts)
	freqTotal := f + totalAnomalyDeltaFrequency

	angle := (freqTotal*2*math.Pi*Ts + e.pAngle)
	angle = wrapAngle(angle)
	e.pAngle = angle

	// positive sequence angle anomaly
	totalAnomalyDeltaPosSeqAng := e.PosSeqAngAnomaly.StepAll(r, Ts)

	PosSeqPhase := e.PhaseOffset + e.pAngle + (math.Pi * totalAnomalyDeltaPosSeqAng / 180.0)

	if math.Abs(e.posSeqMagNew-e.PosSeqMag) >= math.Abs(e.posSeqMagRampRate) {
		e.PosSeqMag = e.PosSeqMag + e.posSeqMagRampRate
	}

	posSeqMag := e.PosSeqMag
	// phaseAMag := e.PosSeqMag
	if /*smpCnt > EmulatedFaultStartSamples && */ e.faultRemainingSamples > 0 {
		posSeqMag = posSeqMag + e.faultPosSeqMag
		e.faultRemainingSamples--
	}

	// positive sequence magnitude anomaly
	totalAnomalyDeltaPosSeqMag := e.PosSeqMagAnomaly.StepAll(r, Ts)
	posSeqMag += totalAnomalyDeltaPosSeqMag

	// phase A mag and ang anomalies
	anomalyPhaseAMag := e.PhaseAMagAnomaly.StepAll(r, Ts)
	anomalyPhaseAAng := e.PhaseAAngAnomaly.StepAll(r, Ts)

	// simplified per-phase mag and angle calculation
	e.AMag = posSeqMag + anomalyPhaseAMag
	e.BMag = posSeqMag
	e.CMag = posSeqMag
	e.AAng = e.PhaseOffset + anomalyPhaseAAng
	e.BAng = e.PhaseOffset - TwoPiOverThree
	e.CAng = e.PhaseOffset + TwoPiOverThree

	// positive sequence
	a1 := fast.Sin(PosSeqPhase) * (posSeqMag + anomalyPhaseAMag)
	b1 := fast.Sin(PosSeqPhase-TwoPiOverThree) * posSeqMag
	c1 := fast.Sin(PosSeqPhase+TwoPiOverThree) * posSeqMag

	// negative sequence
	a2 := fast.Sin(PosSeqPhase+e.NegSeqAng) * e.NegSeqMag * e.PosSeqMag
	b2 := fast.Sin(PosSeqPhase+TwoPiOverThree+e.NegSeqAng) * e.NegSeqMag * e.PosSeqMag
	c2 := fast.Sin(PosSeqPhase-TwoPiOverThree+e.NegSeqAng) * e.NegSeqMag * e.PosSeqMag

	// zero sequence
	abc0 := fast.Sin(PosSeqPhase+e.ZeroSeqAng) * e.ZeroSeqMag * e.PosSeqMag

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

				ah = ah + fast.Sin(n*(PosSeqPhase)+ang)*mag
				bh = bh + fast.Sin(n*(PosSeqPhase-TwoPiOverThree)+ang)*mag
				ch = ch + fast.Sin(n*(PosSeqPhase+TwoPiOverThree)+ang)*mag
			}
		}
	}

	harmonicsScale := e.HarmonicsAnomaly.StepAll(r, Ts)
	ah = ah * (1 + harmonicsScale)
	bh = bh * (1 + harmonicsScale)
	ch = ch * (1 + harmonicsScale)

	// add noise, ensure worst case where noise is uncorrelated across phases
	ra := r.NormFloat64() * e.NoiseMag * e.PosSeqMag
	rb := r.NormFloat64() * e.NoiseMag * e.PosSeqMag
	rc := r.NormFloat64() * e.NoiseMag * e.PosSeqMag

	// combine the output for each phase
	e.A = a1 + a2 + abc0 + ah + ra
	e.B = b1 + b2 + abc0 + bh + rb
	e.C = c1 + c2 + abc0 + ch + rc
}

// Wraps the angle a to the range -pi to pi
func wrapAngle(a float64) float64 {
	if a > math.Pi {
		return a - 2*math.Pi
	}
	return a
}
