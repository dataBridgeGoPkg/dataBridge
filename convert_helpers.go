package databridge

import "fmt"

// YAML conversion utilities

func convertYAMLToMap(in interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	switch v := in.(type) {
	case map[string]interface{}:
		for k, vv := range v {
			out[k] = convertYAMLValue(vv)
		}
	case map[interface{}]interface{}:
		for k, vv := range v {
			out[fmt.Sprintf("%v", k)] = convertYAMLValue(vv)
		}
	}
	return out
}

func convertYAMLValue(v interface{}) interface{} {
	switch vv := v.(type) {
	case map[string]interface{}:
		return convertYAMLToMap(vv)
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, val := range vv {
			m[fmt.Sprintf("%v", k)] = convertYAMLValue(val)
		}
		return m
	case []interface{}:
		for i, e := range vv {
			vv[i] = convertYAMLValue(e)
		}
		return vv
	default:
		return vv
	}
}
