package datasource

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/oliveagle/jsonpath"
	"github.com/tumblr/k8s-config-projector/pkg/types"
)

func validateBeforeStructuredProjection(d *DataSource) error {
	// same here. if we asked for multiple field extractions but we are outputting
	// format that is !(json||yaml), we cant do that.
	if len(d.FieldExtractions) > 0 && !(d.OutputFormat == OutputJSON || d.OutputFormat == OutputYAML) {
		return types.ErrUnsupportedOutputFormat
	}

	// if outputformat is raw but we only are using d.Extract for a single field
	// we cant project a single field into structured output
	if d.Extract != "" {
		if d.OutputFormat != OutputRaw {
			return types.ErrUnsupportedOutputFormat
		}
		if len(d.FieldExtractions) > 0 {
			return types.ErrMultipleExtractorsFound
		}
	}

	if d.Extract == "" && len(d.FieldExtractions) == 0 {
		return types.ErrOutputFormatRequiresExtractors
	}

	return nil
}

// projectJSON will extract fields from the json source and project it into
// the desired output format
// returns a list of data item: result string
func (d *DataSource) projectJSON(basePath string) (map[string][]byte, error) {
	if err := validateBeforeStructuredProjection(d); err != nil {
		return nil, err
	}
	// read the JSON source file
	var jsonData interface{}
	bytes, err := ioutil.ReadFile(filepath.Join(basePath, d.Source))
	if err != nil {
		return nil, err
	}
	// Very large numbers get converted to floating points if you use json.Unmarshall
	// Decoding avoids this issue by converting numbers to json.Number type
	// https://stackoverflow.com/questions/22343083/json-marshaling-with-long-numbers-in-golang-gives-floating-point-number
	decoder := json.NewDecoder(strings.NewReader(string(bytes)))
	decoder.UseNumber()
	if err = decoder.Decode(&jsonData); err != nil {
		return nil, err
	}

	// this is the path for handling the jsonPath entry, it parses the field and returns a raw value
	// NOTE: this bails out before we get to the FieldExtractions projection below
	if d.Extract != "" {
		res, err := jsonpath.JsonPathLookup(jsonData, d.Extract)
		// this will probably explode if the dereferenced value isnt a string
		if err != nil {
			return nil, err
		}
		v, err := convertInterfaceValueToBytes(res)
		return map[string][]byte{d.OutputFile: v}, err
	}

	// this is a map of a subset of labels to json fields (which may or may not be structured)
	resArray := map[string]interface{}{}
	for label, path := range d.FieldExtractions {
		res, err := jsonpath.JsonPathLookup(jsonData, string(path))
		// this will probably explode if the dereferenced value isnt a string
		if err != nil {
			return nil, err
		}
		resArray[label] = res
	}

	// return the map as a re-serialized byte array
	// based on the requested structured output format,
	// and the desired output file name (for projections that are
	// structured, there is only 1 output file)
	switch d.OutputFormat {
	case OutputJSON:
		v, err := json.Marshal(resArray)
		return map[string][]byte{d.OutputFile: v}, err
	case OutputYAML:
		v, err := yaml.Marshal(resArray)
		return map[string][]byte{d.OutputFile: v}, err
	default:
		return nil, types.ErrUnsupportedOutputFormat
	}
}

// projectYAML will extract fields from the yaml source and project it into
// the desired output format
// returns a list of data item: result string
func (d *DataSource) projectYAML(basePath string) (map[string][]byte, error) {
	if err := validateBeforeStructuredProjection(d); err != nil {
		return nil, err
	}
	// read the YAML source file
	var yamlData interface{}
	bytes, err := ioutil.ReadFile(filepath.Join(basePath, d.Source))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &yamlData)
	if err != nil {
		return nil, err
	}

	// this is the path for handling the jsonPath entry, it parses the field and returns a raw value
	// NOTE: this bails out before we get to the FieldExtractions projection below
	if d.Extract != "" {
		res, err := jsonpath.JsonPathLookup(yamlData, d.Extract)
		// this will probably explode if the dereferenced value isnt a string
		if err != nil {
			return nil, err
		}
		v, err := convertInterfaceValueToBytes(res)
		return map[string][]byte{d.OutputFile: v}, err
	}

	// this is a map of a subset of labels to json fields (which may or may not be structured)
	resArray := map[string]interface{}{}
	for label, path := range d.FieldExtractions {
		res, err := jsonpath.JsonPathLookup(yamlData, string(path))
		// this will probably explode if the dereferenced value isnt a string
		if err != nil {
			return nil, err
		}
		resArray[label] = res
	}

	// return the map as a re-serialized byte array
	// based on the requested structured output format,
	// and the desired output file name (for projections that are
	// structured, there is only 1 output file)
	switch d.OutputFormat {
	case OutputJSON:
		v, err := json.Marshal(resArray)
		return map[string][]byte{d.OutputFile: v}, err
	case OutputYAML:
		v, err := yaml.Marshal(resArray)
		return map[string][]byte{d.OutputFile: v}, err
	default:
		return nil, types.ErrUnsupportedOutputFormat
	}
}
