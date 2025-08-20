package anomaly

import (
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

	var raw []map[string]any
	err := unmarshal(&raw)
	if err != nil {
		return err
	}

	var anomaly AnomalyInterface

	// Match on the definition of the anomaly type
	for _, anomalyEntry := range raw {
		if anomalyEntry["Type"] != nil {
			switch anomalyEntry["Type"] {
			case "spike":
				anomaly = &spikeAnomaly{}
			case "trend":
				anomaly = &trendAnomaly{}
			default:
				return fmt.Errorf("unknown anomaly type: %s", anomalyEntry["Type"])
			}
		}
		// Convert the value map into YAML for unmarshalling into an anomaly
		valueYAML, err := yaml.Marshal(anomalyEntry)
		if err != nil {
			return err
		}

		// Unmarshal the YAML into the anomaly
		err = yaml.Unmarshal(valueYAML, anomaly)
		if err != nil {
			return err
		}

		err = c.AddAnomaly(anomaly)
		if err != nil {
			return err
		}
	}
	return nil
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
