package datasource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// takes some interface and returns it converted to a byte buffer
// NOTE: returned []byte is little endian encoded
// this does some reflection to ensure we are rendering a value
// correctly.
// NOTE: this is used by Project to properly format typed extractions
// as string for raw output, preserving formatting as expected.
func convertInterfaceValueToBytes(data interface{}) ([]byte, error) {
	// try to parse the value out as somethign scalar we can
	// convert into a []byte. Never return the wire representation
	// of a value; always convert into a string first!

	switch data.(type) {
	case string:
		return []byte(data.(string)), nil
	case int64:
		return []byte(strconv.FormatInt(data.(int64), 10)), nil
	case int:
		return []byte(strconv.Itoa(data.(int))), nil
	case float64:
		return []byte(strconv.FormatFloat(data.(float64), 'f', -1, 64)), nil
	case json.Number:
		// If the json.Number has an 'e' in it, it's scientific notation and we should leave it as such
		if dataString := data.(json.Number).String(); strings.ContainsAny(dataString, "eE") {
			return convertInterfaceValueToBytes(dataString)
		}

		// JSON Number can be converted into int64 or float64.  Try int64 first, then float 64
		if i, err := data.(json.Number).Int64(); err == nil {
			return convertInterfaceValueToBytes(i)
		}

		if f, err := data.(json.Number).Float64(); err == nil {
			return convertInterfaceValueToBytes(f)
		}

		return nil, fmt.Errorf("Unable to convert json.Number %v to any value of Int64 or Float64", data)
	case bool:
		return []byte(strconv.FormatBool(data.(bool))), nil
	default:
		// fall back to using reflection to see if we can support some odd usecases like []string, etc
		if v := reflect.ValueOf(data); v.Kind() == reflect.Slice {
			// make a slice of strings to hold this data
			dataInterfaceSlice := data.([]interface{})
			s := make([]string, len(dataInterfaceSlice))
			for i, x := range dataInterfaceSlice {
				// convert this slice into a []string{...}. for now, we only support scalar extraction of []string, no other types of slice
				if resx, ok := x.(string); ok {
					s[i] = resx
				} else {
					return nil, fmt.Errorf("unable extract scalar value from slice, only []string are supported currently. try extracting a specific element. unsupported datatype %v", x)
				}
			}
			return []byte(strings.Join(s, ",")), nil
		}
		return nil, fmt.Errorf("unable extract scalar value, unsupported datatype %v", data)
	}
}
