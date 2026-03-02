package main

import (
	"strings"
	"testing"
)

func TestRunUnknownCommand(t *testing.T) {
	err := run([]string{"bogus"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command: bogus") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunHelp(t *testing.T) {
	if err := run([]string{"--help"}); err != nil {
		t.Fatalf("--help returned error: %v", err)
	}
}

func TestRunSubcommandHelp(t *testing.T) {
	for _, cmd := range []string{"init", "pair", "run", "status", "id"} {
		t.Run(cmd, func(t *testing.T) {
			if err := run([]string{cmd, "--help"}); err != nil {
				t.Fatalf("%s --help returned error: %v", cmd, err)
			}
		})
	}
}

func TestRunHomeFlag(t *testing.T) {
	orig := homeDir
	defer func() { homeDir = orig }()

	// --home is parsed before dispatch; "id" will fail because the cert
	// doesn't exist at the temp path, but homeDir should be set.
	_ = run([]string{"--home", "/tmp/plop-test-home", "id"})
	if homeDir != "/tmp/plop-test-home" {
		t.Fatalf("homeDir = %q, want /tmp/plop-test-home", homeDir)
	}
}

func TestInitRejectsExtraArgs(t *testing.T) {
	err := runInit([]string{"extra"})
	if err == nil {
		t.Fatal("expected error for extra args")
	}
}

func TestStatusRejectsExtraArgs(t *testing.T) {
	err := runStatus([]string{"extra"})
	if err == nil {
		t.Fatal("expected error for extra args")
	}
}

func TestIDRejectsExtraArgs(t *testing.T) {
	err := runID([]string{"extra"})
	if err == nil {
		t.Fatal("expected error for extra args")
	}
}

func TestRunRejectsExtraArgs(t *testing.T) {
	err := runRun([]string{"folder1", "folder2"})
	if err == nil {
		t.Fatal("expected error for extra args")
	}
}

func TestPairRequiresArg(t *testing.T) {
	err := runPair([]string{})
	if err == nil {
		t.Fatal("expected error when no device ID given")
	}
}

func TestPairSyncthingAcceptsNoArg(t *testing.T) {
	orig := homeDir
	defer func() { homeDir = orig }()
	homeDir = t.TempDir()

	// --syncthing without a device ID should succeed (prints guide).
	err := runPair([]string{"--syncthing"})
	if err != nil {
		t.Fatalf("pair --syncthing returned error: %v", err)
	}
}

func TestPairSyncthingRejectsExtraArgs(t *testing.T) {
	err := runPair([]string{"--syncthing", "arg1", "arg2"})
	if err == nil {
		t.Fatal("expected error for too many args with --syncthing")
	}
}

func TestPairRejectsInvalidDeviceID(t *testing.T) {
	orig := homeDir
	defer func() { homeDir = orig }()
	homeDir = t.TempDir()

	err := runPair([]string{"not-a-valid-id"})
	if err == nil {
		t.Fatal("expected error for invalid device ID")
	}
	if !strings.Contains(err.Error(), "invalid device ID") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStringSlice(t *testing.T) {
	var s stringSlice
	if err := s.Set("a"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("b"); err != nil {
		t.Fatal(err)
	}
	if len(s) != 2 || s[0] != "a" || s[1] != "b" {
		t.Fatalf("got %v, want [a b]", s)
	}
	if s.String() != "a, b" {
		t.Fatalf("String() = %q, want %q", s.String(), "a, b")
	}
}
