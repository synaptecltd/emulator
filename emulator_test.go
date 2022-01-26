package emulator

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func FloatingPointEqual(expected float64, actual float64, threshold float64) bool {
	absDiff := math.Abs(expected - actual)
	if absDiff < threshold {
		return true
	}
	return false
}

func createEmulator(samplingRate int) *Emulator {
	return &Emulator{
		SamplingRate: samplingRate,
		Fnom:         50.03,
		Fdeviation:   0.0,
		Ts:           1 / float64(samplingRate),
		V: &ThreePhaseEmulation{
			PosSeqMag: 400000.0 / math.Sqrt(3) * math.Sqrt(2),
			NoiseMax:  0.000001,
		},
		I: &ThreePhaseEmulation{
			PosSeqMag:       500.0,
			HarmonicNumbers: []float64{5, 7, 11, 13, 17, 19, 23, 25},
			HarmonicMags:    []float64{0.2164, 0.1242, 0.0892, 0.0693, 0.0541, 0.0458, 0.0370, 0.0332},
			HarmonicAngs:    []float64{171.5, 100.4, -52.4, 128.3, 80.0, 2.9, -146.8, 133.9},
			NoiseMax:        0.00001,
		},
		T: &TemperatureEmulation{
			MeanTemperature: 30.0,
			NoiseMax: 0.01,
			AnomalyMagnitude: 30,
			AnomalyProbability: 0.01,
		},
	}
}

func TestTemperatureEmulatorAnomalies_NoAmomalies(t *testing.T) {
	emulator := createEmulator(14400)

	emulator.T.AnomalyProbability = 0
	step := 0
	var results []bool
	for step < 1E4 {
		emulator.Step()
		results = append(results, emulator.T.IsAnomaly)
		step += 1
	}
	assert.NotContains(t, results, true)
}

func TestTemperatureEmulatorAnomalies_Amomalies(t *testing.T) {
	emulator := createEmulator(14400)

	emulator.T.AnomalyProbability = 0.5
	step := 0
	var results []bool
	var normalValues []float64
	var anomalyValues []float64
	for step < 1E4 {
		emulator.Step()
		results = append(results, emulator.T.IsAnomaly)

		if emulator.T.IsAnomaly == true {
			anomalyValues = append(anomalyValues, emulator.T.T)
		} else {
			normalValues = append(normalValues, emulator.T.T)
		}
		step += 1
	}
	assert.Contains(t, results, true)
	fractionAnomalies := float64(len(anomalyValues))/float64(step)
	assert.True(t, FloatingPointEqual(0.5, fractionAnomalies, 0.1))
}