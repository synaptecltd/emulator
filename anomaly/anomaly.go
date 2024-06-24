package anomaly

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/synaptecltd/emulator/mathfuncs"
	"gopkg.in/yaml.v2"
)

// Container is a collection of anomalies.
type Container []AnomalyInterface

// AnomalyInterface is the interface for all anomaly Types (trends, instantaneous, etc).
type AnomalyInterface interface {
	UnmarshalYAML(unmarshal func(interface{}) error) error // Unmarshals an anomaly entry into the correct type based on the type field

	// Inherited from AnomalyBase
	GetTypeAsString() string          // Returns the type of anomaly as a string
	GetUuid() uuid.UUID               // Returns the unique identifier for the anomaly
	SetUuid(uuid.UUID)                // Sets the unique identifier for the anomaly
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

// Unmarshals a yaml file into the container.
func (c *Container) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Temporary structure to unmarshal the yaml file
	var unmarshaledYaml []map[string]interface{}
	if err := unmarshal(&unmarshaledYaml); err != nil {
		return err
	}

	for _, yamlEntry := range unmarshaledYaml {
		ai, err := createAnomalyFromYamlEntry(yamlEntry)
		if err != nil {
			return err
		}
		*c = append(*c, ai)
	}

	return nil
}

// Returns a decodeHook function that can be used to unmarshal anomalies from a yaml file using mapstructure.
// This supports configuration solutions like spf13/viper that use mapstructure to unmarshal yaml files.
func GetDecodeHook() (mapstructure.DecodeHookFunc, error) {
	decodeHook := func(f reflect.Type, t reflect.Type, yamlEntry interface{}) (interface{}, error) {
		if t == reflect.TypeOf((*AnomalyInterface)(nil)).Elem() {
			// If the target type is AnomalyInterface, create the correct anomaly type from the yaml entry
			return createAnomalyFromYamlEntry(yamlEntry)
		}
		// Otherwise, return the yaml entry as is (default behaviour)
		return yamlEntry, nil
	}

	return decodeHook, nil
}

// Creates a generic anomaly from a yaml entry based on the anomaly "type" (or "Type") field.
func createAnomalyFromYamlEntry(yamlEntry interface{}) (AnomalyInterface, error) {
	// yaml entries should always be a string key with some sort of value
	m, ok := yamlEntry.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("yaml entry cannot be parsed to map[string]interface{}: %v", yamlEntry)
	}

	// must check both m["type"] and m["Type"] because some yaml parsers convert to lower case and some don't
	typeStr, ok := m["type"].(string)
	if !ok {
		typeStr, ok = m["Type"].(string)
		if !ok {
			return nil, errors.New("anomaly type field is missing or not a string")
		}
	}

	var ai AnomalyInterface
	switch typeStr {
	case "trend":
		ai = &trendAnomaly{}
	case "spike":
		ai = &spikeAnomaly{}
	default:
		return nil, fmt.Errorf("unknown anomaly type: %s", typeStr)
	}

	// Unmarshal the yaml entry into the correct anomaly type
	bytes, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(bytes, ai); err != nil {
		return nil, err
	}

	return ai, nil
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

// Add anomaly to container returning the uuid of the added anomaly.
func (c *Container) AddAnomaly(anomaly AnomalyInterface) uuid.UUID {
	uuid := uuid.New()
	anomaly.SetUuid(uuid)
	*c = append(*c, anomaly)
	return uuid
}
