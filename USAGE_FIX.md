# How to Fix Your JSON Array Issue

## The Problem
Your JSON is an **array** `[{}, {}, {}]` but you're decoding it into a **single struct**. This only gives you the first element with incomplete data.

## The Solution

### ✅ Correct Usage (Array → Slice)
```go
package main

import (
    "fmt"
    databridge "github.com/dataBridgeGoPkg/dataBridge"
)

type UserDetails struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

type DemographicDetails struct {
    Age         int    `json:"age"`
    Nationality string `json:"nationality"`
}

type PersonalDetails struct {
    UserDetails        UserDetails        `json:"user_details"`
    DemographicDetails DemographicDetails `json:"demographic_details"`
    EmailIDs           []string           `json:"email_ids"`
    PhoneNumbers       []string           `json:"phone_numbers"`
    Misc               map[string]interface{} `json:"misc,omitempty"`
    Notes              interface{}        `json:"notes,omitempty"`
}

func main() {
    jsonArray := `[/* your JSON array */]`
    
    // Method 1: TransformToStructUniversal
    var people []PersonalDetails
    err := databridge.TransformToStructUniversal(jsonArray, &people)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Got %d people\n", len(people))
    
    // Method 2: Generic Transform (shorter)
    people2, err := databridge.Transform[[]PersonalDetails](jsonArray)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Got %d people\n", len(people2))
}
```

### ❌ What You Were Doing Wrong
```go
// WRONG: Array into single struct
var singlePerson PersonalDetails
err := databridge.TransformToStructUniversal(jsonArray, &singlePerson)
// This only gives you the first element with missing data
```

## Key Points

1. **"frist_name" typo**: Your first record has `"frist_name"` instead of `"first_name"`, so it maps to empty string
2. **Mixed phone arrays**: Numbers and objects in phone arrays are properly coerced to strings
3. **Key normalization**: Various formats (User.Details, USER_DETAILS, user_details) all map correctly
4. **Target type matters**: Array JSON → use slice target; Object JSON → use struct target

## Run the Demo
```bash
cd demo
go run main.go
```

This shows both wrong and correct usage with your exact data!
