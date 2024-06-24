package anomaly_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/synaptecltd/emulator/anomaly"
	"gopkg.in/yaml.v2"
)

// Test anomalies can be unmarshalled from yaml
func TestUnmarshalYAML(t *testing.T) {
	testUuid := uuid.New().String()
	startDelay := rand.Float64()
	duration := rand.Float64()
	probability := rand.Float64()

	// Should be able to cope with different capitalisation on "type" field
	yamlStr := fmt.Sprintf(`
- Type: trend
  Uuid: %s
  StartDelay: %f
  Duration: %f
  MagFunc: sine
- type: spike
  Probability: %f
`,
		testUuid, startDelay, duration, probability)

	container := anomaly.Container{}
	err := yaml.Unmarshal([]byte(yamlStr), &container)
	assert.NoError(t, err)

	trendAnomaly, _ := anomaly.NewTrendAnomaly(
		anomaly.TrendParams{
			Uuid:        testUuid,
			StartDelay:  startDelay,
			Duration:    duration,
			MagFuncName: "sine",
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
		assert.Equal(t, expected.GetUuid(), anom.GetUuid())
		assert.NotEmpty(t, anom.GetUuid())                                  // uuid should be set to 00000000-0000-0000-0000-000000000000 if not provided
		assert.InDelta(t, expected.GetDuration(), anom.GetDuration(), 1e-6) // floating point precision
		assert.InDelta(t, expected.GetStartDelay(), anom.GetStartDelay(), 1e-6)

	}
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
