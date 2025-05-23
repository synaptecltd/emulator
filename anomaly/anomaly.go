package anomaly

import (
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/synaptecltd/emulator/mathfuncs"
	"gopkg.in/yaml.v3"
)

// Container is a collection of anomalies.
type Container map[string]AnomalyInterface

// AnomalyInterface is the interface for all anomaly Types (trends, instantaneous, etc).
type AnomalyInterface interface {
	UnmarshalYAML(unmarshal func(any) error) error // Unmarshals an anomaly entry into the correct type based on the type field

	// Inherited from AnomalyBase
	GetTypeAsString() string          // Returns the type of anomaly as a string
	GetStartDelay() float64           // Returns the start time of anomalies in seconds
	GetDuration() float64             // Returns the duration of each anomaly in seconds
	GetIsAnomalyActive() bool         // Returns whether the anomaly is active this timestep
	GetStartDelayIndex() int          // Returns the start delay of the anomaly in time steps
	GetElapsedActivatedIndex() int    // Returns the number of time steps since the start of the active anomaly trend/burst
	GetElapsedActivatedTime() float64 // Returns the time elapsed since the start of the active anomaly trend/burst
	GetCountRepeats() uint64          // Returns the number of times the anomaly trend/burst has repeated so far
	SetStartDelay(float64) error      // Sets the start time of anomalies in seconds if delay >= 0
	SetFunctionByName(
		string, func(string) (mathfuncs.MathsFunction, error), *string, *mathfuncs.MathsFunction) error // Sets the function used to vary the parameters of an anomaly using a name string (see mathfuncs for available functions)

	stepAnomaly(r *rand.Rand, Ts float64) float64 // Steps the internal time state of an anomaly and returns the change in signal caused by the anomaly
}

// Attempts to cast an AnomalyInterface to a trendAnomaly. Returns the anomaly as a trendAnomaly and boolean indicating success.
func AsTrendAnomaly(a AnomalyInterface) (*trendAnomaly, bool) {
	trendAnomaly, ok := a.(*trendAnomaly)
	return trendAnomaly, ok
}

// Attempts to cast an AnomalyInterface to a spikeAnomaly. Returns the anomaly as a spikeAnomaly and boolean indicating success.
func AsSpikeAnomaly(a AnomalyInterface) (*spikeAnomaly, bool) {
	spikeAnomaly, ok := a.(*spikeAnomaly)
	return spikeAnomaly, ok
}

// Unmarshals a generic anomaly entry into the correct type base on the anomaly "Type" field.
func (c *Container) UnmarshalYAML(unmarshal func(any) error) error {
	// Create the container if passed an empty pointer
	if *c == nil {
		*c = make(Container)
	}

	var raw map[string]map[string]any
	if err := unmarshal(&raw); err != nil {
		return err
	}
	// Match on the definition of the anomaly type
	for key, value := range raw {
		var anomaly AnomalyInterface
		switch value["Type"].(string) {
		case "spike":
			anomaly = &spikeAnomaly{}
		case "trend":
			anomaly = &trendAnomaly{}
		default:
			return fmt.Errorf("unknown anomaly type: %s", value["Type"].(string))
		}

		// Convert the value map into YAML for unmarshalling into an anomaly
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
