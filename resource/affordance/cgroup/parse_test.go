package cgroup

import (
	"errors"
	"testing"
)

func TestParseMajorMinorBare(t *testing.T) {
	maj, min, rest, err := parseMajorMinor("8 0")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if maj != 8 || min != 0 || rest != "" {
		t.Fatalf("got (%d,%d,%q), want (8,0,\"\")", maj, min, rest)
	}
}

func TestParseMajorMinorWithRest(t *testing.T) {
	maj, min, rest, err := parseMajorMinor("8 16 rbps=1024 wbps=2048")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if maj != 8 || min != 16 || rest != "rbps=1024 wbps=2048" {
		t.Fatalf("got (%d,%d,%q)", maj, min, rest)
	}
}

func TestParseMajorMinorMissing(t *testing.T) {
	for _, in := range []string{"", "8", "abc def", "8:16"} {
		t.Run(in, func(t *testing.T) {
			_, _, _, err := parseMajorMinor(in)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v, want ErrInvalidGrant", err)
			}
		})
	}
}

func TestParseMajorMinorOverflow(t *testing.T) {
	_, _, _, err := parseMajorMinor("4294967296 0")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestParseArgsRouting(t *testing.T) {
	var a, b uint64
	err := parseArgs("rbps=1000 wbps=2000", map[string]func(string) error{
		"rbps": func(v string) error { return parseUint64(&a, v) },
		"wbps": func(v string) error { return parseUint64(&b, v) },
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if a != 1000 || b != 2000 {
		t.Fatalf("a=%d b=%d", a, b)
	}
}

func TestParseArgsEmpty(t *testing.T) {
	if err := parseArgs("", map[string]func(string) error{}); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestParseArgsRejectsDuplicates(t *testing.T) {
	err := parseArgs("rbps=1 rbps=2", map[string]func(string) error{
		"rbps": func(string) error { return nil },
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestParseArgsRejectsUnknownKey(t *testing.T) {
	err := parseArgs("foo=1", map[string]func(string) error{})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestParseArgsRejectsMalformed(t *testing.T) {
	for _, in := range []string{"foo", "=1", "foo=", "x"} {
		t.Run(in, func(t *testing.T) {
			err := parseArgs(in, map[string]func(string) error{"foo": func(string) error { return nil }})
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v, want ErrInvalidGrant", err)
			}
		})
	}
}

func TestParseArgsHandlerErrorPropagated(t *testing.T) {
	sentinel := errors.New("handler boom")
	err := parseArgs("k=v", map[string]func(string) error{"k": func(string) error { return sentinel }})
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want sentinel", err)
	}
}

func TestParseUint64BasePrefixes(t *testing.T) {
	cases := map[string]uint64{"42": 42, "0x10": 16, "0o10": 8, "0b1010": 10}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			var got uint64
			if err := parseUint64(&got, in); err != nil {
				t.Fatalf("err = %v", err)
			}
			if got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
		})
	}
}

func TestParseUint64Invalid(t *testing.T) {
	var dst uint64
	if err := parseUint64(&dst, "abc"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseUint32Overflow(t *testing.T) {
	var dst uint32
	if err := parseUint32(&dst, "4294967296"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseUint16Valid(t *testing.T) {
	var dst uint16
	if err := parseUint16(&dst, "65535"); err != nil || dst != 65535 {
		t.Fatalf("dst=%d err=%v", dst, err)
	}
}

func TestParseFloat64(t *testing.T) {
	var dst float64
	if err := parseFloat64(&dst, "3.5"); err != nil || dst != 3.5 {
		t.Fatalf("dst=%v err=%v", dst, err)
	}
	if err := parseFloat64(&dst, "nope"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseBool(t *testing.T) {
	cases := map[string]bool{"true": true, "1": true, "false": false, "0": false}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			var dst bool
			if err := parseBool(&dst, in); err != nil {
				t.Fatalf("err = %v", err)
			}
			if dst != want {
				t.Fatalf("dst = %v, want %v", dst, want)
			}
		})
	}
	for _, in := range []string{"True", "yes", "no", "", "2"} {
		t.Run("invalid_"+in, func(t *testing.T) {
			var dst bool
			if err := parseBool(&dst, in); !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v, want ErrInvalidGrant", err)
			}
		})
	}
}
