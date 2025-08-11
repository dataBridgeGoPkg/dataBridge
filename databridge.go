package databridge

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"time"
)

var (
	ErrUnsupportedInput = errors.New("databridge: unsupported input type")
	ErrDecodeFailed     = errors.New("databridge: failed to decode input into target")
)

// Config and options ---------------------------------------------------------

type config struct {
	EnableYAML      bool
	NormalizeKeys   bool
	Strict          bool
	Logger          func(format string, args ...interface{})
	AllowNumberConv bool
	KeyNormalizer   func(string) string
	// internal hint to enable cache for default normalizer
	isDefaultKeyNormalizer bool
}

type Option func(*config)

func WithYAML(enabled bool) Option {
	return func(c *config) { c.EnableYAML = enabled }
}

func WithKeyNormalization(enabled bool) Option {
	return func(c *config) { c.NormalizeKeys = enabled }
}

func WithStrict(enabled bool) Option {
	return func(c *config) { c.Strict = enabled }
}

func WithLogger(logger func(format string, args ...interface{})) Option {
	return func(c *config) { c.Logger = logger }
}

func WithNumberConversion(enabled bool) Option {
	return func(c *config) { c.AllowNumberConv = enabled }
}

func WithKeyNormalizer(fn func(string) string) Option {
	return func(c *config) {
		c.KeyNormalizer = fn
		// mark as non-default when user supplies a custom fn
		c.isDefaultKeyNormalizer = false
	}
}

// Public API ----------------------------------------------------------------

