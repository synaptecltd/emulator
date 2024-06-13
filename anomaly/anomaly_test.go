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
  type: trend
  start_delay: %f
  duration: %f
inst1:
  type: instantaneous
  probability: %f
`,
		startDelay, duration, probability)

	container := make(anomaly.Container)
	err := yaml.Unmarshal([]byte(yamlStr), &container)
	assert.NoError(t, err)

	trendAnomaly, _ := anomaly.NewTrendAnomaly(anomaly.TrendParams{
		StartDelay: startDelay,
		Duration:   duration,
	})

	instAnomaly := &anomaly.InstantaneousAnomaly{Probability: probability}

	// Todo fix this
	for key, hello := range container {
		if key == "trend1" {
			assert.Equal(t, trendAnomaly, hello)
		} else {
			assert.Equal(t, instAnomaly, hello)
		}
	}
}

func TestGetTypeAsString(t *testing.T) {
	instAnomaly := anomaly.InstantaneousAnomaly{}
	expected := "instantaneous"
	assert.Equal(t, expected, instAnomaly.TypeAsString())

	trendAnomaly, _ := anomaly.NewTrendAnomaly(anomaly.TrendParams{})
	expected = "trend"
	assert.Equal(t, expected, trendAnomaly.TypeAsString())
}
