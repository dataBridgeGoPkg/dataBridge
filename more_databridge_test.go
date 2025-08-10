package databridge

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNestedStructsAndSlices(t *testing.T) {
	type Address struct {
		City string `json:"city"`
		Zip  int    `json:"zip"`
	}
	type Profile struct {
		Bio string `json:"bio"`
	}
	type User struct {
		Name      string    `json:"name"`
		Tags      []string  `json:"tags"`
		Addresses []Address `json:"addresses"`
		Profile   Profile   `json:"profile"`
	}
	in := `{"Name":"Ada","tags":["dev","go"],"addresses":[{"city":"Paris","zip":"75000"},{"city":"Lyon","zip":69000}],"profile":{"Bio":"hi"}}`
	var u User
	if err := TransformToStructUniversal(in, &u); err != nil {
		t.Fatalf("nested structs parse failed: %v", err)
	}
	if u.Name != "Ada" || len(u.Tags) != 2 || len(u.Addresses) != 2 || u.Addresses[0].Zip != 75000 || u.Profile.Bio != "hi" {
		t.Fatalf("unexpected nested result: %+v", u)
	}
}

func TestCustomKeyNormalizer(t *testing.T) {
	norm := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, " ", "_")
		return defaultNormalizeRe.ReplaceAllString(s, "")
	}
	type S struct {
		UserName string `json:"user_name"`
	}
	in := `{"USER NAME":"Bob"}`
	got, err := Transform[S](in, WithKeyNormalizer(norm))
	if err != nil {
		t.Fatalf("custom normalizer failed: %v", err)
	}
	if got.UserName != "Bob" {
		t.Fatalf("want Bob got %+v", got)
	}
}

func TestStrictModeNestedUnknowns(t *testing.T) {
	type Inner struct {
		B int `json:"b"`
	}
	type T struct {
		A Inner `json:"a"`
	}
	var out T
	err := TransformToStructUniversal(`{"a":{"b":1},"c":2}`, &out, WithStrict(true))
	if err == nil {
		t.Fatalf("expected strict mode error for unknown field c")
	}
	err = TransformToStructUniversal(`{"a":{"x":1}}`, &out, WithStrict(true))
	if err == nil {
		t.Fatalf("expected strict mode error for unknown nested field a.x")
	}
}

func TestURLValuesMultipleValuesSlices(t *testing.T) {
	type S1 struct {
		Tags []string `json:"tags"`
	}
	vals := url.Values{"tags": {"a", "b", "c"}}
	s1, err := Transform[S1](vals)
	if err != nil || len(s1.Tags) != 3 {
		t.Fatalf("url.Values to []string failed: %+v, err=%v", s1, err)
	}
	type S2 struct {
		Nums []int `json:"nums"`
	}
	vals2 := url.Values{"nums": {"1", "2", "3"}}
	s2, err := Transform[S2](vals2)
	if err != nil || len(s2.Nums) != 3 || s2.Nums[2] != 3 {
		t.Fatalf("url.Values to []int failed: %+v, err=%v", s2, err)
	}
}

