package cgroup

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseCPUSetPartitionValid(t *testing.T) {
	for _, in := range []partition{partitionMember, partitionRoot, partitionIsolated} {
		t.Run(string(in), func(t *testing.T) {
			got, err := parseCPUSetPartition(string(in))
			if err != nil || got != in {
				t.Fatalf("got %q err %v", got, err)
			}
		})
	}
}

func TestParseCPUSetPartitionInvalid(t *testing.T) {
	for _, in := range []string{"", "Member", "main"} {
		_, err := parseCPUSetPartition(in)
		if !errors.Is(err, ErrInvalidGrant) {
			t.Fatalf("%q: err = %v", in, err)
		}
	}
}

func TestParseIndexTokenSingle(t *testing.T) {
	got, err := parseIndexToken("3")
	if err != nil {
		t.Fatal(err)
	}
	if got != (indexRange{3, 3}) {
		t.Fatalf("got %+v", got)
	}
}

func TestParseIndexTokenRange(t *testing.T) {
	got, err := parseIndexToken("0-3")
	if err != nil {
		t.Fatal(err)
	}
	if got != (indexRange{0, 3}) {
		t.Fatalf("got %+v", got)
	}
}

func TestParseIndexTokenInvalid(t *testing.T) {
	for _, in := range []string{"", "abc", "1-", "-1", "1-2-3", " 1", "1 "} {
		t.Run(in, func(t *testing.T) {
			_, err := parseIndexToken(in)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestParseIndexTokenDescendingRangeRejected(t *testing.T) {
	_, err := parseIndexToken("5-3")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseIndexListNormalizes(t *testing.T) {
	cases := []struct {
		in   string
		want indexList
	}{
		{"0", indexList{{0, 0}}},
		{"0-3", indexList{{0, 3}}},
		{"0-3,5", indexList{{0, 3}, {5, 5}}},
		{"5,0-3", indexList{{0, 3}, {5, 5}}},               // sort
		{"0-2,3-5", indexList{{0, 5}}},                     // adjacent merge
		{"0-2,2-5", indexList{{0, 5}}},                     // overlap merge
		{"0-2,4,5,7-9", indexList{{0, 2}, {4, 5}, {7, 9}}}, // adjacency 4-5
		{"7-9,0-2,4,5", indexList{{0, 2}, {4, 5}, {7, 9}}}, // unsorted input
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			var got indexList
			if err := parseIndexList(&got, c.in); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestParseIndexListEmpty(t *testing.T) {
	var dst indexList
	if err := parseIndexList(&dst, ""); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
	if err := parseIndexList(&dst, "   "); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseIndexTokenStartOverflow(t *testing.T) {
	_, err := parseIndexToken("4294967296")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestParseIndexTokenEndOverflow(t *testing.T) {
	_, err := parseIndexToken("0-4294967296")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestParseIndexListInvalidToken(t *testing.T) {
	var dst indexList
	if err := parseIndexList(&dst, "0,abc,5"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestNormalizeIndexListContainedRange(t *testing.T) {
	got := normalizeIndexList(indexList{{0, 5}, {2, 3}})
	if len(got) != 1 || got[0] != (indexRange{0, 5}) {
		t.Fatalf("got %v, want [{0 5}]", got)
	}
}
