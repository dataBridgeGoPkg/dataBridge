package databridge

import (
	"strings"
	"testing"
)

type benchUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Ok   bool   `json:"ok"`
}

func BenchmarkJSONToStruct(b *testing.B) {
	in := `{"name":"Alice","age":"30","ok":"true"}`
	for i := 0; i < b.N; i++ {
		var u benchUser
		_ = TransformToStructUniversal(in, &u)
	}
}

func BenchmarkFormToStruct(b *testing.B) {
	in := "name=Bob&age=40&ok=true"
	for i := 0; i < b.N; i++ {
		var u benchUser
		_ = TransformToStructUniversal(in, &u)
	}
}

func BenchmarkCSVToSlice(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("name,age,ok\n")
	for i := 0; i < 1000; i++ {
		sb.WriteString("A,30,true\n")
	}
	in := sb.String()
	for i := 0; i < b.N; i++ {
		var arr []benchUser
		_ = TransformToStructUniversal(in, &arr)
	}
}