func TestCSVWithDottedHeadersToNested(t *testing.T) {
	type S struct {
		User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"user"`
		Address struct {
			City string `json:"city"`
		} `json:"address"`
	}
	csv := "user.name,user.age,address.city\nAda,30,Paris\n"
	var arr []S
	if err := TransformToStructUniversal(csv, &arr); err != nil {
		t.Fatalf("csv dotted headers failed: %v", err)
	}
	if len(arr) != 1 || arr[0].User.Name != "Ada" || arr[0].User.Age != 30 || arr[0].Address.City != "Paris" {
		t.Fatalf("unexpected csv nested: %+v", arr)
	}
}

func TestReaderInputs(t *testing.T) {
	type S struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}
	r := strings.NewReader("name=Bob&age=40&active=true")
	var s S
	if err := TransformToStructUniversal(r, &s); err != nil {
		t.Fatalf("reader input form failed: %v", err)
	}
	if s.Name != "Bob" || s.Age != 40 || !s.Active {
		t.Fatalf("unexpected reader form: %+v", s)
	}
	b := bytes.NewBufferString(`{"name":"Ann","age":"22","active":"true"}`)
	if err := TransformToStructUniversal(b, &s); err != nil {
		t.Fatalf("buffer input json failed: %v", err)
	}
	if s.Name != "Ann" || s.Age != 22 || !s.Active {
		t.Fatalf("unexpected buffer json: %+v", s)
	}
}

func TestStructAsInput(t *testing.T) {
	type In struct {
		V string `json:"value"`
	}
	type Out struct {
		V string `json:"value"`
	}
	in := In{V: "ok"}
	var out Out
	if err := TransformToStructUniversal(in, &out); err != nil || out.V != "ok" {
		t.Fatalf("struct input failed: %+v, err=%v", out, err)
	}
}

func TestOutputPointerRequirement(t *testing.T) {
	type S struct {
		A int `json:"a"`
	}
	var s S
	if err := TransformToStructUniversal(`{"a":1}`, s); err == nil {
		t.Fatalf("expected error for non-pointer output")
	}
	var sp *S
	if err := TransformToStructUniversal(`{"a":1}`, sp); err == nil {
		t.Fatalf("expected error for nil pointer output")
	}
}

func TestTransformToJSONNilOutputPtr(t *testing.T) {
	j, err := TransformToJSON("name=Bob&age=40", nil)
	if err != nil {
		t.Fatalf("TransformToJSON nil output failed: %v", err)
	}
	var m map[string]interface{}
	if uerr := json.Unmarshal(j, &m); uerr != nil {
		t.Fatalf("unmarshal result json failed: %v", uerr)
	}
	if m["name"].(string) != "Bob" {
		t.Fatalf("unexpected json map: %v", m)
	}
}

func TestCSVStructTargetUsesFirstRow(t *testing.T) {
	type S struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	csv := "name,age\nAlice,30\nBob,25\n"
	var s S
	if err := TransformToStructUniversal(csv, &s); err != nil {
		t.Fatalf("csv struct target failed: %v", err)
	}
	if s.Name != "Alice" || s.Age != 30 {
		t.Fatalf("expected first row mapping, got %+v", s)
	}
}

func TestTimeParsingAndPointers(t *testing.T) {
	type S struct {
		When time.Time `json:"when"`
		Note *string   `json:"note"`
		N    *int      `json:"n"`
	}
	in := `{"when":"2024-12-01T10:11:12Z","note":"hello","n":"5"}`
	s, err := Transform[S](in)
	if err != nil {
		t.Fatalf("time/pointer parse failed: %v", err)
	}
	if s.When.IsZero() || s.Note == nil || *s.Note != "hello" || s.N == nil || *s.N != 5 {
		t.Fatalf("unexpected time/pointer result: %+v", s)
	}
	in2 := `{"note":""}`
	s2, err := Transform[S](in2)
	if err != nil {
		t.Fatalf("empty pointer check failed: %v", err)
	}
	if s2.Note != nil {
		t.Fatalf("expected nil note pointer for empty string, got %+v", s2.Note)
	}
}

func TestBufferToJSONMap(t *testing.T) {
	buf := bytes.NewBufferString(`{"a":1,"b":"x"}`)
	var m map[string]interface{}
	if err := TransformToStructUniversal(buf, &m); err != nil {
		t.Fatalf("buffer->map failed: %v", err)
	}
	if m["a"].(float64) != 1 || m["b"].(string) != "x" {
		t.Fatalf("unexpected buffer map: %+v", m)
	}
	// buffer->struct
	type S struct {
		A int    `json:"a"`
		B string `json:"b"`
	}
	var s S
	buf2 := bytes.NewBufferString(`{"a":"2","b":"y"}`)
	if err := TransformToStructUniversal(buf2, &s); err != nil {
		t.Fatalf("buffer->struct failed: %v", err)
	}
	if s.A != 2 || s.B != "y" {
		t.Fatalf("unexpected buffer struct: %+v", s)
	}
}

func TestJSONToJSONIdentityAndHeavy(t *testing.T) {
	// identity map
	in := `{"x":1,"y":[1,2,3],"z":{"a":true}}`
	j, err := TransformToJSON(in, nil)
	if err != nil {
		t.Fatalf("json->json(map) failed: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(j, &out); err != nil {
		t.Fatalf("unmarshal back failed: %v", err)
	}
	if _, ok := out["z"].(map[string]interface{}); !ok {
		t.Fatalf("expected nested map in identity json: %v", out)
	}

	// heavy nested
	deep := `{"a":{"b":{"c":{"d":{"e":{"f":{"g":{"h":{"i":{"j":10}}}}}}}}}}`
	type Deep struct {
		A struct {
			B struct {
				C struct {
					D struct {
						E struct {
							F struct {
								G struct {
									H struct {
										I struct {
											J int `json:"j"`
										} `json:"i"`
									} `json:"h"`
								} `json:"g"`
							} `json:"f"`
						} `json:"e"`
					} `json:"d"`
				} `json:"c"`
			} `json:"b"`
		} `json:"a"`
	}
	var dv Deep
	if err := TransformToStructUniversal(deep, &dv); err != nil {
		t.Fatalf("heavy nested json failed: %v", err)
	}
	if dv.A.B.C.D.E.F.G.H.I.J != 10 {
		t.Fatalf("wrong deep value: %+v", dv)
	}

	// array of objects
	arr := `[{"name":"a","n":"1"},{"name":"b","n":2}]`
	type Row struct {
		Name string `json:"name"`
		N    int    `json:"n"`
	}
	var rows []Row
	if err := TransformToStructUniversal(arr, &rows); err != nil {
		t.Fatalf("json array->slice failed: %v", err)
	}
	if len(rows) != 2 || rows[0].N != 1 || rows[1].Name != "b" {
		t.Fatalf("unexpected rows: %+v", rows)
	}

	// json->json for array
	j2, err := TransformToJSON(arr, &[]Row{})
	if err != nil {
		t.Fatalf("json array->json failed: %v", err)
	}
	var rows2 []Row
	if err := json.Unmarshal(j2, &rows2); err != nil || len(rows2) != 2 {
		t.Fatalf("json roundtrip array failed: len=%d err=%v", len(rows2), err)
	}
}

func TestLargeArrayAndMixedTypes(t *testing.T) {
	// simulate a larger payload
	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < 50; i++ { // modest number to keep tests fast
		if i > 0 {
			sb.WriteString(",")
		}
		if i%2 == 0 {
			// string types
			sb.WriteString(`{"id":"`)
			sb.WriteString(strconv.Itoa(i))
			sb.WriteString(`","flag":"true","score":"1.25"}`)
		} else {
			// native types
			sb.WriteString(`{"id":`)
			sb.WriteString(strconv.Itoa(i))
			sb.WriteString(`,"flag":false,"score":2.5}`)
		}
	}
	sb.WriteString("]")

	type Item struct {
		ID    int     `json:"id"`
		Flag  bool    `json:"flag"`
		Score float64 `json:"score"`
	}
	var items []Item
	if err := TransformToStructUniversal(sb.String(), &items); err != nil {
		t.Fatalf("large array mixed types failed: %v", err)
	}
	if len(items) != 50 {
		t.Fatalf("want 50 got %d", len(items))
	}
	if !items[0].Flag || items[1].Flag {
		t.Fatalf("flag coercion failed: first=%v second=%v", items[0].Flag, items[1].Flag)
	}
	if items[0].ID != 0 || items[1].ID != 1 {
		t.Fatalf("id coercion failed: first=%d second=%d", items[0].ID, items[1].ID)
	}
	if items[0].Score != 1.25 || items[1].Score != 2.5 {
		t.Fatalf("score coercion failed: first=%v second=%v", items[0].Score, items[1].Score)
	}

	// json round-trip using TransformToJSON with explicit output type
	j, err := TransformToJSON(sb.String(), &[]Item{})
	if err != nil {
		t.Fatalf("json roundtrip build failed: %v", err)
	}
	var items2 []Item
	if err := json.Unmarshal(j, &items2); err != nil || len(items2) != 50 {
		t.Fatalf("json roundtrip array failed: len=%d err=%v", len(items2), err)
	}
}

