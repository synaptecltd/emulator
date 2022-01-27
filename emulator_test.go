package emulator_test

import (
	"math"
	"testing"

	"github.com/synaptecltd/emulator"
)

// benchmark emulator performance
func BenchmarkEmulator(b *testing.B) {
	emu := createEmulator(4000, 0)

	for i := 0; i < b.N; i++ {
		for j := 0; j < 4000; j++ {
			emu.Step()
		}
	}
}

func createEmulator(samplingRate int, phaseOffsetDeg float64) *emulator.Emulator {
	emu := emulator.NewEmulator(samplingRate, 50.0)

	emu.V = &emulator.ThreePhaseEmulation{
		PosSeqMag:   400000.0 / math.Sqrt(3) * math.Sqrt(2),
		NoiseMax:    0.000001,
		PhaseOffset: phaseOffsetDeg * math.Pi / 180.0,
	}
	emu.I = &emulator.ThreePhaseEmulation{
		PosSeqMag:       500.0,
		PhaseOffset:     phaseOffsetDeg * math.Pi / 180.0,
		HarmonicNumbers: []float64{5, 7, 11, 13, 17, 19, 23, 25},
		HarmonicMags:    []float64{0.2164, 0.1242, 0.0892, 0.0693, 0.0541, 0.0458, 0.0370, 0.0332},
		HarmonicAngs:    []float64{171.5, 100.4, -52.4, 128.3, 80.0, 2.9, -146.8, 133.9},
		NoiseMax:        0.000001,
	}

	return emu
}
