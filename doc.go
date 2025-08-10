// Package databridge provides flexible input-to-struct transformation helpers.
//
// It detects and parses JSON, URL-encoded forms (with dotted keys => nested objects),
// CSV (header row), and optionally YAML. It then maps incoming keys to your target
// struct's JSON tags, with case-insensitive, non-alphanumeric-agnostic matching
// by default, and performs type-aware coercion so values like "30" or "true"
// decode into int/bool fields naturally. Strict mode can be enabled to reject
// unknown fields.
//
// Primary APIs:
//   - TransformToStructUniversal(input, &out, options...)
//   - Transform[T any](input, options...) (T, error)
//   - TransformToJSON(input, &out, options...) ([]byte, error)
//
// Example:
//
//	type Person struct {
//	    FirstName string `json:"First_Name"`
//	    Age       int64  `json:"Age"`
//	    Active    bool   `json:"Active"`
//	    Address   struct { City string `json:"city"` } `json:"address"`
//	}
//
//	p, err := databridge.Transform[Person](`{"first-name":"Ada","Age":"30","active":"true","address":{"City":"Paris"}}`)
//	if err != nil { /* handle */ }
package databridge
