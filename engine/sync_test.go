package engine_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alex-vit/plop/engine"

	"github.com/syncthing/syncthing/lib/protocol"
)

func TestSync(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e sync test in short mode")
	}

	homeA := t.TempDir()
	homeB := t.TempDir()
	syncA := t.TempDir()
	syncB := t.TempDir()

	// Pre-generate certs so we know device IDs before creating engines.
	certA, err := engine.GenerateCert(filepath.Join(homeA, "cert.pem"), filepath.Join(homeA, "key.pem"))
	if err != nil {
		t.Fatalf("GenerateCert A: %v", err)
	}
	certB, err := engine.GenerateCert(filepath.Join(homeB, "cert.pem"), filepath.Join(homeB, "key.pem"))
	if err != nil {
		t.Fatalf("GenerateCert B: %v", err)
	}

	idA := engine.DeviceID(certA)
	idB := engine.DeviceID(certB)
	t.Logf("Device A: %s", idA)
	t.Logf("Device B: %s", idB)

	// Let New() auto-init configs with peers baked in.
	engA, err := engine.New(homeA, syncA, []protocol.DeviceID{idB})
	if err != nil {
		t.Fatalf("engine.New A: %v", err)
	}
	t.Cleanup(engA.Stop)

	engB, err := engine.New(homeB, syncB, []protocol.DeviceID{idA})
	if err != nil {
		t.Fatalf("engine.New B: %v", err)
	}
	t.Cleanup(engB.Stop)

	if err := engA.Start(); err != nil {
		t.Fatalf("Start A: %v", err)
	}
	if err := engB.Start(); err != nil {
		t.Fatalf("Start B: %v", err)
	}

	// Give LAN discovery time to broadcast addresses.
	time.Sleep(2 * time.Second)

	// Write test file on A.
	testContent := []byte("hello from plop")
	if err := os.WriteFile(filepath.Join(syncA, "test.txt"), testContent, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Poll for the file to appear on B.
	deadline := time.After(2 * time.Minute)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	target := filepath.Join(syncB, "test.txt")
	for {
		select {
		case <-deadline:
			// Dump syncB contents for diagnostics.
			entries, _ := os.ReadDir(syncB)
			t.Logf("syncB contents (%d entries):", len(entries))
			for _, e := range entries {
				t.Logf("  %s", e.Name())
			}
			t.Fatalf("file did not sync within 2 minutes")

		case <-ticker.C:
			data, err := os.ReadFile(target)
			if err != nil {
				continue
			}
			if string(data) == string(testContent) {
				t.Log("file synced successfully")
				return
			}
		}
	}
}
