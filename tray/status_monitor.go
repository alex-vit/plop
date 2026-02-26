package tray

import (
	"fmt"
	"strings"
	"sync"

	"github.com/alex-vit/plop/engine"
	"github.com/alex-vit/plop/icon"
	"github.com/energye/systray"
)

type trayStatus struct {
	title     string
	tooltip   string
	iconState icon.StatusLight
}

func startStatusMonitor(updates <-chan engine.StatusSnapshot, item *systray.MenuItem) func() {
	stop := make(chan struct{})
	var once sync.Once

	go func() {
		current := trayStatus{}
		apply := func(next trayStatus) {
			if next.title != current.title {
				item.SetTitle(next.title)
			}
			if next.tooltip != current.tooltip {
				systray.SetTooltip(next.tooltip)
			}
			if next.iconState != current.iconState {
				setTrayIcon(next.iconState)
			}
			current = next
		}

		apply(trayStatusFromSnapshot(engine.StatusSnapshot{State: engine.StatusStateStarting}))

		for {
			select {
			case <-stop:
				return
			case snapshot, ok := <-updates:
				if !ok {
					updates = nil
					apply(trayStatusFromSnapshot(engine.StatusSnapshot{State: engine.StatusStateUnavailable}))
					continue
				}
				apply(trayStatusFromSnapshot(snapshot))
			}
		}
	}()

	return func() {
		once.Do(func() {
			close(stop)
		})
	}
}

func trayStatusFromSnapshot(snapshot engine.StatusSnapshot) trayStatus {
	switch snapshot.State {
	case engine.StatusStateError:
		return trayStatus{
			title:     "Status: Error",
			tooltip:   "plop - Sync error",
			iconState: icon.StatusLightAttention,
		}
	case engine.StatusStateUnavailable:
		return trayStatus{
			title:     "Status: Unavailable",
			tooltip:   "plop - Status unavailable",
			iconState: icon.StatusLightAttention,
		}
	case engine.StatusStateWaitingPeers:
		if snapshot.TotalPeers > 0 {
			return trayStatus{
				title:     "Status: Waiting for peers",
				tooltip:   fmt.Sprintf("plop - Waiting for peers (%d/%d connected)", snapshot.ConnectedPeers, snapshot.TotalPeers),
				iconState: icon.StatusLightAttention,
			}
		}
		return trayStatus{
			title:     "Status: Waiting for peers",
			tooltip:   "plop - Waiting for peers",
			iconState: icon.StatusLightAttention,
		}
	case engine.StatusStateSynced:
		if snapshot.TotalPeers > 0 {
			return trayStatus{
				title:     "Status: Synced",
				tooltip:   fmt.Sprintf("plop - Synced (%d/%d peers connected)", snapshot.ConnectedPeers, snapshot.TotalPeers),
				iconState: icon.StatusLightSynced,
			}
		}
		return trayStatus{
			title:     "Status: Synced",
			tooltip:   "plop - Synced",
			iconState: icon.StatusLightSynced,
		}
	case engine.StatusStateSyncing:
		return trayStatus{
			title:     "Status: Syncing...",
			tooltip:   "plop - Syncing...",
			iconState: icon.StatusLightSyncing,
		}
	case engine.StatusStateStarting:
		return trayStatus{
			title:     "Status: Starting...",
			tooltip:   "plop - Starting...",
			iconState: icon.StatusLightSyncing,
		}
	default:
		// Defensive fallback for unknown states.
		folderState := strings.ToLower(strings.TrimSpace(snapshot.FolderState))
		switch {
		case folderState == "":
			return trayStatus{
				title:     "Status: Starting...",
				tooltip:   "plop - Starting...",
				iconState: icon.StatusLightSyncing,
			}
		case strings.Contains(folderState, "error") || folderState == "unknown":
			return trayStatus{
				title:     "Status: Error",
				tooltip:   "plop - Sync error",
				iconState: icon.StatusLightAttention,
			}
		default:
			return trayStatus{
				title:     "Status: Syncing...",
				tooltip:   "plop - Syncing...",
				iconState: icon.StatusLightSyncing,
			}
		}
	}
}
