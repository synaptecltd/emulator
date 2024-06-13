package emulatoranomaly

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestUnmarshalYAML(t *testing.T) {
	// Define a YAML string that represents a trend anomaly.
	yamlStr := `
- name: trend1
  type: trend
  start_delay: 0.1
  duration: 0.2
- name: trend2
  type: trend
  start_delay: 0.5
  duration: 0.1
- name: inst1
  type: instantaneous
  InstantaneousAnomalyProbability: 0.01
`
	var container Container
	err := yaml.Unmarshal([]byte(yamlStr), &container)
	if err != nil {
		// handle error
	}

	for _, hello := range container {
		t.Logf("name: %s, anomalyparams: %v", hello.Name, hello.Anomaly)
	}
}
