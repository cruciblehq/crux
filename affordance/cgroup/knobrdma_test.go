package cgroup

import (
	"errors"
	"testing"
)

func TestParseRDMA(t *testing.T) {
	got, err := parseRDMA("mlx5_0 hca_handle=4 hca_object=16")
	if err != nil {
		t.Fatal(err)
	}
	if got != (rdma{Device: "mlx5_0", HcaHandle: 4, HcaObject: 16}) {
		t.Fatalf("got %+v", got)
	}
}

func TestParseRDMAWithoutLimits(t *testing.T) {
	got, err := parseRDMA("mlx5_0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Device != "mlx5_0" || got.HcaHandle != 0 || got.HcaObject != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseRDMARequiresDevice(t *testing.T) {
	_, err := parseRDMA("")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseRDMAUnknownKey(t *testing.T) {
	_, err := parseRDMA("mlx5_0 bogus=1")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestRDMAEqualByDevice(t *testing.T) {
	a := rdma{Device: "mlx5_0", HcaHandle: 1}
	if !a.equal(rdma{Device: "mlx5_0", HcaHandle: 99}) {
		t.Fatal("same device should be equal")
	}
}

func TestRDMACheckRejectsDivergence(t *testing.T) {
	a := rdma{Device: "mlx5_0", HcaHandle: 1}
	b := rdma{Device: "mlx5_0", HcaHandle: 2}
	if err := a.check(b); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestRDMAMergeNoOp(t *testing.T) {
	a := rdma{Device: "mlx5_0", HcaHandle: 1}
	if a.merge(a) {
		t.Fatal("merge reported change")
	}
}

func TestRDMACheckDifferentDeviceNoConflict(t *testing.T) {
	a := rdma{Device: "mlx5_0", HcaHandle: 1}
	b := rdma{Device: "mlx5_1", HcaHandle: 999}
	if err := a.check(b); err != nil {
		t.Fatalf("err = %v", err)
	}
}
