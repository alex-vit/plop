package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	writer := newStatusFileWriter(path, func() StatusSnapshot {
		mu.RLock()
		defer mu.RUnlock()
		return snapshot
	})
	writer.writeInterval = 20 * time.Millisecond
	writer.Start()
	t.Cleanup(writer.Stop)

	waitForStatusState(t, path, StatusStateSynced)

	mu.Lock()
	snapshot.State = StatusStateSyncing
	snapshot.NeedTotalItems = 3
	snapshot.UpdatedAt = time.Now().UTC()
	mu.Unlock()

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

func TestStatusFileWriterLogsFailureTransitionAndRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), StatusFileName)

	writer := newStatusFileWriter(path, func() StatusSnapshot {
		return StatusSnapshot{
			State:     StatusStateSynced,
			UpdatedAt: time.Now().UTC(),
		}
	})

	var logs []string
	writer.logf = func(format string, args ...any) {
		logs = append(logs, fmt.Sprintf(format, args...))
	}

	attempt := 0
	writer.writeFile = func(_ string, _ StatusSnapshot) error {
		attempt++
		if attempt <= 2 {
			return errors.New("disk full")
		}
		return nil
	}

	writeFailed := false
	lastWriteErr := ""

	writer.writeSnapshot(&writeFailed, &lastWriteErr)
	writer.writeSnapshot(&writeFailed, &lastWriteErr)
	writer.writeSnapshot(&writeFailed, &lastWriteErr)

	if len(logs) != 2 {
		t.Fatalf("log count = %d, want 2 (%v)", len(logs), logs)
	}
	if !strings.Contains(logs[0], "failed") {
		t.Fatalf("first log = %q, want failure message", logs[0])
	}
	if !strings.Contains(logs[1], "recovered") {
		t.Fatalf("second log = %q, want recovery message", logs[1])
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
