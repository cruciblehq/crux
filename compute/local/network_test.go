package local

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestListBridges_ParsesBridgeNames(t *testing.T) {
	exec := func(_ context.Context, stdout, _ io.Writer, command string, args ...string) (int, error) {
		if command != ipCommand {
			t.Fatalf("command = %q, want %q", command, ipCommand)
		}
		_, _ = stdout.Write([]byte("1: lo: <LOOPBACK>\n2: br-main: <BROADCAST>\n3: br-other: <BROADCAST>\n"))
		return 0, nil
	}

	bridges, err := ListBridges(context.Background(), exec, "br-")
	if err != nil {
		t.Fatalf("ListBridges() error = %v", err)
	}
	if got, want := strings.Join(bridges, ","), "br-main,br-other"; got != want {
		t.Fatalf("ListBridges() = %q, want %q", got, want)
	}
}

func TestInterfaceAddress_ParsesInetField(t *testing.T) {
	exec := func(_ context.Context, stdout, _ io.Writer, command string, args ...string) (int, error) {
		if command != ipCommand {
			t.Fatalf("command = %q, want %q", command, ipCommand)
		}
		_, _ = stdout.Write([]byte("2: eth0    inet 10.0.0.2/24 brd 10.0.0.255 scope global dynamic eth0\n"))
		return 0, nil
	}

	addr, err := InterfaceAddress(context.Background(), exec, "eth0")
	if err != nil {
		t.Fatalf("InterfaceAddress() error = %v", err)
	}
	if addr != "10.0.0.2/24" {
		t.Fatalf("InterfaceAddress() = %q, want %q", addr, "10.0.0.2/24")
	}
}

func TestListBridgeInterfaces_ParsesNames(t *testing.T) {
	exec := func(_ context.Context, stdout, _ io.Writer, command string, args ...string) (int, error) {
		if command != ipCommand {
			t.Fatalf("command = %q, want %q", command, ipCommand)
		}
		_, _ = stdout.Write([]byte("3: ve-a: <BROADCAST>\n4: ve-b: <BROADCAST>\n"))
		return 0, nil
	}

	ifaces, err := ListBridgeInterfaces(context.Background(), exec, "br-main")
	if err != nil {
		t.Fatalf("ListBridgeInterfaces() error = %v", err)
	}
	if got, want := strings.Join(ifaces, ","), "ve-a,ve-b"; got != want {
		t.Fatalf("ListBridgeInterfaces() = %q, want %q", got, want)
	}
}
