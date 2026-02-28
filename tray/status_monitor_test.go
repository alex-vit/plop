package tray

import (
	"testing"

	"github.com/alex-vit/plop/engine"
	"github.com/alex-vit/plop/icon"
)

func TestTrayStatusFromSnapshot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		snapshot engine.StatusSnapshot
		want     trayStatus
	}{
		{
			name:     "starting state",
			snapshot: engine.StatusSnapshot{State: engine.StatusStateStarting},
			want:     trayStatus{label: "Plop: starting...", iconState: icon.StatusLightSyncing},
		},
		{
			name:     "syncing state no peers",
			snapshot: engine.StatusSnapshot{State: engine.StatusStateSyncing},
			want:     trayStatus{label: "Plop: syncing (no peers)", iconState: icon.StatusLightSyncing},
		},
		{
			name:     "syncing state with peers",
			snapshot: engine.StatusSnapshot{State: engine.StatusStateSyncing, ConnectedPeers: 1, TotalPeers: 2},
			want:     trayStatus{label: "Plop: syncing (1/2 peers)", iconState: icon.StatusLightSyncing},
		},
		{
			name:     "error state",
			snapshot: engine.StatusSnapshot{State: engine.StatusStateError},
			want:     trayStatus{label: "Plop: error", iconState: icon.StatusLightAttention},
		},
		{
			name:     "waiting for peers",
			snapshot: engine.StatusSnapshot{State: engine.StatusStateWaitingPeers, ConnectedPeers: 0, TotalPeers: 3},
			want:     trayStatus{label: "Plop: waiting (0/3 peers)", iconState: icon.StatusLightOffline},
		},
		{
			name:     "synced with peers",
			snapshot: engine.StatusSnapshot{State: engine.StatusStateSynced, ConnectedPeers: 1, TotalPeers: 2},
			want:     trayStatus{label: "Plop: synced (1/2 peers)", iconState: icon.StatusLightSynced},
		},
		{
			name:     "synced no peers",
			snapshot: engine.StatusSnapshot{State: engine.StatusStateSynced},
			want:     trayStatus{label: "Plop: synced (no peers)", iconState: icon.StatusLightSynced},
		},
		{
			name:     "unavailable state",
			snapshot: engine.StatusSnapshot{State: engine.StatusStateUnavailable},
			want:     trayStatus{label: "Plop: unavailable", iconState: icon.StatusLightAttention},
		},
		{
			name:     "unknown state with active folder falls back to syncing",
			snapshot: engine.StatusSnapshot{State: "mystery", FolderState: "scanning"},
			want:     trayStatus{label: "Plop: syncing (no peers)", iconState: icon.StatusLightSyncing},
		},
		{
			name:     "unknown state with empty folder falls back to starting",
			snapshot: engine.StatusSnapshot{State: "mystery"},
			want:     trayStatus{label: "Plop: starting...", iconState: icon.StatusLightSyncing},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := trayStatusFromSnapshot(tc.snapshot)
			if got != tc.want {
				t.Fatalf("trayStatusFromSnapshot() =\n  %+v\nwant\n  %+v", got, tc.want)
			}
		})
	}
}
