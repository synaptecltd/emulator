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
	type testcase struct {
		yamlStr     string
		ErrorExp    bool
		Description string
	}

	startDelay := rand.Float64()
	duration := rand.Float64()
	probability := rand.Float64()

	testcases := []testcase{
		{
			fmt.Sprintf(`
- Type: trend
  StartDelay: %f
  Repeats: 3
  Duration: %f
  Name: Trend Anomaly
- Type: spike
  Probability: %f
  Repeats: 3
  Name: Spike Anomaly
`, startDelay, duration, probability),
			false,
			"Correctly Specified, No Error Expected",
		},
		{
			fmt.Sprintf(`
- Type: trend
  StartDelay: %f
  Repeats: 3
  Duration: %f
  Name: Trend Anomaly
- Type: spike
  Probability: %f
  Repeat: 3
  Name: Spike Anomaly
`, startDelay, duration, probability),
			true,
			"Incorrectly Specified, Use of Repeat vs Repeats, Error Expected",
		},
		{
			fmt.Sprintf(`
- StartDelay: %f
  Repeats: 3
  Duration: %f
  Name: Trend Anomaly
`, startDelay, duration),
			true,
			"Incorrectly Specified, No Type, Error Expected",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Description, func(t *testing.T) {
			t.Log(tc.yamlStr)
			// Contains yaml override
			var container anomaly.Container
			// Attempt parsing
			err := yaml.Unmarshal([]byte(tc.yamlStr), &container)
			// Check bad params are rejected
			if tc.ErrorExp {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			// Validate contents

			// These are the expected anomalies
			trendAnomaly, _ := anomaly.NewTrendAnomaly(
				anomaly.TrendParams{
					StartDelay: startDelay,
					Duration:   duration,
				})
			spikeAnomaly, _ := anomaly.NewSpikeAnomaly(
				anomaly.SpikeParams{
					Probability: probability,
				})

			for _, anom := range container {
				var expected anomaly.AnomalyInterface
				switch anom.GetTypeAsString() {
				case "trend":
					expected = trendAnomaly
				case "spike":
					expected = spikeAnomaly
				}
				assert.Equal(t, expected.GetTypeAsString(), anom.GetTypeAsString())
				assert.InDelta(t, expected.GetDuration(), anom.GetDuration(), 1e-6) // floating point precision
				assert.InDelta(t, expected.GetStartDelay(), anom.GetStartDelay(), 1e-6)
			}
		})
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
