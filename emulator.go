package emulator

import "math/rand/v2"

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

// Emulator encapsulates the waveform emulation of three-phase voltage, three-phase current, or temperature
type Emulator struct {
	// common inputs
	SamplingRate int     `yaml:"SamplingRate"` // The sampling rate of the emulator
	Ts           float64 `yaml:"Ts"`           // The time step or sampling period (=1/SamplingRate)
	Fnom         float64 `yaml:"Fnom"`         // Nominal frequency
	Fdeviation   float64 `yaml:"Fdeviation"`   // Frequency deviation

	V *ThreePhaseEmulation `yaml:"VoltageEmulator,omitempty"` // Voltage Emulator
	I *ThreePhaseEmulation `yaml:"CurrentEmulator,omitempty"` // Current Emulator

	T   *TemperatureEmulation `yaml:"TemperatureEmulator,omitempty"` // Temperature Emulation
	Sag *SagEmulation         `yaml:"SagEmulator,omitempty"`         // Sag Emulator

	// common state
	SmpCnt                     int `yaml:"-"`
	fDeviationRemainingSamples int `yaml:"-"`

	r *rand.Rand `yaml:"-"`
}

// StartEvent initiates an emulated event
func (e *Emulator) StartEvent(eventType int) {
	// fmt.Println("StartEvent()", eventType)

	switch eventType {
	case SinglePhaseFault:
		// TODO
		// e.I.FaultPosSeqMag = EmulatedFaultCurrentMagnitude
		// e.I.FaultRemainingSamples = MaxEmulatedFaultDurationSamples
		e.I.FaultPhaseAMag = e.I.PosSeqMag * 1.2 // EmulatedFaultCurrentMagnitude
		e.I.FaultRemainingSamples = MaxEmulatedFaultDurationSamples
		e.V.FaultPhaseAMag = e.V.PosSeqMag * -0.2
		e.V.FaultRemainingSamples = MaxEmulatedFaultDurationSamples
	case ThreePhaseFault:
		e.I.FaultPosSeqMag = e.I.PosSeqMag * 1.2 // EmulatedFaultCurrentMagnitude
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

// Returns a new Emulator instance with a given sampling rate and frequency.
// The emulator's random seed is initialized with a random value.
func NewEmulator(samplingRate int, frequency float64) *Emulator {
	emu := &Emulator{
		SamplingRate: samplingRate,
		Fnom:         frequency,
		Fdeviation:   0.0,
		Ts:           1 / float64(samplingRate),
	}

	emu.r = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))

	return emu
}

// Sets the random seed for the emulator. This is useful for
// generating identical random events across multiple runs.
func (e *Emulator) SetRandomSeed(seed uint64) {
	e.r = rand.New(rand.NewPCG(seed, seed))
}

// Step performs one iteration of the waveform generation for the given time step, Ts
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
