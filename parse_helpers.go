package databridge

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// parseBytesDetect tries formats in order: JSON -> form -> YAML -> XML -> CSV -> fallback string
// Returns either a single map (map[string]interface{}) or an array ([]map[string]interface{}) for multi-row formats (CSV)
func parseBytesDetect(b []byte, cfg *config) (map[string]interface{}, []map[string]interface{}, error) {
	trim := bytes.TrimSpace(b)
	if len(trim) == 0 {
		return map[string]interface{}{}, nil, nil
	}

	// JSON (object)
	var jm map[string]interface{}
	if json.Unmarshal(trim, &jm) == nil {
		return coerceNumbersInMap(jm, cfg), nil, nil
	}
	// JSON (array of objects)
	var jarr []map[string]interface{}
	if json.Unmarshal(trim, &jarr) == nil {
		// Coerce numbers within each object
		if cfg != nil {
			for i := range jarr {
				jarr[i] = coerceNumbersInMap(jarr[i], cfg)
			}
		}
		return nil, jarr, nil
	}

	// form (heuristic)
	str := string(trim)
	if looksLikeForm(str) {
		if vals, err := url.ParseQuery(str); err == nil {
			return formValuesToMapWithDots(vals, cfg), nil, nil
		}
	}

	// YAML
	if cfg.EnableYAML {
		var yv interface{}
		if err := yaml.Unmarshal(trim, &yv); err == nil {
			converted := convertYAMLToMap(yv)
			return coerceNumbersInMap(converted, cfg), nil, nil
		}
	}

	// XML (best-effort)
	var xi interface{}
	if xml.Unmarshal(trim, &xi) == nil {
		var any interface{}
		if err := xml.Unmarshal(trim, &any); err == nil {
			if j, merr := json.Marshal(any); merr == nil {
				var mm map[string]interface{}
				if json.Unmarshal(j, &mm) == nil {
					return coerceNumbersInMap(mm, cfg), nil, nil
				}
			}
		}
	}

	// CSV
	if looksLikeCSV(str) {
		rows, cerr := parseCSVToMaps(str)
		if cerr == nil && len(rows) > 0 {
			return nil, rows, nil
		}
	}

	return map[string]interface{}{"value": str}, nil, nil
}

func looksLikeCSV(s string) bool {
	if !strings.Contains(s, "\n") {
		return false
	}
	firstLine := strings.SplitN(s, "\n", 2)[0]
	return strings.Contains(firstLine, ",")
}

func looksLikeForm(s string) bool {
	if strings.Contains(s, "<") || strings.Contains(s, ">") || strings.Contains(s, "{") || strings.Contains(s, "}") {
		return false
	}
	return strings.Contains(s, "=")
}

// parseCSVToMaps parses CSV assuming first row header and returns slice of row maps.
func parseCSVToMaps(s string) ([]map[string]interface{}, error) {
	r := csv.NewReader(strings.NewReader(s))
	// Allow variable number of fields per record; we'll align using the header
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("databridge: csv read error: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	header := records[0]
	if len(header) > 0 && len(header[0]) > 0 {
		// Strip UTF-8 BOM if present in the first header cell
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}
	out := make([]map[string]interface{}, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		row := records[i]
		m := make(map[string]interface{}, len(header))
		for j, h := range header {
			var val string
			if j < len(row) {
				val = row[j]
			} else {
				val = ""
			}
			m[h] = stringToBestType(val)
		}
		out = append(out, normalizeDotsToNested(m))
	}
	return out, nil
}
