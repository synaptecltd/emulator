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

// define some anomalies
// spikes of magnitude +/- 30.0, triggering with probability 1% at each time step
spikes, _ := anomaly.NewSpikeAnomaly(anomaly.SpikeParams{
    Probability: 0.01,
    Magnitude:   30.0,
})

// a repeating linear ramp
ramp, _ := anomaly.NewTrendAnomaly(
    anomaly.TrendParams{
        Magnitude:   5, // ramp magnitude
        Duration:    0.7, // ramp duration, seconds
        MagFuncName: "linear",
    },
)

// create an emulator instance
emu := emulator.NewEmulator(samplingRate, freq)

// specify three-phase voltage parameters
emu.V = &emulator.ThreePhaseEmulation{
    PosSeqMag:   400000.0 / math.Sqrt(3) * math.Sqrt(2),
    NoiseMag:    0.000001,
    PhaseOffset: phaseOffsetDeg * math.Pi / 180.0,
}

// specify three-phase current parameters, add the spike anomaly
emu.I = &emulator.ThreePhaseEmulation{
    PosSeqMag:       500.0,
    PhaseOffset:     phaseOffsetDeg * math.Pi / 180.0,
    HarmonicNumbers: []float64{5, 7, 11, 13, 17, 19, 23, 25},
    HarmonicMags:    []float64{0.2164, 0.1242, 0.0892, 0.0693, 0.0541, 0.0458, 0.0370, 0.0332},
    HarmonicAngs:    []float64{171.5, 100.4, -52.4, 128.3, 80.0, 2.9, -146.8, 133.9},
    NoiseMag:        0.000001,
    PhaseAMagAnomaly: anomaly.Container{
        "events": spikes,
    },
}

// Create an anomaly container for temperature and add anomalies
container := anomaly.Container{}
spikes.Magnitude = 1.0 // re-use an anomaly with reduced magnitude
_ = container.AddAnomaly(spikes) // returns uuid of anomaly
_ = container.AddAnomaly(ramp)

// Specify tempertaure parameters
emu.T = &emulator.TemperatureEmulation{
    MeanTemperature: 30.0,
    NoiseMag:        0.01,
    Anomaly:         container,
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

Alternatively, emulators can be define via yaml:

```go
fileBytes, _ := os.ReadFile("foo.yml")
emu.T = &emulator.TemperatureEmulation{}
yaml.Unmarshal(fileBytes, emu.T)
```

where `foo.yml` is e.g.:

```yaml
MeanTemperature: 20.0
NoiseMag: 0.1
Anomaly:
  repeating_ramp:   # anomaly name
    type: trend     # type of anomaly: trend
    magnitude: 5    # params
    duration: 0.7
  blips:            # anomaly name
    type: spike     # type of anomaly: spike
    probability: 0.01
    magnitude: 2
  # etc
```

## Anomalies

Two types of "anomaly" can be added to the data to create interesting scenarios:
1. Instantaneous: based on a probability factor, activate an instantaneous change to the selected parameter
2. Periodic "trends": apply a continuous changes to the parameter, including ramps, sinusoids, and additional noise. See `./mathfuncs` for a full list.

The parameter `TrendAnomalyMagnitude` has the following effects:

| Sensor type     | Name of item       | Modulated parameter         | Effect                                         | Units         |
| --------------- | ------------------ | --------------------------- | ---------------------------------------------- | ------------- |
| Voltage/current | `PosSeqMagAnomaly` | Positive sequence magnitude | Adds/subtracts positive sequence magnitude     | Volts or Amps |
| Voltage/current | `PosSeqAngAnomaly` | Positive sequence angle     | Adds/subtracts positive sequence angle         | Degrees       |
| Voltage/current | `PhaseAMagAnomaly` | Phase A magnitude           | Adds/subtracts phase A magnitude               | Volts or Amps |
| Voltage/current | `FreqAnomaly`      | Frequency                   | Adds/subtracts signal frequency                | Hz            |
| Voltage/current | `HarmonicsAnomaly` | All harmonics magnitudes    | Adds/subtracts all harmonic magnitudes         | per unit      |
| Temperature     | `Anomaly`          | Temperature value           | Adds/subtracts instantaneous temperature value | Degrees C     |
