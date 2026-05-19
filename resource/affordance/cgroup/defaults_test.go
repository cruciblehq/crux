package cgroup

import (
	"reflect"
	"testing"
)

func TestSetDefaultPrimitiveKinds(t *testing.T) {
	type leaf struct {
		B  bool   `default:"true"`
		U8 uint8  `default:"7"`
		U  uint64 `default:"42"`
		I  int16  `default:"-3"`
		S  string `default:"hello"`
	}
	v := leaf{}
	setDefaults(reflect.TypeFor[leaf](), reflect.ValueOf(&v).Elem())
	want := leaf{B: true, U8: 7, U: 42, I: -3, S: "hello"}
	if v != want {
		t.Fatalf("got %+v, want %+v", v, want)
	}
}

func TestSetDefaultPanicsOnUnsupportedKind(t *testing.T) {
	type bad struct {
		F float64 `default:"1.5"`
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unsupported kind")
		}
	}()
	v := bad{}
	setDefaults(reflect.TypeFor[bad](), reflect.ValueOf(&v).Elem())
}
