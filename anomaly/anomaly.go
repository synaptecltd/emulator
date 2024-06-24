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

// Unmarshals a generic anomaly entry into the correct type base on the anomaly "type" field.
func (c *Container) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux []map[string]interface{}
	if err := unmarshal(&aux); err != nil {
		return err
	}

	for _, m := range aux {
		var a AnomalyInterface
		// always check m["type"] with lower case regardless of the case in the yaml file because it is converted to lower case by the yaml parser
		typeStr, ok := m["type"].(string)
		if !ok {
			return errors.New("type field is missing or not a string")
		}
		switch typeStr {
		case "trend":
			a = &trendAnomaly{}
		case "spike":
			a = &spikeAnomaly{}
		default:
			return fmt.Errorf("unknown anomaly type: %s", m["type"].(string))
		}

		bytes, err := yaml.Marshal(m)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(bytes, a); err != nil {
			return err
		}

		*c = append(*c, a)
	}

	return nil
}

func GetDecodeHook() (mapstructure.DecodeHookFunc, error) {
	decodeHook := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t == reflect.TypeOf((*AnomalyInterface)(nil)).Elem() {
			m, ok := data.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("data is not a map[string]interface{}: %v", data)
			}

			var a AnomalyInterface
			// always check m["type"] with lower case regardless of the case in the yaml file because it is converted to lower case by the yaml parser
			typeStr, ok := m["type"].(string)
			if !ok {
				return nil, fmt.Errorf("type field is missing or not a string")
			}
			switch typeStr {
			case "trend":
				a = &trendAnomaly{}
			case "spike":
				a = &spikeAnomaly{}
			default:
				return nil, fmt.Errorf("unknown anomaly type: %s", m["type"].(string))
			}

			bytes, err := yaml.Marshal(m)
			if err != nil {
				return nil, err
			}

			if err := yaml.Unmarshal(bytes, a); err != nil {
				return nil, err
			}

			return a, nil
		}

		// Fallback to the default behavior
		return data, nil
	}
	return decodeHook, nil
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
