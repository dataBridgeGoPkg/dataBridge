# DataBridge: flexible input-to-struct transformer for Go

[![CI](https://github.com/dataBridgeGoPkg/dataBridge/actions/workflows/ci.yml/badge.svg)](https://github.com/dataBridgeGoPkg/dataBridge/actions/workflows/ci.yml)

DataBridge helps you accept many input shapes (JSON, strings, URL-encoded forms, CSV, best-effort YAML) and map them into your own structs or slices with minimal fuss. It also offers JSON output helpers.
Status: library-only package.

## Why
- Accepts string, []byte, io.Reader, url.Values, map[string]interface{}.
- Detects JSON (objects and arrays of objects), URL-encoded form (with dotted keys => nested objects), CSV (header row), and optionally YAML.
- Maps incoming keys to your struct JSON tags, with normalization (case-insensitive, ignores non-alphanumerics) by default.
- Converts string numbers/bools into the right target types automatically.
- Strict mode rejects unknown fields.

## Install
Replace the module path with your published path when you push this repo.

```bash
# in your project
go get github.com/dataBridgeGoPkg/dataBridge
```

Go 1.21+ (CI tests 1.21/1.22/1.23).

## Quick start

```go
package main

import (
    "fmt"
    "net/url"

    "github.com/dataBridgeGoPkg/dataBridge"
)

type Person struct {
    FirstName string `json:"First_Name"`
    Age       int64  `json:"Age"`
    Active    bool   `json:"Active"`
    Address   struct {
        City string `json:"city"`
    } `json:"address"`
}

func main() {
    // JSON string
    in := `{"first-name":"John","Age":"30","active":"true","address":{"City":"Paris"}}`
    p, err := databridge.Transform[Person](in)
    if err != nil { panic(err) }
    fmt.Printf("%+v\n", p)

    // URL form with dotted keys
    f := url.Values{"First_Name": {"John"}, "Age": {"30"}, "Active": {"true"}, "address.city": {"Lyon"}}
    var p2 Person
    if err := databridge.TransformToStructUniversal(f, &p2); err != nil { panic(err) }

    // CSV -> slice
    csv := "name,age\nAlice,30\nBob,25\n"
    type Small struct { Name string `json:"name"`; Age int64 `json:"age"` }
    s, err := databridge.Transform[[]Small](csv)
    if err != nil { panic(err) }
    fmt.Println(s)
}
```

## API

- TransformToStructUniversal(input, outputPtr, options...)
  - Accepts: string, []byte, io.Reader, *bytes.Buffer, url.Values, map[string]interface{}, and structs (marshaled then parsed).
    - JSON arrays of objects are supported: decode directly into []T when output is a slice.
  - Options:
    - WithYAML(true)
    - WithKeyNormalization(true|false)
    - WithStrict(true)
    - WithLogger(fn)
    - WithNumberConversion(true|false)
    - WithKeyNormalizer(fn)
- Transform[T any](input, options...) (T, error): generic convenience wrapper.
- TransformToJSON(input, outputPtr, options...) ([]byte, error): decode and marshal in one step. If outputPtr is nil, returns a generic map as JSON.

## Notes
- YAML support is optional and off by default; enable with WithYAML(true). Uses gopkg.in/yaml.v3.
- CSV expects a header row; returns a slice when your target is []T. If target is a struct, the first row is used.
- Strict mode applies to arrays too: each element is validated for unknown fields.
- XML support is best-effort only; if you need robust XML mapping, we can wire a proper decoder.

### Key conflicts and normalization
- Dotted keys (e.g., `user.name`) nest under `user`. If a flat key (`user`) also exists, the nested map takes precedence to avoid type conflicts.
- Normalization lowers case and strips non-alphanumerics by default. You can override via `WithKeyNormalizer(fn)`.
- Colliding keys after normalization map deterministically; prefer the struct tag matches. Unknown leftovers are preserved unless `WithStrict(true)` is used.

### CSV behavior and quirks
- Header row determines field names; dotted headers create nested objects.
- Rows with fewer columns than headers fill missing values with empty strings; extra columns are ignored.
- Duplicate header names keep the last occurrence for that column position.
- UTF‑8 BOM at the start of the header is stripped.

## Development

Run tests:

```bash
go test ./... -v
```

CI
- GitHub Actions runs vet, tests, race, and a short fuzz pass across Go 1.21/1.22/1.23.
- Nightly workflow runs longer fuzz and uploads benchmark results (non-blocking).

### Optional fuzzing (Go 1.18+)

```bash
go test -fuzz=Fuzz -fuzztime=10s
```

## License
MIT (or your preferred license).
