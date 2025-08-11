package main

import (
	"go/ast"
	"strings"
	"testing"
)

func TestSplitCSV(t *testing.T) {
	got := splitCSV("User, Order , , Item")
	want := []string{"User", "Order", "Item"}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("idx %d got %q want %q", i, got[i], want[i])
		}
	}
}

func TestContains(t *testing.T) {
	if !contains([]string{"A", "B"}, "B") {
		t.Fatal("expected contains to be true for B")
	}
	if contains([]string{"A", "B"}, "C") {
		t.Fatal("expected contains to be false for C")
	}
}

func TestJSONNameFromTag(t *testing.T) {
	if got := jsonNameFromTag(`json:"name,omitempty"`, "Name"); got != "name" {
		t.Fatalf("got %q want name", got)
	}
	if got := jsonNameFromTag(`json:"-"`, "Name"); got != "" {
		t.Fatalf("got %q want empty for -", got)
	}
	if got := jsonNameFromTag("", "Name"); got != "Name" {
		t.Fatalf("got %q want fallback Name", got)
	}
}

func TestCollectFieldsAndEmit(t *testing.T) {
	// Build an AST representing a struct:
	// type S struct { Name string `json:"name"`; Age int64 `json:"age"`; Tags []string `json:"tags"`; Address struct{ City string `json:"city"` } `json:"address"` }
	fl := &ast.FieldList{List: []*ast.Field{
		{Names: []*ast.Ident{{Name: "Name"}}, Type: &ast.Ident{Name: "string"}, Tag: &ast.BasicLit{Kind: 1, Value: "`json:\"name\"`"}},
		{Names: []*ast.Ident{{Name: "Age"}}, Type: &ast.Ident{Name: "int64"}, Tag: &ast.BasicLit{Kind: 1, Value: "`json:\"age\"`"}},
		{Names: []*ast.Ident{{Name: "Tags"}}, Type: &ast.ArrayType{Elt: &ast.Ident{Name: "string"}}, Tag: &ast.BasicLit{Kind: 1, Value: "`json:\"tags\"`"}},
		{Names: []*ast.Ident{{Name: "Address"}}, Type: &ast.StructType{Fields: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{{Name: "City"}}, Type: &ast.Ident{Name: "string"}, Tag: &ast.BasicLit{Kind: 1, Value: "`json:\"city\"`"}},
		}}}, Tag: &ast.BasicLit{Kind: 1, Value: "`json:\"address\"`"}},
	}}

	fields := collectFields(fl, "")
	if len(fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(fields))
	}
	// Verify names
	if fields[0].JSONName != "name" || fields[1].JSONName != "age" || !fields[2].IsSlice || fields[3].TypeExpr != "struct" {
		t.Fatalf("unexpected fields: %+v", fields)
	}

	var b strings.Builder
	emitFieldAssignments(&b, "out", "", fields)
	src := b.String()
	// Expect code for scalars
	if !strings.Contains(src, `if s := vals.Get("name"); s != ""`) || !strings.Contains(src, "out.Name = s") {
		t.Fatalf("missing scalar assignment for name in: %s", src)
	}
	if !strings.Contains(src, `if s := vals.Get("age"); s != ""`) || !strings.Contains(src, "ParseInt(") {
		t.Fatalf("missing int assignment for age in: %s", src)
	}
	// Expect slice handling
	if !strings.Contains(src, `if vs, ok := vals["tags"]; ok`) || !strings.Contains(src, "append(out.Tags") {
		t.Fatalf("missing slice handling for tags in: %s", src)
	}
	// Expect nested dotted key
	if !strings.Contains(src, `if s := vals.Get("address.city"); s != ""`) {
		t.Fatalf("missing nested dotted key for address.city in: %s", src)
	}
}
