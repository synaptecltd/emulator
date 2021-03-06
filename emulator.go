package emulator

import (
	"math"
	"math/rand"
	"time"
)

// Emulated event types
const (
	SinglePhaseFault     = iota
	ThreePhaseFault      = iota
	OverVoltage          = iota
	UnderVoltage         = iota
	OverFrequency        = iota
	UnderFrequency       = iota
	CapacitorOverCurrent = iota
)

// EmulatedFaultStartSamples is the number of samples before initiating an emulated fault
const EmulatedFaultStartSamples = 1000

// MaxEmulatedFaultDurationSamples is the number of samples for emulating a fault
const MaxEmulatedFaultDurationSamples = 6000

// MaxEmulatedCapacitorOverCurrentSamples is the number of samples for emulating a fault
const MaxEmulatedCapacitorOverCurrentSamples = 8000

// MaxEmulatedFrequencyDurationSamples is the number of samples for emulating frequency deviations
const MaxEmulatedFrequencyDurationSamples = 8000

// EmulatedFaultCurrentMagnitude is the additional fault current magnitude added to one circuit end
const EmulatedFaultCurrentMagnitude = 80

const TwoPiOverThree = 2 * math.Pi / 3

type ThreePhaseEmulation struct {
	// inputs
	PosSeqMag       float64
	PhaseOffset     float64
	NegSeqMag       float64
	NegSeqAng       float64
	ZeroSeqMag      float64
	ZeroSeqAng      float64
	HarmonicNumbers []float64 `mapstructure:",omitempty,flow"`
	HarmonicMags    []float64 `mapstructure:",omitempty,flow"` // pu, relative to PosSeqMag
	HarmonicAngs    []float64 `mapstructure:",omitempty,flow"`
	NoiseMax        float64

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

type TemperatureEmulation struct {
	MeanTemperature float64
	NoiseMax        float64
	ModulationMag   float64

	// instantaneous anomalies
	isInstantaneousAnomaly          bool // private
	InstantaneousAnomalyProbability float64
	InstantaneousAnomalyMagnitude   float64

	// trend anomalies
	IsTrendAnomaly        bool
	TrendAnomalyLength    int
	TrendAnomalyIndex     int
	TrendAnomalyMagnitude float64

	T float64
}

type SagEmulation struct {
	MeanStrain                float64
	MeanSag                   float64
	MeanCalculatedTemperature float64

	// outputs
	TotalStrain           float64
	Sag                   float64
	CalculatedTemperature float64
}

// Emulator encapsulates the waveform emulation of three-phase voltage, three-phase current, or temperature
type Emulator struct {
	// common inputs
	SamplingRate int
	Ts           float64
	Fnom         float64
	Fdeviation   float64

	V *ThreePhaseEmulation
	I *ThreePhaseEmulation

	T   *TemperatureEmulation
	Sag *SagEmulation

	// common state
	SmpCnt                     int
	fDeviationRemainingSamples int

	r *rand.Rand
}

func wrapAngle(a float64) float64 {
	if a > math.Pi {
		return a - 2*math.Pi
	}
	return a
}

// StartEvent initiates an emulated event
func (e *Emulator) StartEvent(eventType int) {
	// fmt.Println("StartEvent()", eventType)

	switch eventType {
	case SinglePhaseFault:
		// TODO
		// e.I.FaultPosSeqMag = EmulatedFaultCurrentMagnitude
		// e.I.FaultRemainingSamples = MaxEmulatedFaultDurationSamples
		e.I.FaultPhaseAMag = e.I.PosSeqMag * 1.2 //EmulatedFaultCurrentMagnitude
		e.I.FaultRemainingSamples = MaxEmulatedFaultDurationSamples
		e.V.FaultPhaseAMag = e.V.PosSeqMag * -0.2
		e.V.FaultRemainingSamples = MaxEmulatedFaultDurationSamples
	case ThreePhaseFault:
		e.I.FaultPosSeqMag = e.I.PosSeqMag * 1.2 //EmulatedFaultCurrentMagnitude
		e.I.FaultRemainingSamples = MaxEmulatedFaultDurationSamples
		e.V.FaultPosSeqMag = e.V.PosSeqMag * -0.2
		e.V.FaultRemainingSamples = MaxEmulatedFaultDurationSamples
	case OverVoltage:
		e.V.FaultPosSeqMag = e.V.PosSeqMag * 0.2
		e.V.FaultRemainingSamples = MaxEmulatedFaultDurationSamples
	case UnderVoltage:
		e.V.FaultPosSeqMag = e.V.PosSeqMag * -0.2
		e.V.FaultRemainingSamples = MaxEmulatedFaultDurationSamples
	case OverFrequency:
		e.Fdeviation = 0.1
		e.fDeviationRemainingSamples = MaxEmulatedFrequencyDurationSamples
	case UnderFrequency:
		e.Fdeviation = -0.1
		e.fDeviationRemainingSamples = MaxEmulatedFrequencyDurationSamples
	case CapacitorOverCurrent:
		// TODO
		e.I.FaultPosSeqMag = e.I.PosSeqMag * 0.01
		e.I.FaultRemainingSamples = MaxEmulatedCapacitorOverCurrentSamples
	default:
	}
}

func NewEmulator(samplingRate int, frequency float64) *Emulator {
	emu := &Emulator{
		SamplingRate: samplingRate,
		Fnom:         frequency,
		Fdeviation:   0.0,
		Ts:           1 / float64(samplingRate),
	}

	emu.r = rand.New(rand.NewSource(time.Now().Unix()))

	return emu
}

// Step performs one iteration of the waveform generation
func (e *Emulator) Step() {
	f := e.Fnom + e.Fdeviation

	if e.fDeviationRemainingSamples > 0 {
		e.fDeviationRemainingSamples--
		if e.fDeviationRemainingSamples == 0 {
			e.Fdeviation = 0.0
		}
	}

	if e.V != nil {
		e.V.stepThreePhase(e.r, f, e.Ts, e.SmpCnt)
	}
	if e.I != nil {
		e.I.stepThreePhase(e.r, f, e.Ts, e.SmpCnt)
	}
	if e.T != nil {
		e.T.stepTemperature(e.r, e.Ts)
	}
	if e.Sag != nil {
		e.Sag.stepSag(e.r)
	}

	e.SmpCnt++
	if int(e.SmpCnt) >= e.SamplingRate {
		e.SmpCnt = 0
	}
}

func (t *TemperatureEmulation) stepTemperature(r *rand.Rand, Ts float64) {
	varyingT := t.MeanTemperature * (1 + t.ModulationMag*math.Cos(1000.0*Ts))

	trendAnomalyDelta := 0.0
	trendAnomalyStep := t.TrendAnomalyMagnitude / float64(t.TrendAnomalyLength)
	if t.IsTrendAnomaly == true {
		trendAnomalyDelta = float64(t.TrendAnomalyIndex) * trendAnomalyStep
		if t.TrendAnomalyIndex == t.TrendAnomalyLength-1 {
			t.TrendAnomalyIndex = 0
		} else {
			t.TrendAnomalyIndex += 1
		}
	}

	instantaneousAnomalyDelta := 0.0
	t.isInstantaneousAnomaly = false
	if t.InstantaneousAnomalyProbability > rand.Float64() {
		instantaneousAnomalyDelta = t.InstantaneousAnomalyMagnitude
		t.isInstantaneousAnomaly = true
	}

	totalAnomalyDelta := trendAnomalyDelta + instantaneousAnomalyDelta

	t.T = varyingT + r.NormFloat64()*t.NoiseMax*t.MeanTemperature + totalAnomalyDelta
}

func (e *ThreePhaseEmulation) stepThreePhase(r *rand.Rand, f float64, Ts float64, smpCnt int) {
	angle := (f*2*math.Pi*Ts + e.pAngle)
	angle = wrapAngle(angle)
	e.pAngle = angle

	PosSeqPhase := e.PhaseOffset + e.pAngle

	if math.Abs(e.PosSeqMagNew-e.PosSeqMag) >= math.Abs(e.PosSeqMagRampRate) {
		e.PosSeqMag = e.PosSeqMag + e.PosSeqMagRampRate
	}

	posSeqMag := e.PosSeqMag
	// phaseAMag := e.PosSeqMag
	if /*smpCnt > EmulatedFaultStartSamples && */ e.FaultRemainingSamples > 0 {
		posSeqMag = posSeqMag + e.FaultPosSeqMag
		e.FaultRemainingSamples--
	}

	// positive sequence
	a1 := math.Sin(PosSeqPhase) * posSeqMag
	b1 := math.Sin(PosSeqPhase-TwoPiOverThree) * posSeqMag
	c1 := math.Sin(PosSeqPhase+TwoPiOverThree) * posSeqMag

	// negative sequence
	a2 := math.Sin(PosSeqPhase+e.NegSeqAng) * e.NegSeqMag * e.PosSeqMag
	b2 := math.Sin(PosSeqPhase+TwoPiOverThree+e.NegSeqAng) * e.NegSeqMag * e.PosSeqMag
	c2 := math.Sin(PosSeqPhase-TwoPiOverThree+e.NegSeqAng) * e.NegSeqMag * e.PosSeqMag

	// zero sequence
	abc0 := math.Sin(PosSeqPhase+e.ZeroSeqAng) * e.ZeroSeqMag

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

	// add noise, ensure worst case where noise is uncorrelated across phases
	ra := r.NormFloat64() * e.NoiseMax * e.PosSeqMag
	rb := r.NormFloat64() * e.NoiseMax * e.PosSeqMag
	rc := r.NormFloat64() * e.NoiseMax * e.PosSeqMag

	// combine the output for each phase
	e.A = a1 + a2 + abc0 + ah + ra
	e.B = b1 + b2 + abc0 + bh + rb
	e.C = c1 + c2 + abc0 + ch + rc
}

func (e *SagEmulation) stepSag(r *rand.Rand) {
	r.Seed(time.Now().UnixNano())
	e.TotalStrain = e.MeanStrain * r.Float64()
	e.Sag = e.MeanSag * r.Float64()
	e.CalculatedTemperature = e.MeanCalculatedTemperature * r.Float64()
}