// TransformToStructUniversal attempts to decode `input` into `output`.
// output must be a non-nil pointer to a struct OR pointer to a slice (e.g. *[]T).
//
// Supported input types:
//   - string / []byte / io.Reader / *bytes.Buffer : guessed as JSON, form, YAML, XML, CSV (CSV if looks like comma-separated lines + header)
//   - url.Values (form)
//   - map[string]interface{}
//
// Options:
//
//	WithYAML(true) - enable YAML parsing for strings/bytes
//	WithStrict(true) - use DisallowUnknownFields on JSON decode
//	WithKeyNormalizer(fn) - override default key normalization
//
// Behavior highlights:
//   - Forms with dotted keys (e.g., address.city) produce nested maps.
//   - If output is slice type, CSV or multi-row input will map to slice elements.
//   - If output is a struct and CSV contains multiple rows, the first row is used.
func TransformToStructUniversal(input interface{}, output interface{}, opts ...Option) error {
	// validate output
	if output == nil {
		return fmt.Errorf("output must be non-nil pointer")
	}
	outV := reflect.ValueOf(output)
	if outV.Kind() != reflect.Ptr || outV.IsNil() {
		return fmt.Errorf("output must be a non-nil pointer")
	}

	// default config
	cfg := &config{
		EnableYAML:      false,
		NormalizeKeys:   true,
		Strict:          false,
		Logger:          func(string, ...interface{}) {},
		AllowNumberConv: true,
		KeyNormalizer:   defaultNormalizer,
		// default path uses our built-in normalizer
		isDefaultKeyNormalizer: true,
	}
	for _, o := range opts {
		o(cfg)
	}

	// parse input into an intermediate structure:
	// - if CSV => []map[string]interface{}
	// - else => map[string]interface{}
	var (
		intermediateMap map[string]interface{}
		intermediateArr []map[string]interface{}
		err             error
	)

	switch v := input.(type) {
	case string:
		b := []byte(v)
		if !cfg.NormalizeKeys && isLikelyJSON(b) {
			if ok, ferr := fastJSONIntoOutput(b, outV, cfg); ok {
				return ferr
			}
		}
		intermediateMap, intermediateArr, err = parseBytesDetect(b, cfg)
	case []byte:
		if !cfg.NormalizeKeys && isLikelyJSON(v) {
			if ok, ferr := fastJSONIntoOutput(v, outV, cfg); ok {
				return ferr
			}
		}
		intermediateMap, intermediateArr, err = parseBytesDetect(v, cfg)
	case *bytes.Buffer:
		b := v.Bytes()
		if !cfg.NormalizeKeys && isLikelyJSON(b) {
			if ok, ferr := fastJSONIntoOutput(b, outV, cfg); ok {
				return ferr
			}
		}
		intermediateMap, intermediateArr, err = parseBytesDetect(b, cfg)
	case io.Reader:
		b, rerr := io.ReadAll(v)
		if rerr != nil {
			return fmt.Errorf("databridge: read error: %w", rerr)
		}
		if !cfg.NormalizeKeys && isLikelyJSON(b) {
			if ok, ferr := fastJSONIntoOutput(b, outV, cfg); ok {
				return ferr
			}
		}
		intermediateMap, intermediateArr, err = parseBytesDetect(b, cfg)
	case url.Values:
		intermediateMap = formValuesToMapWithDots(v, cfg)
	case map[string]interface{}:
		intermediateMap = cloneMap(v)
	default:
		// if struct / ptr to struct: marshal to JSON then parse
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Struct || (rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct) {
			j, jerr := json.Marshal(v)
			if jerr != nil {
				return fmt.Errorf("databridge: marshal struct: %w", jerr)
			}
			intermediateMap, intermediateArr, err = parseBytesDetect(j, cfg)
		} else {
			return fmt.Errorf("%w: %T", ErrUnsupportedInput, v)
		}
	}
	if err != nil {
		return err
	}

	// normalize keys if needed
	if cfg.NormalizeKeys && cfg.KeyNormalizer != nil {
		if intermediateMap != nil {
			intermediateMap = normalizeMapKeysDeep(intermediateMap, cfg.KeyNormalizer)
		}
		for i := range intermediateArr {
			intermediateArr[i] = normalizeMapKeysDeep(intermediateArr[i], cfg.KeyNormalizer)
		}
	}

	// determine output kind (struct or slice)
	outElem := outV.Elem()
	outElemType := outElem.Type()
	targetIsSlice := outElem.Kind() == reflect.Slice

	// If we have an array input and target is slice => map/normalize/coerce per element then decode whole array
	if intermediateArr != nil && targetIsSlice {
		elemType := outElemType.Elem()
		prepared := make([]map[string]interface{}, 0, len(intermediateArr))
		for _, m := range intermediateArr {
			mapped, unmatched := mapToStructKeysRecursive(m, elemType, cfg)
			if cfg.Strict && len(unmatched) > 0 {
				return fmt.Errorf("databridge: strict mode - unknown fields present: %v", unmatched)
			}
			mapped = coerceAccordingToType(mapped, elemType)
			prepared = append(prepared, mapped)
		}
		// convert []map -> []byte JSON -> unmarshal into output
		j, merr := json.Marshal(prepared)
		if merr != nil {
			return fmt.Errorf("databridge: marshal intermediate array: %w", merr)
		}
		if cfg.Strict {
			dec := json.NewDecoder(bytes.NewReader(j))
			dec.DisallowUnknownFields()
			if derr := dec.Decode(output); derr != nil {
				return fmt.Errorf("%w: %v", ErrDecodeFailed, derr)
			}
			return nil
		}
		if uerr := json.Unmarshal(j, output); uerr != nil {
			// best effort convert and retry
			cfg.Logger("unmarshal slice failed: %v; attempting best-effort conversion", uerr)
			converted := make([]map[string]interface{}, 0, len(prepared))
			for _, mm := range prepared {
				converted = append(converted, bestEffortConvert(mm))
			}
			j2, _ := json.Marshal(converted)
			if err2 := json.Unmarshal(j2, output); err2 != nil {
				return fmt.Errorf("%w: %v", ErrDecodeFailed, err2)
			}
		}
		return nil
	}

	// If array input but target is single struct, use first row
	if intermediateArr != nil && !targetIsSlice {
		if len(intermediateArr) == 0 {
			intermediateMap = map[string]interface{}{}
		} else {
			intermediateMap = intermediateArr[0]
		}
	}

	// map incoming keys to struct field JSON names (struct-aware)
	mapped, unmatched := mapToStructKeysRecursive(intermediateMap, outElemType, cfg)

	// strict top-level check
	if cfg.Strict && len(unmatched) > 0 {
		return fmt.Errorf("databridge: strict mode - unknown fields present: %v", unmatched)
	}

	// marshal mapped and unmarshal into output
	// Coerce primitive types according to target shape to handle strings like "30" -> int
	mapped = coerceAccordingToType(mapped, outElemType)

	j, merr := json.Marshal(mapped)
	if merr != nil {
		return fmt.Errorf("databridge: marshal mapped: %w", merr)
	}

	if cfg.Strict {
		dec := json.NewDecoder(bytes.NewReader(j))
		dec.DisallowUnknownFields()
		if derr := dec.Decode(output); derr != nil {
			return fmt.Errorf("%w: %v", ErrDecodeFailed, derr)
		}
		return nil
	}

	if uerr := json.Unmarshal(j, output); uerr != nil {
		cfg.Logger("unmarshal to output failed: %v; trying best-effort conversion", uerr)
		relaxed := bestEffortConvert(mapped)
		j2, _ := json.Marshal(relaxed)
		if uerr2 := json.Unmarshal(j2, output); uerr2 != nil {
			return fmt.Errorf("%w: %v", ErrDecodeFailed, uerr2)
		}
	}

	return nil
}

