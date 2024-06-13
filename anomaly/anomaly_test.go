package anomaly_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/synaptecltd/emulator/anomaly"
	"gopkg.in/yaml.v2"
)

func TestUnmarshalYAML(t *testing.T) {
	// Define a YAML string that represents a trend anomaly.
	yamlStr := `
trend1:
  type: trend
  start_delay: 0.1
  duration: 0.2
trend2:
  type: trend
  start_delay: 0.5
  duration: 0.1
inst1:
  type: instantaneous
  InstantaneousAnomalyProbability: 0.01
`
	container := make(anomaly.Container)
	err := yaml.Unmarshal([]byte(yamlStr), &container)
	assert.NoError(t, err)

	t.Logf("container: %v", container)

	for _, hello := range container {
		t.Logf("anomalyparams: %v", hello)
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