func TestStrictModeUnknownFieldInArray(t *testing.T) {
	arr := `[{"name":"a","n":1,"unknown":1}]`
	type Row struct {
		Name string `json:"name"`
		N    int    `json:"n"`
	}
	var rows []Row
	err := TransformToStructUniversal(arr, &rows, WithStrict(true))
	if err == nil {
		t.Fatalf("expected strict error for unknown field in array element")
	}
}

func TestCSVQuotedFieldsAndBOM(t *testing.T) {
	// Note: leading BOM before header
	csv := "\ufeffname,desc\n\"Doe, John\",\"Line1\nLine2\"\n"
	type Rec struct {
		Name string `json:"name"`
		Desc string `json:"desc"`
	}
	var out []Rec
	if err := TransformToStructUniversal(csv, &out); err != nil {
		t.Fatalf("csv quoted/BOM failed: %v", err)
	}
	if len(out) != 1 || out[0].Name != "Doe, John" || !strings.Contains(out[0].Desc, "Line1") || !strings.Contains(out[0].Desc, "Line2") {
		t.Fatalf("unexpected csv quoted/BOM result: %+v", out)
	}
}

func TestFormArrayBracketKeys(t *testing.T) {
	vals := url.Values{
		"tags[]": {"a", "b", "c"},
	}
	type S struct {
		Tags []string `json:"tags"`
	}
	s, err := Transform[S](vals)
	if err != nil {
		t.Fatalf("form [] keys failed: %v", err)
	}
	if len(s.Tags) != 3 || s.Tags[0] != "a" || s.Tags[2] != "c" {
		t.Fatalf("unexpected tags: %+v", s.Tags)
	}
}

