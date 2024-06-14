package anomaly_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/synaptecltd/emulator/anomaly"
	"gopkg.in/yaml.v2"
)

func TestUnmarshalYAML(t *testing.T) {
	startDelay := rand.Float64()
	duration := rand.Float64()
	probability := rand.Float64()

	// Define a YAML string that represents a trend anomaly.
	yamlStr := fmt.Sprintf(`
trend1:
  Type: trend
  StartDelay: %f
  Duration: %f
inst1:
  Type: spike
  Probability: %f
`,
		startDelay, duration, probability)

	container := make(anomaly.Container)
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

func TestGetTypeAsString(t *testing.T) {
	instAnomaly, _ := anomaly.NewSpikeAnomaly(anomaly.SpikeParams{})
	expected := "spike"
	assert.Equal(t, expected, instAnomaly.GetTypeAsString())

	trendAnomaly, _ := anomaly.NewTrendAnomaly(anomaly.TrendParams{})
	expected = "trend"
	assert.Equal(t, expected, trendAnomaly.GetTypeAsString())
}
