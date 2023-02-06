package emulator

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// benchmark emulator performance
func BenchmarkEmulator(b *testing.B) {
	emu := createEmulatorForBenchmark(4000, 0)

	for i := 0; i < b.N; i++ {
		for j := 0; j < 4000; j++ {
			emu.Step()
		}
	}
}

func createEmulatorForBenchmark(samplingRate int, phaseOffsetDeg float64) *Emulator {
	emu := NewEmulator(samplingRate, 50.0)

	emu.V = &ThreePhaseEmulation{
		PosSeqMag:   400000.0 / math.Sqrt(3) * math.Sqrt(2),
		NoiseMax:    0.000001,
		PhaseOffset: phaseOffsetDeg * math.Pi / 180.0,
	}
	emu.I = &ThreePhaseEmulation{
		PosSeqMag:       500.0,
		PhaseOffset:     phaseOffsetDeg * math.Pi / 180.0,
		HarmonicNumbers: []float64{5, 7, 11, 13, 17, 19, 23, 25},
		HarmonicMags:    []float64{0.2164, 0.1242, 0.0892, 0.0693, 0.0541, 0.0458, 0.0370, 0.0332},
		HarmonicAngs:    []float64{171.5, 100.4, -52.4, 128.3, 80.0, 2.9, -146.8, 133.9},
		NoiseMax:        0.000001,
	}

	return emu
}

func createEmulator(samplingRate int, phaseOffsetDeg float64) *Emulator {
	emu := NewEmulator(samplingRate, 50.0)

	emu.V = &ThreePhaseEmulation{
		PosSeqMag:   400000.0 / math.Sqrt(3) * math.Sqrt(2),
		NoiseMax:    0.000001,
		PhaseOffset: phaseOffsetDeg * math.Pi / 180.0,
	}
	emu.I = &ThreePhaseEmulation{
		PosSeqMag:       500.0,
		PhaseOffset:     phaseOffsetDeg * math.Pi / 180.0,
		HarmonicNumbers: []float64{5, 7, 11, 13, 17, 19, 23, 25},
		HarmonicMags:    []float64{0.2164, 0.1242, 0.0892, 0.0693, 0.0541, 0.0458, 0.0370, 0.0332},
		HarmonicAngs:    []float64{171.5, 100.4, -52.4, 128.3, 80.0, 2.9, -146.8, 133.9},
		NoiseMax:        0.000001,
	}
	emu.T = &TemperatureEmulation{
		MeanTemperature: 30.0,
		NoiseMax:        0.01,
		Anomaly: Anomaly{
			InstantaneousAnomalyMagnitude:   30,
			InstantaneousAnomalyProbability: 0.01,
		},
	}
	return emu
}

func createEmulatorForAnomaly(samplingRate int, phaseOffsetDeg float64) *Emulator {
	emu := NewEmulator(samplingRate, 50.0)

	emu.I = &ThreePhaseEmulation{
		PosSeqMag:   350.0,
		PhaseOffset: 0.0,
		PosSeqMagAnomaly: Anomaly{
			IsTrendAnomaly:        true,
			IsRisingTrendAnomaly:  true,
			TrendAnomalyDuration:  10,
			TrendAnomalyMagnitude: 1060.32,
		},
	}

	return emu
}

func FloatingPointEqual(expected float64, actual float64, threshold float64) bool {
	absDiff := math.Abs(expected - actual)
	if absDiff < threshold {
		return true
	}
	return false
}

