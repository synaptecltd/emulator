package anomaly

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func TestSpikeAnomalyMarshalYaml(t *testing.T) {
	spike := SpikeParams{
		Type:          "spike",
		Name:          "test_spike",
		Repeats:       3,
		Off:           false,
		StartDelay:    2.0,
		Duration:      5.0,
		Magnitude:     4.0,
		MagFuncName:   "linear",
		VaryMagnitude: true,
		SpikeSign:     +1.0,
		Probability:   0.7,
		ProbFuncName:  "step",
	}
	spikeAnomaly, err := NewSpikeAnomaly(spike)
	require.NoError(t, err)

	yamlData, err := yaml.Marshal(spikeAnomaly)
	require.NoError(t, err)

	expectedYAML := `Type: spike
Name: test_spike
Repeats: 3
"Off": false
StartDelay: 2.0
Duration: 5.0
Magnitude: 4.0
MagFunc: linear
VaryMagnitude: true
Sign: 1.0
Probability: 0.7
ProbFunc: step
`
	require.Equal(t, expectedYAML, string(yamlData))
}