// isLikelyJSON performs a quick check for JSON payloads ('{' or '[' after trimming spaces/BOM).
func isLikelyJSON(b []byte) bool {
	// strip potential UTF-8 BOM
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		b = b[3:]
	}
	// trim leading spaces
	i := 0
	for i < len(b) && (b[i] == ' ' || b[i] == '\n' || b[i] == '\r' || b[i] == '\t') {
		i++
	}
	if i >= len(b) {
		return false
	}
	c := b[i]
	return c == '{' || c == '['
}

// fastJSONIntoOutput tries to decode JSON directly into the output type to bypass mapping when possible.
// Returns (handled, err). When handled is true, caller should return err.
func fastJSONIntoOutput(b []byte, outV reflect.Value, cfg *config) (bool, error) {
	// Only attempt direct decode when key normalization isn't required
	// or when the payload already matches struct tags; unknown fields will fail in Strict mode.
	// We optimistically try; on failure we signal not handled so the slower path can run.
	// Create a fresh value of the output element type to avoid partially mutating caller's value on failure.
	outElem := outV.Elem()
	outElemType := outElem.Type()
	tmpPtr := reflect.New(outElemType)
	dec := json.NewDecoder(bytes.NewReader(b))
	if cfg.Strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(tmpPtr.Interface()); err != nil {
		return false, nil
	}
	// success: set into caller's output
	outElem.Set(tmpPtr.Elem())
	return true, nil
}

// Transform is a generic convenience wrapper that returns a value of type T.
// Example: user := databridge.Transform[User](formOrJSON)
func Transform[T any](input interface{}, opts ...Option) (T, error) {
	var out T
	if err := TransformToStructUniversal(input, &out, opts...); err != nil {
		var zero T
		return zero, err
	}
	return out, nil
}

// TransformToJSON marshals the decoded struct into JSON bytes.
// outputPtr should be a pointer to the desired struct or slice type; if nil, a generic map will be produced.
func TransformToJSON(input interface{}, outputPtr interface{}, opts ...Option) ([]byte, error) {
	if outputPtr == nil {
		// default to map[string]interface{}
		m := map[string]interface{}{}
		if err := TransformToStructUniversal(input, &m, opts...); err != nil {
			return nil, err
		}
		return json.Marshal(m)
	}
	if err := TransformToStructUniversal(input, outputPtr, opts...); err != nil {
		return nil, err
	}
	return json.Marshal(outputPtr)
}

// FromJSON decodes JSON bytes into T using the fastest path (no key normalization),
// honoring Strict mode if provided via options.
// Use when your payload keys already match your struct json tags.
func FromJSON[T any](b []byte, opts ...Option) (T, error) {
	// force normalization off to enable the fast path
	opts = append(opts, WithKeyNormalization(false))
	var out T
	if err := TransformToStructUniversal(b, &out, opts...); err != nil {
		var zero T
		return zero, err
	}
	return out, nil
}

// FromJSONString is a convenience wrapper over FromJSON.
func FromJSONString[T any](s string, opts ...Option) (T, error) {
	return FromJSON[T]([]byte(s), opts...)
}

//go:generate go run ./cmd/databridge-gen -types User,Order -out zz_databridge_gen.go

// Sample domain structs for generator demo.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Active    bool      `json:"active"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	Address   struct {
		Line1 string `json:"line1"`
		City  string `json:"city"`
		Zip   string `json:"zip"`
	} `json:"address"`
}

type Order struct {
	OrderID   string    `json:"order_id"`
	UserID    int64     `json:"user_id"`
	Amount    float64   `json:"amount"`
	Paid      bool      `json:"paid"`
	Items     []string  `json:"items"`
	CreatedAt time.Time `json:"created_at"`
}
