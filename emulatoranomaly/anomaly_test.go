package emulatoranomaly

import (
	"testing"

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
	container := make(Container)
	err := yaml.Unmarshal([]byte(yamlStr), &container)
	if err != nil {
		// handle error
	}

	t.Logf("container: %v", container)

	for _, hello := range container {
		t.Logf("anomalyparams: %v", hello)
	}
}
