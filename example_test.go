package databridge_test

import (
	"fmt"
	"net/url"

	"github.com/atmadeep/databridge"
)

type Person struct {
	FirstName string `json:"First_Name"`
	Age       int64  `json:"Age"`
	Active    bool   `json:"Active"`
	Address   struct {
		City string `json:"city"`
	} `json:"address"`
}

func ExampleTransform() {
	in := `{"first-name":"Atmadeep","Age":"30","active":"true","address":{"City":"Paris"}}`
	p, _ := databridge.Transform[Person](in)
	fmt.Println(p.Address.City)
	// Output: Paris
}

func ExampleTransformToStructUniversal() {
	f := url.Values{"First_Name": {"Atmadeep"}, "Age": {"30"}, "Active": {"true"}, "address.city": {"Lyon"}}
	var p Person
	_ = databridge.TransformToStructUniversal(f, &p)
	fmt.Println(p.Address.City)
	// Output: Lyon
}
