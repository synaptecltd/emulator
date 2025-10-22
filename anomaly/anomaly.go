package anomaly

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/goccy/go-yaml"
	"github.com/synaptecltd/emulator/mathfuncs"
)

// Container is a collection of anomalies.
type Container []AnomalyInterface

// AnomalyInterface is the interface for all anomaly Types (trends, instantaneous, etc).
type AnomalyInterface interface {
	UnmarshalYAML(unmarshal func(any) error) error // Unmarshals an anomaly entry into the correct type based on the type field

	// Inherited from AnomalyBase
	GetName() string                  // Returns the name of the anomaly, used for identification
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
		*c = make(Container, 0)
	}

	// Reading in generically first
	var raw []map[string]any
	err := unmarshal(&raw)
	if err != nil {
		return err
	}

	// iterate through each, action depending on the field of "Type"
	for _, item := range raw {
		typeVal, ok := item["Type"]
		if !ok {
			return fmt.Errorf("missing Type field in anomaly entry")
		}

		typeAsStr, ok := typeVal.(string)
		if !ok {
			return fmt.Errorf("field Type must be a string")
		}

		// Remove Type from map to prevent duplication
		delete(item, "Type")

		// Marshal map back to YAML bytes
		anomalyParams, err := yaml.Marshal(item)
		if err != nil {
			return err
		}
		// Creates correctly typed anomaly and calls its method for parsing via the decodeStrict.
		// This uses its defined UnmarshalYAML method, which populates its fields, and then adds it to the container.
		switch typeAsStr {
		case "spike":
			anomaly := &spikeAnomaly{}
			err := decodeStrict(anomalyParams, anomaly)
			if err != nil {
				return err
			}
			err = c.AddAnomaly(anomaly)
			if err != nil {
				return err
			}
		case "trend":
			anomaly := &trendAnomaly{}
			err := decodeStrict(anomalyParams, anomaly)
			if err != nil {
				return err
			}
			err = c.AddAnomaly(anomaly)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown anomaly type: %s", typeAsStr)
		}
	}
	return nil
}

// decodeStrict decodes data into out using goccy/go-yaml in strict mode.
func decodeStrict(data []byte, out any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.Strict())
	return decoder.Decode(out)
}

// Steps all anomalies within a container and returns the sum of their effects.
func (c Container) StepAll(r *rand.Rand, Ts float64) float64 {
	value := 0.0
	for i := range c {
		// Do by index to not work on copy
		value += c[i].stepAnomaly(r, Ts)
	}
	return value
}

// Add anomaly to container.
func (c *Container) AddAnomaly(anomaly AnomalyInterface) error {
	// Check that the name hasn't already been used
	if c.GetAnomalyByName(anomaly.GetName()) != nil {
		return errors.New("anomaly with name " + anomaly.GetName() + " already exists")
	}
	*c = append(*c, anomaly)
	return nil
}

// GetAnomalyByName returns the first anomaly in the container with the specified name, or nil if not found.
func (c Container) GetAnomalyByName(name string) *AnomalyInterface {
	for _, anomaly := range c {
		if anomaly.GetName() == name {
			return &anomaly
		}
	}
	return nil
}

func (c Container) UpdateAnomalyByName(name string, newAnomaly AnomalyInterface) error {
	for i, anomaly := range c {
		if anomaly.GetName() == name && anomaly.GetTypeAsString() == newAnomaly.GetTypeAsString() {
			c[i] = newAnomaly
			return nil
		}
	}
	return fmt.Errorf("anomaly with name %s and type %s not found", name, newAnomaly.GetTypeAsString())
}
