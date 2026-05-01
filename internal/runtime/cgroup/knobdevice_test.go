package cgroup

import (
	"errors"
	"strings"
	"testing"
)

func TestParseDeviceTypeValid(t *testing.T) {
	for _, in := range []deviceType{deviceTypeChar, deviceTypeBlock, deviceTypeAll} {
		t.Run(string(in), func(t *testing.T) {
			got, err := parseDeviceType(string(in))
			if err != nil || got != in {
				t.Fatalf("got %q err %v", got, err)
			}
		})
	}
}

func TestParseDeviceTypeInvalid(t *testing.T) {
	for _, in := range []string{"", "x", "C"} {
		_, err := parseDeviceType(in)
		if !errors.Is(err, ErrInvalidGrant) {
			t.Fatalf("%q err = %v", in, err)
		}
	}
}

func TestParseDevice(t *testing.T) {
	got, err := parseDevice("c 8 0 rw")
	if err != nil {
		t.Fatal(err)
	}
	want := device{Type: deviceTypeChar, Major: 8, Minor: 0, Access: "rw"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseDeviceInvalid(t *testing.T) {
	for _, in := range []string{
		"",
		"c 8 0",     // missing access
		"c 8 0 xyz", // bad access bits
		"x 8 0 rw",  // bad type
		"c -1 0 r",  // negative
		"c 8",
	} {
		t.Run(in, func(t *testing.T) {
			_, err := parseDevice(in)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestDeviceEqualByIdentity(t *testing.T) {
	a := device{Type: deviceTypeChar, Major: 8, Minor: 0, Access: "r"}
	b := device{Type: deviceTypeChar, Major: 8, Minor: 0, Access: "wm"}
	if !a.equal(b) {
		t.Fatal("identity-equal entries should compare equal")
	}
	c := device{Type: deviceTypeBlock, Major: 8, Minor: 0, Access: "r"}
	if a.equal(c) {
		t.Fatal("different type should not be equal")
	}
}

func TestDeviceCheckAlwaysNil(t *testing.T) {
	a := device{Type: deviceTypeChar, Major: 8, Minor: 0, Access: "r"}
	b := device{Type: deviceTypeChar, Major: 8, Minor: 0, Access: "wm"}
	if err := a.check(b); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestDeviceMergeUnionsAndNormalizes(t *testing.T) {
	a := device{Type: deviceTypeChar, Major: 8, Minor: 0, Access: "r"}
	b := device{Type: deviceTypeChar, Major: 8, Minor: 0, Access: "wm"}
	if !a.merge(b) {
		t.Fatal("merge added bits but reported no change")
	}
	for _, c := range "rwm" {
		if !strings.ContainsRune(a.Access, c) {
			t.Errorf("access %q missing %c", a.Access, c)
		}
	}
}

func TestDeviceMergeIdempotent(t *testing.T) {
	a := device{Type: deviceTypeChar, Major: 8, Minor: 0, Access: "rw"}
	b := a
	if a.merge(b) {
		t.Fatal("merge of identical entries reported change")
	}
}
