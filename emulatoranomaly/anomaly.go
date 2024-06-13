package emulatoranomaly

import (
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

// Container is a collection of anomalies.
type Container []AnomalyEntry

// AnomalyEntry is a named anomaly.
type AnomalyEntry struct {
	Name    string           `yaml:"name"`
	Anomaly AnomalyInterface `yaml:",inline"`
}

// AnomalyInterface is the interface for all anomaly types (trends, instantaneous, etc).
type AnomalyInterface interface {
	UnmarshalYAML(unmarshal func(interface{}) error) error
	stepAnomaly(r *rand.Rand, Ts float64) float64
	typeAsString() string
}

// UnmarshalYAML unmarshals an anomaly entry into the correct type base on the type field.
func (ae *AnomalyEntry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var obj map[string]interface{}
	if err := unmarshal(&obj); err != nil {
		return err
	}

	ae.Name = obj["name"].(string)
	switch obj["type"].(string) {
	case "instantaneous":
		ae.Anomaly = &InstantaneousAnomaly{}
	case "trend":
		ae.Anomaly = &trendAnomaly{}
	default:
		return fmt.Errorf("unknown anomaly type: %s", obj["type"].(string))
	}

	return mapstructure.Decode(obj, ae.Anomaly)
}

// Steps all anomalies within a container and returns the sum of their effects.
func (c Container) StepAll(r *rand.Rand, Ts float64) float64 {
	value := 0.0
	for i := range c {
		// Do by index to not work on copy
		value += c[i].Anomaly.stepAnomaly(r, Ts)
	}
	return value
}

// Add anomaly to container and return the UUID of the anomaly.
func (c *Container) AddAnomaly(anomaly AnomalyInterface) uuid.UUID {
	uuid := uuid.New()
	*c = append(*c, AnomalyEntry{Name: uuid.String(), Anomaly: anomaly})
	return uuid
}
