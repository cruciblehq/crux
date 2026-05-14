package manifest

import (
	"errors"
	"testing"
)

type encodeTestStruct struct {
	Value string `codec:"value"`
}

func TestEncodeToMapWithEncodable(t *testing.T) {
	a := &Affordance{}
	m, err := encodeToMap(a)
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Fatal("expected non-nil map")
	}
}

func TestEncodeToMapEncodableReturnsNonMap(t *testing.T) {
	g := &Grant{Source: ".cap effective net_admin"}
	_, err := encodeToMap(g)
	if !errors.Is(err, ErrEncodeFailed) {
		t.Fatalf("err = %v, want ErrEncodeFailed", err)
	}
}

func TestEncodeToMapNonEncodable(t *testing.T) {
	v := encodeTestStruct{Value: "hello"}
	m, err := encodeToMap(v)
	if err != nil {
		t.Fatal(err)
	}
	if m["value"] != "hello" {
		t.Fatalf("value = %v, want %q", m["value"], "hello")
	}
}

func TestMergeMapDisjoint(t *testing.T) {
	dst := map[string]any{"a": "1"}
	src := map[string]any{"b": "2"}
	got, err := mergeMap(dst, src)
	if err != nil {
		t.Fatal(err)
	}
	if got["a"] != "1" || got["b"] != "2" {
		t.Fatalf("got = %v", got)
	}
}

func TestMergeMapConflict(t *testing.T) {
	dst := map[string]any{"a": "1"}
	src := map[string]any{"a": "2"}
	_, err := mergeMap(dst, src)
	if !errors.Is(err, ErrEncodeFailed) {
		t.Fatalf("err = %v, want ErrEncodeFailed", err)
	}
}

func TestMergeMapEmpty(t *testing.T) {
	dst := map[string]any{}
	src := map[string]any{"x": "y"}
	got, err := mergeMap(dst, src)
	if err != nil {
		t.Fatal(err)
	}
	if got["x"] != "y" {
		t.Fatalf("got = %v", got)
	}
}
