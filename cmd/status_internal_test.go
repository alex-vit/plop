package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alex-vit/plop/engine"
)

func TestReadInternalStatusFresh(t *testing.T) {
	home := t.TempDir()
	now := time.Date(2026, time.February, 26, 18, 0, 0, 0, time.UTC)

	snapshot := engine.StatusSnapshot{
		DeviceID:       "DEV-A",
		State:          engine.StatusStateSynced,
		NeedTotalItems: 0,
		UpdatedAt:      now.Add(-2 * time.Second),
	}
	writeInternalStatusFile(t, home, snapshot)

	got, err := readInternalStatus(home, now)
	if err != nil {
		t.Fatalf("readInternalStatus() error = %v", err)
	}
	if got.State != engine.StatusStateSynced {
		t.Fatalf("state = %q, want %q", got.State, engine.StatusStateSynced)
	}
	if got.DeviceID != "DEV-A" {
		t.Fatalf("deviceID = %q, want DEV-A", got.DeviceID)
	}
}

func TestReadInternalStatusStale(t *testing.T) {
	home := t.TempDir()
	now := time.Date(2026, time.February, 26, 18, 0, 0, 0, time.UTC)

	snapshot := engine.StatusSnapshot{
		DeviceID:       "DEV-A",
		State:          engine.StatusStateSynced,
		NeedTotalItems: 0,
		UpdatedAt:      now.Add(-(internalStatusMaxAge + time.Second)),
	}
	writeInternalStatusFile(t, home, snapshot)

	_, err := readInternalStatus(home, now)
	if !errors.Is(err, errInternalStatusStale) {
		t.Fatalf("readInternalStatus() error = %v, want %v", err, errInternalStatusStale)
	}
}

func TestReadInternalStatusMalformed(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, engine.StatusFileName)

	if err := os.WriteFile(path, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := readInternalStatus(home, time.Now())
	if err == nil {
		t.Fatal("readInternalStatus() error = nil, want parse error")
	}
}

func writeInternalStatusFile(t *testing.T, home string, snapshot engine.StatusSnapshot) {
	t.Helper()

	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	path := filepath.Join(home, engine.StatusFileName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
