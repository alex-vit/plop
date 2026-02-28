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
	label     string
	iconState icon.StatusLight
}

func startStatusMonitor(updates <-chan engine.StatusSnapshot, item *systray.MenuItem, peerItems []*systray.MenuItem, version string) func() {
	stop := make(chan struct{})
	var once sync.Once

	go func() {
		current := trayStatus{}
		apply := func(next trayStatus) {
			if next.label != current.label {
				display := "Plop " + version + next.label[len("Plop"):]
				item.SetTitle(display)
				systray.SetTooltip(display)
			}
			if next.iconState != current.iconState {
				setTrayIcon(next.iconState)
			}
			current = next
		}

		applyPeers := func(peers []engine.PeerStatus) {
			for i, slot := range peerItems {
				if i < len(peers) {
					indicator := "✗ "
					if peers[i].Connected {
						indicator = "✓ "
					}
					label := peers[i].Name
					if label == "" {
						label = peers[i].ShortID
					}
					slot.SetTitle(indicator + label)
					slot.Show()
				} else {
					slot.Hide()
				}
			}
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
					applyPeers(nil)
					continue
				}
				apply(trayStatusFromSnapshot(snapshot))
				applyPeers(snapshot.Peers)
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
	peers := func() string {
		if snapshot.TotalPeers == 0 {
			return "no peers"
		}
		return fmt.Sprintf("%d/%d peers", snapshot.ConnectedPeers, snapshot.TotalPeers)
	}

	switch snapshot.State {
	case engine.StatusStateError:
		return trayStatus{label: "Plop: error", iconState: icon.StatusLightAttention}
	case engine.StatusStateUnavailable:
		return trayStatus{label: "Plop: unavailable", iconState: icon.StatusLightAttention}
	case engine.StatusStateWaitingPeers:
		return trayStatus{label: fmt.Sprintf("Plop: waiting (%s)", peers()), iconState: icon.StatusLightOffline}
	case engine.StatusStateSynced:
		return trayStatus{label: fmt.Sprintf("Plop: synced (%s)", peers()), iconState: icon.StatusLightSynced}
	case engine.StatusStateSyncing:
		return trayStatus{label: fmt.Sprintf("Plop: syncing (%s)", peers()), iconState: icon.StatusLightSyncing}
	case engine.StatusStateStarting:
		return trayStatus{label: "Plop: starting...", iconState: icon.StatusLightSyncing}
	default:
		folderState := strings.ToLower(strings.TrimSpace(snapshot.FolderState))
		switch {
		case folderState == "":
			return trayStatus{label: "Plop: starting...", iconState: icon.StatusLightSyncing}
		case strings.Contains(folderState, "error") || folderState == "unknown":
			return trayStatus{label: "Plop: error", iconState: icon.StatusLightAttention}
		default:
			return trayStatus{label: fmt.Sprintf("Plop: syncing (%s)", peers()), iconState: icon.StatusLightSyncing}
		}
	}
}
