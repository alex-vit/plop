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
			name: "starting state",
			snapshot: engine.StatusSnapshot{
				State: engine.StatusStateStarting,
			},
			want: trayStatus{
				title:     "Status: Starting...",
				tooltip:   "plop - Starting...",
				iconState: icon.StatusLightSyncing,
			},
		},
		{
			name: "syncing state",
			snapshot: engine.StatusSnapshot{
				State: engine.StatusStateSyncing,
			},
			want: trayStatus{
				title:     "Status: Syncing...",
				tooltip:   "plop - Syncing...",
				iconState: icon.StatusLightSyncing,
			},
		},
		{
			name: "error state",
			snapshot: engine.StatusSnapshot{
				State: engine.StatusStateError,
			},
			want: trayStatus{
				title:     "Status: Error",
				tooltip:   "plop - Sync error",
				iconState: icon.StatusLightAttention,
			},
		},
		{
			name: "waiting for peers state",
			snapshot: engine.StatusSnapshot{
				State:          engine.StatusStateWaitingPeers,
				ConnectedPeers: 0,
				TotalPeers:     3,
			},
			want: trayStatus{
				title:     "Status: Waiting for peers",
				tooltip:   "plop - Waiting for peers (0/3 connected)",
				iconState: icon.StatusLightAttention,
			},
		},
		{
			name: "synced with connected peers",
			snapshot: engine.StatusSnapshot{
				State:          engine.StatusStateSynced,
				ConnectedPeers: 1,
				TotalPeers:     2,
			},
			want: trayStatus{
				title:     "Status: Synced",
				tooltip:   "plop - Synced (1/2 peers connected)",
				iconState: icon.StatusLightSynced,
			},
		},
		{
			name: "synced without peers",
			snapshot: engine.StatusSnapshot{
				State: engine.StatusStateSynced,
			},
			want: trayStatus{
				title:     "Status: Synced",
				tooltip:   "plop - Synced",
				iconState: icon.StatusLightSynced,
			},
		},
		{
			name: "unavailable state",
			snapshot: engine.StatusSnapshot{
				State: engine.StatusStateUnavailable,
			},
			want: trayStatus{
				title:     "Status: Unavailable",
				tooltip:   "plop - Status unavailable",
				iconState: icon.StatusLightAttention,
			},
		},
		{
			name: "unknown state fallback to syncing",
			snapshot: engine.StatusSnapshot{
				State:       "mystery",
				FolderState: "scanning",
			},
			want: trayStatus{
				title:     "Status: Syncing...",
				tooltip:   "plop - Syncing...",
				iconState: icon.StatusLightSyncing,
			},
		},
		{
			name: "unknown state with empty folder state falls back to starting",
			snapshot: engine.StatusSnapshot{
				State: "mystery",
			},
			want: trayStatus{
				title:     "Status: Starting...",
				tooltip:   "plop - Starting...",
				iconState: icon.StatusLightSyncing,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := trayStatusFromSnapshot(tc.snapshot)
			if got != tc.want {
				t.Fatalf("trayStatusFromSnapshot() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
