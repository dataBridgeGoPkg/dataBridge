package databridge

import (
	"net/url"
	"testing"
)

type Person struct {
	FirstName string `json:"First_Name"`
	Age       int64  `json:"Age"`
	Active    bool   `json:"Active"`
	Address   struct {
		City string `json:"city"`
	} `json:"address"`
}

func TestJSONToStruct(t *testing.T) {
	var p Person
	in := `{"first-name":"Atmadeep","Age":"30","active":"true","address":{"City":"Paris"}}`
	if err := TransformToStructUniversal(in, &p); err != nil {
		t.Fatalf("json parse failed: %v", err)
	}
	if p.FirstName != "Atmadeep" || p.Age != 30 || p.Active != true || p.Address.City != "Paris" {
		t.Fatalf("unexpected parsed result: %+v", p)
	}
}

func TestFormDottedToStruct(t *testing.T) {
	var p Person
	f := url.Values{"First_Name": {"Atmadeep"}, "Age": {"30"}, "Active": {"true"}, "address.city": {"Lyon"}}
	if err := TransformToStructUniversal(f, &p); err != nil {
		t.Fatalf("form parse failed: %v", err)
	}
	if p.Address.City != "Lyon" || p.FirstName != "Atmadeep" {
		t.Fatalf("unexpected parsed form result: %+v", p)
	}
}

func TestYAMLToStruct(t *testing.T) {
	var p Person
	y := `
First_Name: Atmadeep
Age: 30
Active: true
address:
  city: Marseille
`
	if err := TransformToStructUniversal(y, &p, WithYAML(true)); err != nil {
		t.Fatalf("yaml parse failed: %v", err)
	}
	if p.Address.City != "Marseille" {
		t.Fatalf("yaml mapping failed: %+v", p)
	}
}

func TestCSVToSlice(t *testing.T) {
	type Small struct {
		Name string `json:"name"`
		Age  int64  `json:"age"`
	}
	csvStr := "name,age\nAlice,30\nBob,25\n"
	var out []Small
	if err := TransformToStructUniversal(csvStr, &out); err != nil {
		t.Fatalf("csv->slice failed: %v", err)
	}
	if len(out) != 2 || out[0].Name != "Alice" || out[1].Age != 25 {
		t.Fatalf("csv content mismatch: %+v", out)
	}
}

func TestStrictModeUnknownField(t *testing.T) {
	type T struct {
		A string `json:"a"`
	}
	var out T
	err := TransformToStructUniversal(`{"a":"x","b":"y"}`, &out, WithStrict(true))
	if err == nil {
		t.Fatalf("expected strict mode to error on unknown field")
	}
}