func mean(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func TestTemperatureEmulationAnomalies_NoAnomalies(t *testing.T) {
	emulator := createEmulator(14400, 0)

	emulator.T.Anomaly.InstantaneousAnomalyProbability = 0
	step := 0
	var results []bool
	for step < 1e4 {
		emulator.Step()
		results = append(results, emulator.T.Anomaly.isInstantaneousAnomaly)
		step += 1
	}
	assert.NotContains(t, results, true)
}

func TestTemperatureEmulationAnomalies_Anomalies(t *testing.T) {
	emulator := createEmulator(14400, 0)

	emulator.T.Anomaly.InstantaneousAnomalyProbability = 0.5
	step := 0
	var results []bool
	var normalValues []float64
	var anomalyValues []float64
	for step < 1e4 {
		emulator.Step()
		results = append(results, emulator.T.Anomaly.isInstantaneousAnomaly)

		if emulator.T.Anomaly.isInstantaneousAnomaly == true {
			anomalyValues = append(anomalyValues, emulator.T.T)
		} else {
			normalValues = append(normalValues, emulator.T.T)
		}
		step += 1
	}
	assert.Contains(t, results, true)

	fractionAnomalies := float64(len(anomalyValues)) / float64(step)
	assert.True(t, FloatingPointEqual(0.5, fractionAnomalies, 0.1))

	assert.True(t, mean(anomalyValues) > mean(normalValues))
}

func TestTemperatureEmulationAnomalies_RisingTrend(t *testing.T) {
	emulator := createEmulator(14400, 0)
	emulator.T.Anomaly.IsTrendAnomaly = true
	emulator.T.Anomaly.TrendAnomalyMagnitude = 30.0
	emulator.T.Anomaly.TrendAnomalyDuration = 10
	emulator.T.Anomaly.IsRisingTrendAnomaly = true

	step := 0.0
	var results []float64
	for step < emulator.T.Anomaly.TrendAnomalyDuration*float64(emulator.SamplingRate) {
		emulator.Step()
		results = append(results, emulator.T.T)
		step += 1

		if step < float64(emulator.SamplingRate) {
			assert.NotEqual(t, 0, emulator.T.Anomaly.TrendAnomalyIndex)
		}
	}

	assert.True(t, mean(results) > emulator.T.MeanTemperature)
}

func TestTemperatureEmulationAnomalies_DecreasingTrend(t *testing.T) {
	emulator := createEmulator(1, 0)
	emulator.T.Anomaly.IsTrendAnomaly = true
	emulator.T.Anomaly.TrendAnomalyMagnitude = 30.0
	emulator.T.Anomaly.TrendAnomalyDuration = 10
	emulator.T.Anomaly.IsRisingTrendAnomaly = false
	step := 0
	var results []float64
	for step < 10 {
		emulator.Step()
		results = append(results, emulator.T.T)
		step += 1
	}

	assert.True(t, mean(results) < emulator.T.MeanTemperature)
}

func TestCurrentPosSeqAnomalies_RisingTrend(t *testing.T) {
	emulator := createEmulatorForAnomaly(4000, 0)

	step := 0.0
	var results []float64
	for step < emulator.I.PosSeqMagAnomaly.TrendAnomalyDuration*float64(emulator.SamplingRate) {
		emulator.Step()
		results = append(results, emulator.I.A)
		step += 1

		if step < float64(emulator.SamplingRate) {
			assert.NotEqual(t, 0, emulator.I.PosSeqMagAnomaly.TrendAnomalyIndex)
		}
	}

	// find maxMag value of array
	maxMag := 0.0
	for _, value := range results {
		if value > maxMag {
			maxMag = value
		}
	}
	targetMag := emulator.I.PosSeqMag + emulator.I.PosSeqMagAnomaly.TrendAnomalyMagnitude

	assert.True(t, FloatingPointEqual(targetMag, maxMag, 50))
}

func TestSagEmulation(t *testing.T) {
	emulator := createEmulator(14400, 0)
	emulator.Sag = &SagEmulation{
		MeanCalculatedTemperature: 30.0,
		MeanStrain:                100.0,
		MeanSag:                   0.5,
	}

	step := 0
	results := make(map[string][]float64)
	results["CalculatedTemperature"] = make([]float64, 0)
	results["TotalStrain"] = make([]float64, 0)
	results["Sag"] = make([]float64, 0)

	for step < 1e4 {
		emulator.Step()
		results["CalculatedTemperature"] = append(results["CalculatedTemperature"], emulator.Sag.CalculatedTemperature)
		results["TotalStrain"] = append(results["TotalStrain"], emulator.Sag.CalculatedTemperature)
		results["Sag"] = append(results["Sag"], emulator.Sag.CalculatedTemperature)
		step += 1
	}

	for _, field := range []string{"CalculatedTemperature", "TotalStrain", "Sag"} {
		assert.IsType(t, []float64{}, results[field])
	}
}