func TestMapWithNaNFloatReturnsError(t *testing.T) {
	type S struct {
		Score float64 `json:"score"`
	}
	m := map[string]interface{}{"score": math.NaN()}
	var s S
	err := TransformToStructUniversal(m, &s)
	if err == nil {
		t.Fatalf("expected error for NaN float JSON marshal")
	}
}

func TestCSVDuplicateHeadersAndUnevenRows(t *testing.T) {
	csv := "a,a,b\r\na1,a2,b1\r\naonly,,b2,EXTRA\r\n"
	type R struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	var out []R
	if err := TransformToStructUniversal(csv, &out); err != nil {
		t.Fatalf("csv dup/uneven failed: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("want 2 rows got %d", len(out))
	}
	// duplicate header 'a' should take the last column's value per row
	if out[0].A != "a2" || out[0].B != "b1" {
		t.Fatalf("row0 mismatch: %+v", out[0])
	}
	// second row has extra value ignored and missing second 'a' becomes empty
	if out[1].A != "" || out[1].B != "b2" {
		t.Fatalf("row1 mismatch: %+v", out[1])
	}
}

func TestConcurrentTransforms(t *testing.T) {
	type S struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}
	inputs := []interface{}{
		`{"name":"A","age":"20","active":"true"}`,
		url.Values{"name": {"B"}, "age": {"30"}, "active": {"false"}},
		"name,age,active\nC,25,true\n",
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			in := inputs[i%len(inputs)]
			var s S
			_ = TransformToStructUniversal(in, &s)
			_, _ = Transform[S](in)
		}(i)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("concurrent transforms timed out")
	}
}

func TestVeryDeepNestedMapJSON(t *testing.T) {
	depth := 25
	// build {"a":{"a":{...{"val":42}}}}
	b := strings.Builder{}
	for i := 0; i < depth; i++ {
		b.WriteString(`{"a":`)
	}
	b.WriteString(`{"val":42}`)
	for i := 0; i < depth; i++ {
		b.WriteString("}")
	}
	// Use TransformToJSON then navigate
	data, err := TransformToJSON(b.String(), nil)
	if err != nil {
		t.Fatalf("deep json to json failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal deep json failed: %v", err)
	}
	cur := m
	for i := 0; i < depth; i++ {
		next, ok := cur["a"].(map[string]interface{})
		if !ok {
			t.Fatalf("level %d missing 'a'", i)
		}
		cur = next
	}
	if int(cur["val"].(float64)) != 42 {
		t.Fatalf("unexpected deep val: %v", cur["val"])
	}
}

func TestGenericTransformPointerTarget(t *testing.T) {
	type S struct {
		A int `json:"a"`
	}
	out, err := Transform[*S](`{"a":"5"}`)
	if err != nil {
		t.Fatalf("Transform[*S] failed: %v", err)
	}
	if out == nil || out.A != 5 {
		t.Fatalf("unexpected result: %#v", out)
	}
}

// failing reader for error path testing
type errReader struct{ once bool }

func (e *errReader) Read(p []byte) (int, error) {
	if !e.once {
		e.once = true
		copy(p, `{"x":1}`)
		return 5, nil
	}
	return 0, errors.New("boom")
}

func TestReaderErrorPropagation(t *testing.T) {
	var m map[string]interface{}
	if err := TransformToStructUniversal(&errReader{}, &m); err == nil {
		t.Fatalf("expected read error")
	}
}

