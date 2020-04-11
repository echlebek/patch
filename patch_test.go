package patch_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/echlebek/patch"
)

type simpleResource struct {
	A int
	B string
}

type simpleWithTags struct {
	A int    `json:"a"`
	B string `json:"b"`
}

type resourceWithUnexported struct {
	a int `json:"a"`
	B int `json:"b"`
}

type resourceWithPointer struct {
	A *simpleWithTags
}

func TestPatchStruct(t *testing.T) {
	tests := []struct {
		Name     string
		Resource interface{}
		Patch    string
		Want     interface{}
		WantErr  bool
	}{
		{
			Name: "simple_populated",
			Resource: &simpleResource{
				A: 1,
				B: "2",
			},
			Patch: `{"A":2}`,
			Want: &simpleResource{
				A: 2,
				B: "2",
			},
		},
		{
			Name:     "simple_unpopulated",
			Resource: &simpleResource{},
			Patch:    `{"B":"BEE"}`,
			Want:     &simpleResource{B: "BEE"},
		},
		{
			Name: "simple_populated_with_tags",
			Resource: &simpleWithTags{
				A: 1,
				B: "2",
			},
			Patch: `{"a":2}`,
			Want: &simpleWithTags{
				A: 2,
				B: "2",
			},
		},
		{
			Name:     "simple_unpopulated_with_tags",
			Resource: &simpleWithTags{},
			Patch:    `{"b":"BEE"}`,
			Want:     &simpleWithTags{B: "BEE"},
		},
		{
			Name:     "empty_valid_patch",
			Resource: &simpleResource{B: "A"},
			Patch:    `{}`,
			Want:     &simpleResource{B: "A"},
		},
		{
			Name:     "not_a_struct",
			Resource: 5,
			Patch:    `{}`,
			WantErr:  true,
		},
		{
			Name:     "unaddressable struct",
			Resource: struct{}{},
			Patch:    `{}`,
			WantErr:  true,
		},
		{
			Name:     "unexported fields are ignored",
			Resource: &resourceWithUnexported{a: 5, B: 10},
			Patch:    `{"a":10,"b":20}`,
			Want:     &resourceWithUnexported{a: 5, B: 20},
		},
		{
			Name:     "null patch results in zero value",
			Resource: &simpleResource{A: 5, B: "5"},
			Patch:    `{"A":null}`,
			Want:     &simpleResource{B: "5"},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var msg map[string]*json.RawMessage
			if err := json.Unmarshal([]byte(test.Patch), &msg); err != nil {
				t.Fatal(err)
			}
			err := patch.Struct(&test.Resource, msg)
			if err != nil {
				if test.WantErr {
					return
				}
				t.Fatal(err)
			}
			if test.WantErr {
				t.Error("expected non-nil error")
			}
			if got, want := test.Resource, test.Want; !reflect.DeepEqual(got, want) {
				t.Errorf("bad PatchStruct: got %v, want %v", got, want)
			}
		})
	}
}

func TestInvalidRawMessage(t *testing.T) {
	r := simpleResource{A: 1}
	msg := map[string]*json.RawMessage{}
	if err := patch.Struct(&r, msg); err != nil {
		t.Fatal("buggy test")
	}
	badBytes := []byte("not json")
	msg["A"] = (*json.RawMessage)(&badBytes)
	if err := patch.Struct(&r, msg); err == nil {
		t.Fatal("want non-nil error")
	}
}

func BenchmarkPatchStruct(b *testing.B) {
	resource := &simpleWithTags{
		A: 1,
		B: "2",
	}
	ptch := `{"a":2}`
	var msg map[string]*json.RawMessage
	if err := json.Unmarshal([]byte(ptch), &msg); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = patch.Struct(resource, msg)
	}

}
