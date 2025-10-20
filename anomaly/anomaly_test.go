package anomaly_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/synaptecltd/emulator/anomaly"
)

// Test anomalies can be unmarshalled from yaml
func TestUnmarshalYAML(t *testing.T) {
	startDelay := rand.Float64()
	duration := rand.Float64()
	probability := rand.Float64()

	yamlStr := fmt.Sprintf(`
- Type: trend
  StartDelay: %f
  Duration: %f
  Name: Trend Anomaly
- Type: spike
  Probability: %f
  Name: Spike Anomaly
`,
		startDelay, duration, probability)

	var container anomaly.Container
	err := yaml.Unmarshal([]byte(yamlStr), &container)
	assert.NoError(t, err)

	trendAnomaly, _ := anomaly.NewTrendAnomaly(
		anomaly.TrendParams{
			StartDelay: startDelay,
			Duration:   duration,
		})

	instAnomaly, _ := anomaly.NewSpikeAnomaly(
		anomaly.SpikeParams{
			Probability: probability,
		})

	for _, anom := range container {
		var expected anomaly.AnomalyInterface
		switch anom.GetTypeAsString() {
		case "trend":
			expected = trendAnomaly
		case "spike":
			expected = instAnomaly
		}
		assert.Equal(t, expected.GetTypeAsString(), anom.GetTypeAsString())
		assert.InDelta(t, expected.GetDuration(), anom.GetDuration(), 1e-6) // floating point precision
		assert.InDelta(t, expected.GetStartDelay(), anom.GetStartDelay(), 1e-6)

	}
}

func TestUnmarshalYAMLUnknownType(t *testing.T) {
	yamlStr := `
- Type: NotARealType
  SomeField: 1.0
`
	var container anomaly.Container
	err := yaml.Unmarshal([]byte(yamlStr), &container)
	assert.Error(t, err)
}

func TestUnmarshalYAMLDuplicateName(t *testing.T) {
	yamlStr := `
- Type: trend
  StartDelay: 1.0
  Duration: 2.0
  Name: Trend Anomaly
- Type: spike
  Probability: 0.5
  Name: Trend Anomaly
`
	var container anomaly.Container
	err := yaml.Unmarshal([]byte(yamlStr), &container)
	assert.ErrorContains(t, err, "already exists")
}

// Get type of anomaly as string
func TestGetTypeAsString(t *testing.T) {
	instAnomaly, _ := anomaly.NewSpikeAnomaly(anomaly.SpikeParams{})
	expected := "spike"
	assert.Equal(t, expected, instAnomaly.GetTypeAsString())

	trendAnomaly, _ := anomaly.NewTrendAnomaly(anomaly.TrendParams{})
	expected = "trend"
	assert.Equal(t, expected, trendAnomaly.GetTypeAsString())
}

// Test converting AnomalyInterface to trendAnomaly
func TestAsTrendAnomaly(t *testing.T) {
	trendAnomaly, _ := anomaly.NewTrendAnomaly(anomaly.TrendParams{})
	result, ok := anomaly.AsTrendAnomaly(trendAnomaly)
	assert.True(t, ok)
	assert.NotNil(t, result)

	spikeAnomaly, _ := anomaly.NewSpikeAnomaly(anomaly.SpikeParams{})
	result, ok = anomaly.AsTrendAnomaly(spikeAnomaly)
	assert.False(t, ok)
	assert.Nil(t, result)
}

// Test converting AnomalyInterface to spikeAnomaly
func TestAsSpikeAnomaly(t *testing.T) {
	trendAnomaly, _ := anomaly.NewTrendAnomaly(anomaly.TrendParams{})
	result, ok := anomaly.AsSpikeAnomaly(trendAnomaly)
	assert.False(t, ok)
	assert.Nil(t, result)

	spikeAnomaly, _ := anomaly.NewSpikeAnomaly(anomaly.SpikeParams{})
	result, ok = anomaly.AsSpikeAnomaly(spikeAnomaly)
	assert.True(t, ok)
	assert.NotNil(t, result)
}

// Test updating anomaly by name
func TestUpdateAnomalyByName(t *testing.T) {
	container := anomaly.Container{}

	// Add initial anomalies
	trendAnomaly, _ := anomaly.NewTrendAnomaly(anomaly.TrendParams{
		StartDelay: 1.0,
		Duration:   2.0,
		Name:       "TestTrend",
	})

	spikeAnomaly, _ := anomaly.NewSpikeAnomaly(anomaly.SpikeParams{
		Probability: 0.5,
		Name:        "TestSpike",
	})

	container.AddAnomaly(trendAnomaly)
	container.AddAnomaly(spikeAnomaly)

	// Update trend anomaly with new parameters
	newTrendAnomaly, _ := anomaly.NewTrendAnomaly(anomaly.TrendParams{
		StartDelay: 5.0,
		Duration:   10.0,
		Name:       "TestTrend",
	})

	err := container.UpdateAnomalyByName("TestTrend", newTrendAnomaly)
	assert.NoError(t, err)

	// Verify the anomaly was updated
	updatedAnomaly := container.GetAnomalyByName("TestTrend")
	assert.NotNil(t, updatedAnomaly)
	assert.InDelta(t, 5.0, (*updatedAnomaly).GetStartDelay(), 1e-6)
	assert.InDelta(t, 10.0, (*updatedAnomaly).GetDuration(), 1e-6)
}

func TestUpdateAnomalyByNameNotFound(t *testing.T) {
	container := anomaly.Container{}

	// Add an anomaly
	trendAnomaly, _ := anomaly.NewTrendAnomaly(anomaly.TrendParams{Name: "ExistingAnomaly"})
	container.AddAnomaly(trendAnomaly)

	// Try to update non-existent anomaly
	newAnomaly, _ := anomaly.NewSpikeAnomaly(anomaly.SpikeParams{})
	err := container.UpdateAnomalyByName("NonExistentAnomaly", newAnomaly)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "not found")
}

func TestUpdateAnomalyByNameChangeType(t *testing.T) {
	container := anomaly.Container{}

	// Add a trend anomaly
	trendAnomaly, _ := anomaly.NewTrendAnomaly(anomaly.TrendParams{
		StartDelay: 1.0,
		Duration:   2.0,
		Name:       "TestAnomaly",
	})
	container.AddAnomaly(trendAnomaly)

	// Update with a spike anomaly (different type)
	spikeAnomaly, _ := anomaly.NewSpikeAnomaly(anomaly.SpikeParams{
		Probability: 0.8,
		Name:        "TestAnomaly",
	})

	err := container.UpdateAnomalyByName("TestAnomaly", spikeAnomaly)
	assert.Error(t, err)
}

// Checking reverse trend paramater functions as intended
func TestReverseTrend(t *testing.T) {
	container := anomaly.Container{}
	params := anomaly.TrendParams{
		Name:         "ReverseTrendTest",
		StartDelay:   0.0,
		Duration:     5.0,
		Magnitude:    10.0,
		Repeats:      1,
		InvertTrend:  false,
		ReverseTrend: true,
		Off:          false,
	}

	trendAnomaly, err := anomaly.NewTrendAnomaly(params)
	assert.NoError(t, err)
	container.AddAnomaly(trendAnomaly)

	Ts := 1.0 // time step of 1 second
	rng := rand.New(rand.NewPCG(42, 0))
	expectedValues := []float64{10.0, 8.0, 6.0, 4.0, 2.0}

	for i, expected := range expectedValues {
		value := container.StepAll(rng, Ts)
		assert.InDelta(t, expected, value, 1e-6, "At step %d", i)
	}
}
