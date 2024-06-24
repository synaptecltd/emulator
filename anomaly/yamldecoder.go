package anomaly

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

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

	// Use mapstructure to decode the map into AnomalyInterface
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			trendAnomalyDecodeHookFunc(), // decodeHook for trendAnomaly
			spikeAnomalyDecodeHookFunc(), // decodeHook for spikeAnomaly
			// add more decoders here as required
		),
		Result: &ai,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(m); err != nil {
		return nil, err
	}

	return ai, nil
}

// Returns a DecodeHookFunc that can be used to unmarshal a trendAnomaly from a yaml file.
func trendAnomalyDecodeHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t == reflect.TypeOf(trendAnomaly{}) {
			// unmarshal into TrendParams and use constructor function to create trendAnomaly
			var params TrendParams
			anomalyParamsDecodeHookFunc(&params, data)
			return NewTrendAnomaly(params)
		}
		// If the type is not trendAnomaly, return data unchanged
		return data, nil
	}
}

// Returns a DecodeHookFunc that can be used to unmarshal a spikeAnomaly from a yaml file.
func spikeAnomalyDecodeHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t == reflect.TypeOf(spikeAnomaly{}) {
			// unmarshal into SpikeParams and use constructor function to create spikeAnomaly
			var params SpikeParams
			anomalyParamsDecodeHookFunc(&params, data)
			return NewSpikeAnomaly(params)
		}
		// If the type is not spikeAnomaly, return data unchanged
		return data, nil
	}
}

// Use mapstructure to unmarshal data into anomalyParams.
func anomalyParamsDecodeHookFunc[T any](anomalyParams *T, data interface{}) error {
	m, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, got %T", data)
	}

	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(), // parses Uuids
		),
		Result: &anomalyParams,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}
	if err := decoder.Decode(m); err != nil {
		return err
	}
	return nil
}
