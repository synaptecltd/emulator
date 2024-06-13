package emulatoranomaly

import (
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

// Container is a collection of anomalies.
type Container map[string]AnomalyInterface

// AnomalyInterface is the interface for all anomaly types (trends, instantaneous, etc).
type AnomalyInterface interface {
	UnmarshalYAML(unmarshal func(interface{}) error) error // Unmarshals an anomaly entry into the correct type based on the type field
	TypeAsString() string                                  // Returns the anomaly type as a string
	GetIsAnomalyActive() bool                              // Returns whether the anomaly is active this timestep
	stepAnomaly(r *rand.Rand, Ts float64) float64          // Steps the internal time state of an anomaly and returns the change in signal caused by the anomaly
}

// UnmarshalYAML unmarshals an anomaly entry into the correct type base on the type field.
func (c *Container) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var obj map[string]map[string]interface{}
	if err := unmarshal(&obj); err != nil {
		return err
	}

	for name, anomalyData := range obj {
		var anomaly AnomalyInterface
		switch anomalyData["type"].(string) {
		case "instantaneous":
			anomaly = &InstantaneousAnomaly{}
		case "trend":
			anomaly = &trendAnomaly{}
		default:
			return fmt.Errorf("unknown anomaly type: %s", anomalyData["type"].(string))
		}

		if err := mapstructure.Decode(anomalyData, anomaly); err != nil {
			return err
		}

		(*c)[name] = anomaly
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
