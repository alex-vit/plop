package engine

import (
	"sync"
	"testing"
	"time"

	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/events"
	"github.com/syncthing/syncthing/lib/protocol"
)

func TestDeriveStatusStateMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		folderState   string
		needTotal     int
		connected     int
		total         int
		expectedState StatusState
	}{
		{
			name:          "idle and no need is synced",
			folderState:   "idle",
			needTotal:     0,
			connected:     1,
			total:         1,
			expectedState: StatusStateSynced,
		},
		{
			name:          "idle and pending need is syncing",
			folderState:   "idle",
			needTotal:     2,
			connected:     1,
			total:         1,
			expectedState: StatusStateSyncing,
		},
		{
			name:          "idle and no peers connected waits for peers",
			folderState:   "idle",
			needTotal:     0,
			connected:     0,
			total:         2,
			expectedState: StatusStateWaitingPeers,
		},
		{
			name:          "folder error state maps to error",
			folderState:   "error",
			needTotal:     0,
			connected:     1,
			total:         1,
			expectedState: StatusStateError,
		},
		{
			name:          "empty state is starting",
			folderState:   "",
			needTotal:     0,
			connected:     0,
			total:         0,
			expectedState: StatusStateStarting,
		},
		{
			name:          "non-idle active state is syncing",
			folderState:   "scanning",
			needTotal:     0,
			connected:     1,
			total:         1,
			expectedState: StatusStateSyncing,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := deriveStatusState(tc.folderState, tc.needTotal, tc.connected, tc.total)
			if got != tc.expectedState {
				t.Fatalf("deriveStatusState() = %q, want %q", got, tc.expectedState)
			}
		})
	}
}

func TestStatusServicePublishesUpdatesWithMonotonicTimestamps(t *testing.T) {
	localID := protocol.DeviceID{1}
	peerID := protocol.DeviceID{2}

	cfgSource := &fakeStatusConfigSource{
		cfg: config.Configuration{
			Folders: []config.FolderConfiguration{
				{ID: "default"},
			},
			Devices: []config.DeviceConfiguration{
				{DeviceID: localID},
				{DeviceID: peerID},
			},
		},
	}
	runtime := &fakeStatusRuntime{
		folderState:    "idle",
		needTotalItems: 0,
		connected: map[protocol.DeviceID]bool{
			peerID: true,
		},
	}
	eventSub := &fakeStatusEventSubscription{events: make(chan events.Event, 4)}
	eventSource := &fakeStatusEventSource{sub: eventSub}

	svc := newStatusService(cfgSource, runtime, eventSource, localID)
	svc.pollInterval = time.Hour
	baseTime := time.Date(2026, time.February, 26, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time {
		// Constant timestamp to validate monotonic adjustment logic.
		return baseTime
	}

	svc.Start()
	t.Cleanup(svc.Stop)

	first := waitForSnapshot(t, svc.Updates())
	if first.State != StatusStateSynced {
		t.Fatalf("first snapshot state = %q, want %q", first.State, StatusStateSynced)
	}

	runtime.setNeedTotalItems(3)
	eventSub.emit()

	second := waitForSnapshot(t, svc.Updates())
	if second.State != StatusStateSyncing {
		t.Fatalf("second snapshot state = %q, want %q", second.State, StatusStateSyncing)
	}
	if !second.UpdatedAt.After(first.UpdatedAt) {
		t.Fatalf("second UpdatedAt (%s) is not after first (%s)", second.UpdatedAt, first.UpdatedAt)
	}
}

func waitForSnapshot(t *testing.T, ch <-chan StatusSnapshot) StatusSnapshot {
	t.Helper()

	select {
	case snapshot := <-ch:
		return snapshot
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for status snapshot")
		return StatusSnapshot{}
	}
}

type fakeStatusConfigSource struct {
	mu  sync.RWMutex
	cfg config.Configuration
}

func (f *fakeStatusConfigSource) RawCopy() config.Configuration {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.cfg.Copy()
}

type fakeStatusRuntime struct {
	mu             sync.RWMutex
	folderState    string
	folderErr      error
	needTotalItems int
	needErr        error
	connected      map[protocol.DeviceID]bool
}

func (f *fakeStatusRuntime) FolderState(_ string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.folderState, f.folderErr
}

func (f *fakeStatusRuntime) NeedTotalItems(_ string) (int, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.needTotalItems, f.needErr
}

func (f *fakeStatusRuntime) IsConnectedTo(deviceID protocol.DeviceID) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.connected[deviceID]
}

func (f *fakeStatusRuntime) setNeedTotalItems(need int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.needTotalItems = need
}

type fakeStatusEventSource struct {
	sub *fakeStatusEventSubscription
}

func (f *fakeStatusEventSource) Subscribe(_ events.EventType) statusEventSubscription {
	return f.sub
}

type fakeStatusEventSubscription struct {
	events chan events.Event
	once   sync.Once
}

func (f *fakeStatusEventSubscription) C() <-chan events.Event {
	return f.events
}

func (f *fakeStatusEventSubscription) Unsubscribe() {
	f.once.Do(func() {
		close(f.events)
	})
}

func (f *fakeStatusEventSubscription) emit() {
	select {
	case f.events <- events.Event{}:
	default:
	}
}
