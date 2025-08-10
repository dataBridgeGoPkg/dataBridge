package databridge

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

// Simple fuzz to ensure we don't panic on arbitrary JSON objects
func FuzzTransformMap(f *testing.F) {
	seeds := []string{
		`{}`,
		`{"a":1}`,
		`{"a":{"b":[1,2,3]},"x":"y"}`,
		`[{"n":"1"},{"n":2}]`,
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		var m map[string]interface{}
		_ = TransformToStructUniversal(input, &m)
		// also try array target
		var ms []map[string]interface{}
		_ = TransformToStructUniversal(input, &ms)
		// and TransformToJSON
		if b, err := TransformToJSON(input, nil); err == nil {
			var tmp interface{}
			_ = json.Unmarshal(b, &tmp)
		}
	})
}

// Fuzz URL-encoded forms with random key/value pairs and dotted keys
func FuzzTransformForm(f *testing.F) {
	// seeds for (a,b,c)
	f.Add("x", "y", "z")
	f.Add("Alice", "Bob", "Carol")
	f.Fuzz(func(t *testing.T, a, b, c string) {
		vals := url.Values{}
		vals.Set("a", a)
		vals.Add("b", b)
		vals.Add("b", c)
		vals.Set("user.name", a)
		vals.Set("user.age", "30")
		var m map[string]interface{}
		_ = TransformToStructUniversal(vals, &m)
	})
}

// Fuzz CSV: random headers and rows (bounded) with commas and quotes
func FuzzTransformCSV(f *testing.F) {
	// seeds for (h1,h2,v1,v2)
	f.Add("name", "age", "Alice", "30")
	f.Add("a", "b", "1", "2")
	f.Add("user.name", "user.age", "Ada", "30")
	f.Fuzz(func(t *testing.T, h1, h2, v1, v2 string) {
		// sanitize headers minimally
		h1 = strings.ReplaceAll(h1, ",", "_")
		h2 = strings.ReplaceAll(h2, ",", "_")
		csv := fmt.Sprintf("%s,%s\n%q,%q\n", h1, h2, v1, v2)
		var rows []map[string]interface{}
		_ = TransformToStructUniversal(csv, &rows)
	})
}
