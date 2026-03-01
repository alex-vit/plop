package engine

import (
	"net"
	"strconv"
	"testing"

	stconfig "github.com/syncthing/syncthing/lib/config"
)

func TestEnsureRuntimeGUIAddressAssignsPort(t *testing.T) {
	cfg := stconfig.Configuration{}
	cfg.GUI.Enabled = true
	cfg.GUI.RawAddress = "127.0.0.1:0" //nolint:goconst

	if err := ensureRuntimeGUIAddress(&cfg); err != nil {
		t.Fatalf("ensureRuntimeGUIAddress: %v", err)
	}

	host, port, err := net.SplitHostPort(cfg.GUI.RawAddress)
	if err != nil {
		t.Fatalf("split host/port: %v", err)
	}
	if host != "127.0.0.1" { //nolint:goconst
		t.Fatalf("host = %q, want 127.0.0.1", host)
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("port parse: %v", err)
	}
	if portNum <= 0 {
		t.Fatalf("port = %d, want > 0", portNum)
	}
}

func TestEnsureRuntimeGUIAddressKeepsConfiguredPort(t *testing.T) {
	cfg := stconfig.Configuration{}
	cfg.GUI.Enabled = true
	cfg.GUI.RawAddress = "127.0.0.1:8384"

	if err := ensureRuntimeGUIAddress(&cfg); err != nil {
		t.Fatalf("ensureRuntimeGUIAddress: %v", err)
	}
	if cfg.GUI.RawAddress != "127.0.0.1:8384" {
		t.Fatalf("raw address = %q, want unchanged", cfg.GUI.RawAddress)
	}
}

func TestEnsureRuntimeGUIAddressNormalizesWildcardHost(t *testing.T) {
	cfg := stconfig.Configuration{}
	cfg.GUI.Enabled = true
	cfg.GUI.RawAddress = "0.0.0.0:0"

	if err := ensureRuntimeGUIAddress(&cfg); err != nil {
		t.Fatalf("ensureRuntimeGUIAddress: %v", err)
	}

	host, _, err := net.SplitHostPort(cfg.GUI.RawAddress)
	if err != nil {
		t.Fatalf("split host/port: %v", err)
	}
	if host != "127.0.0.1" {
		t.Fatalf("host = %q, want 127.0.0.1", host)
	}
}
