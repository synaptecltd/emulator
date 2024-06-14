package anomaly

import (
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

// Container is a collection of anomalies.
type Container map[string]AnomalyInterface

// AnomalyInterface is the interface for all anomaly types (trends, instantaneous, etc).
type AnomalyInterface interface {
	UnmarshalYAML(unmarshal func(interface{}) error) error // Unmarshals an anomaly entry into the correct type based on the type field
	TypeAsString() string                                  // Returns the anomaly type as a string
	GetIsAnomalyActive() bool                              // Returns whether the anomaly is active this timestep
	GetDuration() float64                                  // Returns the duration of each anomaly in seconds
	GetStartDelay() float64                                // Returns the start time of anomalies in seconds
	stepAnomaly(r *rand.Rand, Ts float64) float64          // Steps the internal time state of an anomaly and returns the change in signal caused by the anomaly
}

// UnmarshalYAML unmarshals an anomaly entry into the correct type base on the type field.
func (c *Container) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	for key, value := range raw {
		var anomaly AnomalyInterface
		switch value["type"].(string) {
		case "spike":
			anomaly = &SpikeAnomaly{}
		case "trend":
			anomaly = &trendAnomaly{}
		default:
			return fmt.Errorf("unknown anomaly type: %s", value["type"].(string))
		}

		// Convert the value map back into YAML
		valueYAML, err := yaml.Marshal(value)
		if err != nil {
			return err
		}

		// Unmarshal the YAML into the anomaly
		if err := yaml.Unmarshal(valueYAML, anomaly); err != nil {
			return err
		}

		(*c)[key] = anomaly
	}

	return nil
}

// Steps all anomalies within a container and returns the sum of their effects.
func (c Container) StepAll(r *rand.Rand, Ts float64) float64 {
	value := 0.0
	for key := range c {
		// Do by index to not work on copy
		value += c[key].stepAnomaly(r, Ts)
	}
	return value
}

// Add anomaly to container with a UUID and returns the UUID.
func (c *Container) AddAnomaly(anomaly AnomalyInterface) uuid.UUID {
	uuid := uuid.New()
	(*c)[uuid.String()] = anomaly
	return uuid
}
