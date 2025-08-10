package databridge

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// numeric coercion helpers

func stringToBestType(s string) interface{} {
	if s == "" {
		return ""
	}
	// int
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	// float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	// bool (after numeric parse to avoid "1" => true)
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return s
}

func coerceNumbersInMap(m map[string]interface{}, cfg *config) map[string]interface{} {
	for k, v := range m {
		switch vv := v.(type) {
		case map[string]interface{}:
			m[k] = coerceNumbersInMap(vv, cfg)
		case []interface{}:
			for i, e := range vv {
				if em, ok := e.(map[string]interface{}); ok {
					vv[i] = coerceNumbersInMap(em, cfg)
				} else {
					vv[i] = coerceNumberValue(e)
				}
			}
			m[k] = vv
		default:
			m[k] = coerceNumberValue(vv)
		}
	}
	return m
}

func coerceNumberValue(v interface{}) interface{} {
	switch n := v.(type) {
	case float64:
		if float64(int64(n)) == n {
			return int64(n)
		}
		return n
	default:
		return v
	}
}

func cloneMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func defaultNormalizer(s string) string {
	s = strings.ToLower(s)
	return defaultNormalizeRe.ReplaceAllString(s, "")
}

func bestEffortConvert(m map[string]interface{}) map[string]interface{} {
	return coerceNumbersInMap(m, &config{AllowNumberConv: true})
}

// normalizeMapKeysDeep applies a key normalizer to all keys in the map recursively.
// It preserves the original value shapes and recurses through maps and slices.
func normalizeMapKeysDeep(m map[string]interface{}, normalizer func(string) string) map[string]interface{} {
	if m == nil || normalizer == nil {
		return m
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		nk := normalizer(k)
		switch vv := v.(type) {
		case map[string]interface{}:
			out[nk] = normalizeMapKeysDeep(vv, normalizer)
		case []interface{}:
			arr := make([]interface{}, len(vv))
			for i, e := range vv {
				if em, ok := e.(map[string]interface{}); ok {
					arr[i] = normalizeMapKeysDeep(em, normalizer)
				} else {
					arr[i] = e
				}
			}
			out[nk] = arr
		default:
			out[nk] = v
		}
	}
	return out
}

// --- Type-aware coercion based on target struct shape ---

// coerceAccordingToType walks the input map and converts primitive values (strings, numbers)
// into the types expected by the provided struct type. It handles nested structs and slices.
func coerceAccordingToType(in map[string]interface{}, typ reflect.Type) map[string]interface{} {
	if in == nil {
		return nil
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return in
	}
	// Build map: json field name -> reflect.Type
	fields := buildFieldLookup(typ, nil) // use raw json tags/names (no normalization here)
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		if fi, ok := fields[k]; ok {
			out[k] = coerceValueForType(v, fi.FieldType)
		} else {
			out[k] = v
		}
	}
	return out
}

func coerceValueForType(v interface{}, t reflect.Type) interface{} {
	// Track pointer and unwrap
	isPtr := false
	if t.Kind() == reflect.Ptr {
		isPtr = true
		t = t.Elem()
	}
	// If original field is pointer and incoming value is an empty string, keep it nil
	if isPtr {
		if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
			return nil
		}
	}
	switch t.Kind() {
	case reflect.Bool:
		switch x := v.(type) {
		case string:
			if b, err := strconv.ParseBool(x); err == nil {
				return b
			}
		}
		return v
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch x := v.(type) {
		case string:
			if i, err := strconv.ParseInt(x, 10, 64); err == nil {
				// return as int64; json will fit into desired int size on unmarshal
				return i
			}
		case float64:
			return int64(x)
		}
		return v
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch x := v.(type) {
		case string:
			if u, err := strconv.ParseUint(x, 10, 64); err == nil {
				return u
			}
		case float64:
			if x < 0 {
				return uint64(0)
			}
			return uint64(x)
		}
		return v
	case reflect.Float32, reflect.Float64:
		switch x := v.(type) {
		case string:
			if f, err := strconv.ParseFloat(x, 64); err == nil {
				return f
			}
		case int64:
			return float64(x)
		}
		return v
	case reflect.Struct:
		// Special-case time.Time
		if t == reflect.TypeOf(time.Time{}) {
			switch x := v.(type) {
			case string:
				if tt, ok := parseTimeFlexible(x); ok {
					return tt
				}
			}
			return v
		}
		if m, ok := v.(map[string]interface{}); ok {
			return coerceAccordingToType(m, t)
		}
		return v
	case reflect.Slice, reflect.Array:
		// Expect []T
		if arr, ok := v.([]interface{}); ok {
			elemT := t.Elem()
			out := make([]interface{}, len(arr))
			for i := range arr {
				out[i] = coerceValueForType(arr[i], elemT)
			}
			return out
		}
		return v
	default:
		return v
	}
}

// parseTimeFlexible tries several common timestamp formats including RFC3339.
func parseTimeFlexible(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