func TestFormDottedKeyConflict(t *testing.T) {
	// user (string) collides with user.name (nested), nested should win and original flat value discarded
	vals := url.Values{
		"user":      {"root"},
		"user.name": {"Ada"},
		"user.age":  {"30"},
	}
	type User struct {
		User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"user"`
	}
	var u User
	if err := TransformToStructUniversal(vals, &u); err != nil {
		t.Fatalf("conflict dotted keys failed: %v", err)
	}
	if u.User.Name != "Ada" || u.User.Age != 30 {
		t.Fatalf("unexpected mapped user: %+v", u)
	}
}

func TestUnicodeKeysWithCustomNormalizer(t *testing.T) {
	// map café -> cafe using a custom normalizer
	in := `{"café":"au lait"}`
	type S struct {
		Cafe string `json:"cafe"`
	}
	norm := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "é", "e")
		return defaultNormalizeRe.ReplaceAllString(s, "")
	}
	s, err := Transform[S](in, WithKeyNormalizer(norm))
	if err != nil {
		t.Fatalf("unicode normalizer failed: %v", err)
	}
	if s.Cafe != "au lait" {
		t.Fatalf("unexpected unicode value: %+v", s)
	}
}

func TestVeryDeepNestedMapJSON_60(t *testing.T) {
	depth := 60
	b := strings.Builder{}
	for i := 0; i < depth; i++ {
		b.WriteString(`{"a":`)
	}
	b.WriteString(`{"val":99}`)
	for i := 0; i < depth; i++ {
		b.WriteString("}")
	}
	data, err := TransformToJSON(b.String(), nil)
	if err != nil {
		t.Fatalf("deep 60 json failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal deep 60 failed: %v", err)
	}
	cur := m
	for i := 0; i < depth; i++ {
		next, ok := cur["a"].(map[string]interface{})
		if !ok {
			t.Fatalf("level %d missing 'a'", i)
		}
		cur = next
	}
	if int(cur["val"].(float64)) != 99 {
		t.Fatalf("unexpected deep val: %v", cur["val"])
	}
}

func TestUserReportedPayloadMapping(t *testing.T) {
	// User provided JSON array with varied key casings, separators, and mixed-type arrays
	in := `[
        { "User.Details": { "First-Name": "Ada", "lastName": "Lovelace", "ignored": "x" }, "Demographic Details": { "AGE": "36", "Nationality": "British", "extra": true }, "email-ids": ["ada@example.com", "a.love@example.co.uk", ""], "phone_numbers": ["+44 20 7946 0018", 4479460018, { "type": "mobile", "value": "+44-7700-900123" }], "misc": { "notes": ["genius", null], "tags": ["math", "poet"] } },
        { "USER_DETAILS": { "first_name": "alan", "LAST_NAME": "Turing" }, "demographic-details": { "age": " 41 ", "NATIONALITY": "English" }, "EmailIds": ["alan@computing.org", "a.turing@bletchley.gov.uk"], "PHONE-NUMBERS": [ "+44-20-7000-0000", { "kind": "office", "ext": "123" }, 2070000000 ], "extra_top_level": { "foo": "bar" } },
        { "user_details": { "FIRST NAME": "grace", "last-name": "Hopper" }, "Demographic_Details": { "Age": 85, "nationality": "American" }, "EMAIL_IDS": [ "grace@example.mil", "ghopper@nvy.mil", " " ], "phone_numbers": [ "+1 (202) 555-0185", { "type": "fax", "value": "202-555-0199" }, 12025550185 ], "notes": "legend" }
    ]`

	type Name struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	type Demographic struct {
		Age         int    `json:"age"`
		Nationality string `json:"nationality"`
	}
	type Row struct {
		UserDetails        Name        `json:"user_details"`
		DemographicDetails Demographic `json:"demographic_details"`
		EmailIDs           []string    `json:"email_ids"`
		PhoneNumbers       []string    `json:"phone_numbers"`
		// allow passthrough extras without failing
		Misc map[string]interface{} `json:"misc"`
	}

	var out []Row
	if err := TransformToStructUniversal(in, &out); err != nil {
		t.Fatalf("failed to transform reported payload: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(out))
	}
	// Check first element mapping
	if out[0].UserDetails.FirstName != "Ada" || out[0].UserDetails.LastName != "Lovelace" {
		t.Fatalf("name mapping failed: %+v", out[0].UserDetails)
	}
	if out[0].DemographicDetails.Age != 36 || out[0].DemographicDetails.Nationality != "British" {
		t.Fatalf("demographic mapping failed: %+v", out[0].DemographicDetails)
	}
	if len(out[0].EmailIDs) < 2 || out[0].EmailIDs[0] != "ada@example.com" {
		t.Fatalf("email ids mapping failed: %+v", out[0].EmailIDs)
	}
	// Phone numbers should coerce to strings from mixed types
	if len(out[0].PhoneNumbers) < 3 {
		t.Fatalf("phone numbers length unexpected: %+v", out[0].PhoneNumbers)
	}
	if !strings.Contains(out[0].PhoneNumbers[2], "+44-7700-900123") {
		t.Fatalf("phone numbers coercion missing value field: %+v", out[0].PhoneNumbers)
	}
}

func TestUserActualPayloadWithTypo(t *testing.T) {
	// User's actual payload with "frist_name" typo
	actualPayload := `[
        { "User.Details": { "frist_name": "Ada", "last_name": "Lovelace", "ignored": "x" }, "Demographic Details": { "age": "36", "nationality": "British", "extra": true }, "email_ids": ["ada@example.com", "a.love@example.co.uk", ""], "phone_numbers": ["+44 20 7946 0018", 4479460018, { "type": "mobile", "value": "+44-7700-900123" }], "misc": { "notes": ["genius", null], "tags": ["math", "poet"] } },
        { "USER_DETAILS": { "first_name": "alan", "LAST_NAME": "Turing" }, "demographic-details": { "age": " 41 ", "NATIONALITY": "English" }, "EmailIds": ["alan@computing.org", "a.turing@bletchley.gov.uk"], "PHONE-NUMBERS": [ "+44-20-7000-0000", { "kind": "office", "ext": "123" }, 2070000000 ], "extra_top_level": { "foo": "bar" } },
        { "user_details": { "FIRST NAME": "grace", "last-name": "Hopper" }, "Demographic_Details": { "Age": 85, "nationality": "American" }, "EMAIL_IDS": [ "grace@example.mil", "ghopper@nvy.mil", " " ], "phone_numbers": [ "+1 (202) 555-0185", { "type": "fax", "value": "202-555-0199" }, 12025550185 ], "notes": "legend" }
    ]`

	type UserDetails struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	type DemographicDetails struct {
		Age         int    `json:"age"`
		Nationality string `json:"nationality"`
	}
	type PersonalDetails struct {
		UserDetails        UserDetails            `json:"user_details"`
		DemographicDetails DemographicDetails     `json:"demographic_details"`
		EmailIDs           []string               `json:"email_ids"`
		PhoneNumbers       []string               `json:"phone_numbers"`
		Misc               map[string]interface{} `json:"misc,omitempty"`
		Notes              interface{}            `json:"notes,omitempty"`
	}

	t.Run("CorrectUsage_ArrayToSlice", func(t *testing.T) {
		var people []PersonalDetails
		if err := TransformToStructUniversal(actualPayload, &people); err != nil {
			t.Fatalf("Array→Slice failed: %v", err)
		}
		if len(people) != 3 {
			t.Fatalf("expected 3 people, got %d", len(people))
		}
		// First person should have empty first_name due to "frist_name" typo
		if people[0].UserDetails.FirstName != "" {
			t.Logf("Note: first_name is '%s' - 'frist_name' typo in source didn't map", people[0].UserDetails.FirstName)
		}
		if people[0].UserDetails.LastName != "Lovelace" {
			t.Fatalf("expected Lovelace, got %s", people[0].UserDetails.LastName)
		}
		if people[0].DemographicDetails.Age != 36 {
			t.Fatalf("expected age 36, got %d", people[0].DemographicDetails.Age)
		}
		if len(people[0].PhoneNumbers) < 3 {
			t.Fatalf("expected 3+ phone numbers, got %d", len(people[0].PhoneNumbers))
		}
	})

	t.Run("IncorrectUsage_ArrayToSingleStruct", func(t *testing.T) {
		// This mimics the user's wrong usage - array into single struct
		var singlePerson PersonalDetails
		if err := TransformToStructUniversal(actualPayload, &singlePerson); err != nil {
			t.Fatalf("Array→Single failed: %v", err)
		}
		// Should get first element only
		if singlePerson.UserDetails.LastName != "Lovelace" {
			t.Fatalf("expected Lovelace from first element, got %s", singlePerson.UserDetails.LastName)
		}
		t.Logf("Single struct result: first_name='%s', last_name='%s', age=%d",
			singlePerson.UserDetails.FirstName, singlePerson.UserDetails.LastName,
			singlePerson.DemographicDetails.Age)
	})
}
