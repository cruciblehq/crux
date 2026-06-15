package net

import "testing"

func TestIsPortBased(t *testing.T) {
	based := []string{protoTCP, protoUDP, protoSCTP, protoDCCP}
	for _, p := range based {
		if !isPortBased(p) {
			t.Errorf("isPortBased(%q) = false, want true", p)
		}
	}
	notBased := []string{protoICMP, protoICMPv6, protoGRE, protoESP, protoAH, protoIP, "89", "unknown"}
	for _, p := range notBased {
		if isPortBased(p) {
			t.Errorf("isPortBased(%q) = true, want false", p)
		}
	}
}

func TestIsPortless(t *testing.T) {
	portless := []string{protoICMP, protoICMPv6, protoGRE, protoESP, protoAH}
	for _, p := range portless {
		if !isPortless(p) {
			t.Errorf("isPortless(%q) = false, want true", p)
		}
	}
	notPortless := []string{protoTCP, protoUDP, protoSCTP, protoDCCP, protoIP, "89", "unknown"}
	for _, p := range notPortless {
		if isPortless(p) {
			t.Errorf("isPortless(%q) = true, want false", p)
		}
	}
}

func TestIsValidProtocol(t *testing.T) {
	valid := []string{
		protoTCP, protoUDP, protoSCTP, protoDCCP,
		protoICMP, protoICMPv6, protoGRE, protoESP, protoAH,
		protoIP, "0", "89", "255",
	}
	for _, p := range valid {
		if !IsValidProtocol(p) {
			t.Errorf("isValidProtocol(%q) = false, want true", p)
		}
	}
	invalid := []string{"", "all", "tcpx", "256", "-1", "0x10", "proto"}
	for _, p := range invalid {
		if IsValidProtocol(p) {
			t.Errorf("isValidProtocol(%q) = true, want false", p)
		}
	}
}
