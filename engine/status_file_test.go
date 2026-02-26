package engine

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestStatusFileWriterWritesUpdatesAndRemovesFileOnStop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, StatusFileName)

	var (
		mu       sync.RWMutex
		snapshot = StatusSnapshot{
			DeviceID:       "DEV-A",
			State:          StatusStateSynced,
			NeedTotalItems: 0,
			UpdatedAt:      time.Now().UTC(),
		}
	)

	updates := make(chan StatusSnapshot, 4)
	writer := newStatusFileWriter(path, updates, func() StatusSnapshot {
		mu.RLock()
		defer mu.RUnlock()
		return snapshot
	})
	writer.writeInterval = time.Hour
	writer.Start()
	t.Cleanup(writer.Stop)

	waitForStatusState(t, path, StatusStateSynced)

	mu.Lock()
	snapshot.State = StatusStateSyncing
	snapshot.NeedTotalItems = 3
	snapshot.UpdatedAt = time.Now().UTC()
	mu.Unlock()

	select {
	case updates <- StatusSnapshot{}:
	default:
		t.Fatal("failed to trigger status file write update")
	}

	waitForStatusState(t, path, StatusStateSyncing)

	writer.Stop()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("status file still exists after stop: err=%v", err)
	}
}

func TestWriteStatusSnapshotFileAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, StatusFileName)

	first := StatusSnapshot{
		DeviceID:       "DEV-A",
		State:          StatusStateStarting,
		NeedTotalItems: 0,
		UpdatedAt:      time.Now().UTC(),
	}
	second := StatusSnapshot{
		DeviceID:       "DEV-A",
		State:          StatusStateSynced,
		NeedTotalItems: 0,
		UpdatedAt:      time.Now().UTC().Add(time.Second),
	}

	if err := writeStatusSnapshotFile(path, first); err != nil {
		t.Fatalf("writeStatusSnapshotFile(first): %v", err)
	}
	if err := writeStatusSnapshotFile(path, second); err != nil {
		t.Fatalf("writeStatusSnapshotFile(second): %v", err)
	}

	got := readStatusSnapshotFromFile(t, path)
	if got.State != StatusStateSynced {
		t.Fatalf("state = %q, want %q", got.State, StatusStateSynced)
	}
}

func waitForStatusState(t *testing.T, path string, want StatusState) {
	t.Helper()

	deadline := time.After(5 * time.Second)
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for state %q in %s", want, path)
		case <-ticker.C:
			got := readStatusSnapshotFromFile(t, path)
			if got.State == want {
				return
			}
		}
	}
}

func readStatusSnapshotFromFile(t *testing.T, path string) StatusSnapshot {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		return StatusSnapshot{}
	}

	var snapshot StatusSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("unmarshal status snapshot from %s: %v", path, err)
	}
	return snapshot
}
