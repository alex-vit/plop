package tray

import (
	"fmt"
	"strings"
	"sync"
	"time"

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
				item.SetTitle("Plop " + version + next.label[len("Plop"):])
				systray.SetTooltip(next.label)
			}
			if next.iconState != current.iconState {
				setTrayIcon(next.iconState)
			}
			current = next
		}

		applyPeers := func(peers []engine.PeerStatus, state engine.StatusState) {
			now := time.Now()
			for i, slot := range peerItems {
				if i < len(peers) {
					peer := peers[i]
					label := peer.Name
					if label == "" {
						label = peer.ShortID
					}
					var status string
					if peer.Connected {
						status = peerConnectedLabel(state)
					} else {
						status = peerLastSeenLabel(peer.LastSeen, now)
					}
					slot.SetTitle(label + " - " + status)
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
					applyPeers(nil, engine.StatusStateUnavailable)
					continue
				}
				apply(trayStatusFromSnapshot(snapshot))
				applyPeers(snapshot.Peers, snapshot.State)
			}
		}
	}()

	return func() {
		once.Do(func() {
			close(stop)
		})
	}
}

func peerConnectedLabel(state engine.StatusState) string {
	switch state {
	case engine.StatusStateSynced:
		return "synced"
	case engine.StatusStateSyncing:
		return "syncing"
	case engine.StatusStateError:
		return "error"
	default:
		return "online"
	}
}

var epochTime = time.Unix(0, 0)

func peerLastSeenLabel(t time.Time, now time.Time) string {
	if t.IsZero() || !t.After(epochTime) {
		return "offline"
	}
	ty, tm, td := t.Date()
	ny, nm, nd := now.Date()
	if ty == ny && tm == nm && td == nd {
		return "last seen " + t.Format("3:04 PM")
	}
	if ty == ny {
		return "last seen " + t.Format("Jan 2")
	}
	return "last seen " + t.Format("Jan 2, 2006")
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
