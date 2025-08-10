package databridge

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// formValuesToMapWithDots converts url.Values to a nested map, handling dotted keys like "address.city".
// For keys without dots it behaves like normal form parsing.
func formValuesToMapWithDots(vals url.Values, cfg *config) map[string]interface{} {
	out := make(map[string]interface{}, len(vals))
	// Pass 1: dotted keys -> nested maps
	for k, arr := range vals {
		if strings.Contains(k, ".") {
			parts := strings.Split(k, ".")
			assignNestedValue(out, parts, arr, cfg)
		}
	}
	// Pass 2: non-dotted keys, but don't overwrite if dotted created a nested map already
	for k, arr := range vals {
		if strings.Contains(k, ".") {
			continue
		}
		if _, exists := out[k]; exists {
			// preserve nested map produced by dotted keys
			continue
		}
		if len(arr) == 1 {
			if cfg.AllowNumberConv {
				out[k] = stringToBestType(arr[0])
			} else {
				out[k] = arr[0]
			}
		} else {
			tmp := make([]interface{}, 0, len(arr))
			for _, s := range arr {
				if cfg.AllowNumberConv {
					tmp = append(tmp, stringToBestType(s))
				} else {
					tmp = append(tmp, s)
				}
			}
			out[k] = tmp
		}
	}
	return out
}

func assignNestedValue(m map[string]interface{}, parts []string, arr []string, cfg *config) {
	if len(parts) == 0 {
		return
	}
	head := parts[0]
	if len(parts) == 1 {
		// assign value
		if len(arr) == 1 {
			if cfg.AllowNumberConv {
				m[head] = stringToBestType(arr[0])
			} else {
				m[head] = arr[0]
			}
		} else {
			tmp := make([]interface{}, 0, len(arr))
			for _, s := range arr {
				if cfg.AllowNumberConv {
					tmp = append(tmp, stringToBestType(s))
				} else {
					tmp = append(tmp, s)
				}
			}
			m[head] = tmp
		}
		return
	}
	// ensure nested map exists
	next, ok := m[head]
	if !ok {
		nm := make(map[string]interface{})
		m[head] = nm
		assignNestedValue(nm, parts[1:], arr, cfg)
		return
	}
	if nm, ok := next.(map[string]interface{}); ok {
		assignNestedValue(nm, parts[1:], arr, cfg)
	} else {
		// conflict: overwrite with nested map
		nm := make(map[string]interface{})
		m[head] = nm
		assignNestedValue(nm, parts[1:], arr, cfg)
	}
}

// normalizeDotsToNested converts keys containing dots into nested maps (used for CSV header normalization)
func normalizeDotsToNested(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range row {
		if strings.Contains(k, ".") {
			parts := strings.Split(k, ".")
			assignNestedValue(out, parts, []string{fmt.Sprintf("%v", v)}, &config{AllowNumberConv: true})
			continue
		}
		out[k] = v
	}
	return out
}

// mapToStructKeysRecursive and related helpers

type fieldInfo struct {
	JSONName  string
	FieldType reflect.Type
}

func mapToStructKeysRecursive(in map[string]interface{}, typ reflect.Type, cfg *config) (map[string]interface{}, []string) {
	if in == nil {
		in = map[string]interface{}{}
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	out := make(map[string]interface{})
	unmatched := []string{}

	// build normalized lookup of struct fields
	fieldLookup := buildFieldLookup(typ, cfg.KeyNormalizer)

	seen := map[string]bool{}
	for normKey, info := range fieldLookup {
		if v, ok := in[normKey]; ok {
			// nested struct handling
			if info.FieldType.Kind() == reflect.Struct || (info.FieldType.Kind() == reflect.Ptr && info.FieldType.Elem().Kind() == reflect.Struct) {
				if subMap, ok := v.(map[string]interface{}); ok {
					mappedSub, subUnmatched := mapToStructKeysRecursive(subMap, info.FieldType, cfg)
					out[info.JSONName] = mappedSub
					for _, um := range subUnmatched {
						unmatched = append(unmatched, info.JSONName+"."+um)
					}
				} else {
					out[info.JSONName] = v
				}
			} else {
				out[info.JSONName] = v
			}
			seen[normKey] = true
		}
	}

	// keep leftover keys (unmatched)
	for k, v := range in {
		if _, s := seen[k]; s {
			continue
		}
		out[k] = v
		unmatched = append(unmatched, k)
	}

	return out, unmatched
}

func buildFieldLookup(typ reflect.Type, normalizer func(string) string) map[string]fieldInfo {
	out := map[string]fieldInfo{}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return out
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.PkgPath != "" { // unexported
			continue
		}
		tag := f.Tag.Get("json")
		jsonName := ""
		if tag == "-" {
			continue
		}
		if tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				jsonName = parts[0]
			}
		}
		if jsonName == "" {
			jsonName = f.Name
		}
		norm := jsonName
		if normalizer != nil {
			norm = normalizer(norm)
		}
		out[norm] = fieldInfo{JSONName: jsonName, FieldType: f.Type}
		// alias by normalized field name
		if normalizer != nil {
			norm2 := normalizer(f.Name)
			out[norm2] = out[norm]
		}
	}
	return out
}
