# Emulator - generating high-resolution sensor data

This Go module emulates data for voltage, current, and temperature sensors.

For voltage and current sensors, it allows typical parameters for three-phase systems to be specified, and it outputs waveform samples. It supports arbitrary sampling rates and other signal parameters.

"Anomalies" can be superimposed on the generated data to create abnormal conditions for testing alarms or other scenarios.

## Code example

Each `emulator` instance can implement up to one of each of: three-phase voltage (`V`), three-phase current (`I`), and temperature (`T`). Outputs values are updated every time step, `Ts`, for a given sampling rate. Only the outputs for initialised `V`, `I`, and `T` objects will be updated each time step.

```go
// base parameters
samplingRate := 14400
freq := 50.0
phaseOffsetDeg := 0.0

// create an emulator instance
emu := emulator.NewEmulator(samplingRate, freq)

// specify three-phase voltage parameters
emu.V = &emulator.ThreePhaseEmulation{
    PosSeqMag:   400000.0 / math.Sqrt(3) * math.Sqrt(2),
    NoiseMax:    0.000001,
    PhaseOffset: phaseOffsetDeg * math.Pi / 180.0,
}

// specify three-phase current parameters
emu.I = &emulator.ThreePhaseEmulation{
    PosSeqMag:       500.0,
    PhaseOffset:     phaseOffsetDeg * math.Pi / 180.0,
    HarmonicNumbers: []float64{5, 7, 11, 13, 17, 19, 23, 25},
    HarmonicMags:    []float64{0.2164, 0.1242, 0.0892, 0.0693, 0.0541, 0.0458, 0.0370, 0.0332},
    HarmonicAngs:    []float64{171.5, 100.4, -52.4, 128.3, 80.0, 2.9, -146.8, 133.9},
    NoiseMax:        0.000001,
}

// specify temperature parameters
emu.T = &emulator.TemperatureEmulation{
    MeanTemperature: 30.0,
    NoiseMax:        0.01,
    Anomaly: emulator.Anomaly{
        InstantaneousAnomalyMagnitude:   30,
        InstantaneousAnomalyProbability: 0.01,
    },
}

// execute one full waveform period of samples using the Step() function
step := 0
var results []float64
for step < samplingRate {
    emu.Step()
    results = append(results, emu.T.T)
    step += 1
}
```

## Anomalies

Two types of "anomaly" can be added to the data to create interesting scenarios:
1. Instantaneous: based on a probability factor, activate an instantaneous change to the selected parameter
2. Periodic "trends": apply a sawtooth shape to the parameter

The parameter `TrendAnomalyMagnitude` has the following effects:

| Sensor type     | Name of item       | Modulated parameter         | Effect                                         | Units         |
| --------------- | ------------------ | --------------------------- | ---------------------------------------------- | ------------- |
| Voltage/current | `PosSeqMagAnomaly` | Positive sequence magnitude | Adds/subtracts positive sequence magnitude     | Volts or Amps |
| Voltage/current | `PosSeqAngAnomaly` | Positive sequence angle     | Adds/subtracts positive sequence angle         | Degrees       |
| Voltage/current | `PhaseAMagAnomaly` | Phase A magnitude           | Adds/subtracts phase A magnitude               | Volts or Amps |
| Voltage/current | `FreqAnomaly`      | Frequency                   | Adds/subtracts signal frequency                | Hz            |
| Voltage/current | `HarmonicsAnomaly` | All harmonics magnitudes    | Adds/subtracts all harmonic magnitudes         | per unit      |
| Temperature     | `Anomaly`          | Temperature value           | Adds/subtracts instantaneous temperature value | Degrees C     |
