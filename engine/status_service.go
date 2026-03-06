package engine

import (
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/events"
	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/syncthing"
)

const defaultStatusPollInterval = 3 * time.Second

const statusRefreshEventMask = events.StateChanged |
	events.LocalIndexUpdated |
	events.RemoteIndexUpdated |
	events.DeviceConnected |
	events.DeviceDisconnected |
	events.FolderErrors |
	events.FolderWatchStateChanged |
	events.ConfigSaved

var errInternalsUnavailable = errors.New("syncthing internals unavailable")

type statusConfigSource interface {
	RawCopy() config.Configuration
}

type statusRuntime interface {
	FolderState(folderID string) (string, error)
	NeedTotalItems(folderID string) (int, error)
	NeedFolderFiles(folderID string, max int) ([]string, error)
	IsConnectedTo(deviceID protocol.DeviceID) bool
	DeviceLastSeen(deviceID protocol.DeviceID) time.Time
	DeviceNeedBytes(folderID string, deviceID protocol.DeviceID) int64
}

type statusEventSubscription interface {
	C() <-chan events.Event
	Unsubscribe()
}

type statusEventSource interface {
	Subscribe(mask events.EventType) statusEventSubscription
}

type statusEventLogger struct {
	logger events.Logger
}

func newStatusEventSource(logger events.Logger) statusEventSource {
	if logger == nil {
		return nil
	}
	return statusEventLogger{logger: logger}
}

func (s statusEventLogger) Subscribe(mask events.EventType) statusEventSubscription {
	return s.logger.Subscribe(mask)
}

type syncthingStatusRuntime struct {
	internals *syncthing.Internals
}

func newSyncthingStatusRuntime(internals *syncthing.Internals) statusRuntime {
	return &syncthingStatusRuntime{internals: internals}
}

func (r *syncthingStatusRuntime) FolderState(folderID string) (string, error) {
	if r.internals == nil {
		return "", errInternalsUnavailable
	}
	state, _, err := r.internals.FolderState(folderID)
	return state, err
}

func (r *syncthingStatusRuntime) NeedTotalItems(folderID string) (int, error) {
	if r.internals == nil {
		return 0, errInternalsUnavailable
	}
	counts, err := r.internals.NeedSize(folderID, protocol.LocalDeviceID)
	if err != nil {
		return 0, err
	}
	return counts.TotalItems(), nil
}

func (r *syncthingStatusRuntime) NeedFolderFiles(folderID string, max int) ([]string, error) {
	if r.internals == nil {
		return nil, errInternalsUnavailable
	}
	progress, queued, rest, err := r.internals.NeedFolderFiles(folderID, 0, max)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, list := range [3][]protocol.FileInfo{progress, queued, rest} {
		for _, fi := range list {
			paths = append(paths, fi.Name)
			if len(paths) >= max {
				return paths, nil
			}
		}
	}
	return paths, nil
}

func (r *syncthingStatusRuntime) IsConnectedTo(deviceID protocol.DeviceID) bool {
	if r.internals == nil {
		return false
	}
	return r.internals.IsConnectedTo(deviceID)
}

func (r *syncthingStatusRuntime) DeviceLastSeen(deviceID protocol.DeviceID) time.Time {
	if r.internals == nil {
		return time.Time{}
	}
	stats, err := r.internals.DeviceStatistics()
	if err != nil {
		return time.Time{}
	}
	if s, ok := stats[deviceID]; ok {
		return s.LastSeen
	}
	return time.Time{}
}

func (r *syncthingStatusRuntime) DeviceNeedBytes(folderID string, deviceID protocol.DeviceID) int64 {
	if r.internals == nil {
		return 0
	}
	counts, err := r.internals.NeedSize(folderID, deviceID)
	if err != nil {
		return 0
	}
	return counts.Bytes
}

type statusService struct {
	cfg         statusConfigSource
	runtime     statusRuntime
	eventSource statusEventSource
	localID     protocol.DeviceID

	mu        sync.RWMutex
	snapshot  StatusSnapshot
	published bool

	updates chan StatusSnapshot

	pollInterval time.Duration
	now          func() time.Time

	running     bool
	loggedStuck bool
	stopCh      chan struct{}
	doneCh      chan struct{}
}

func newStatusService(cfg statusConfigSource, runtime statusRuntime, eventSource statusEventSource, localID protocol.DeviceID) *statusService {
	return &statusService{
		cfg:          cfg,
		runtime:      runtime,
		eventSource:  eventSource,
		localID:      localID,
		updates:      make(chan StatusSnapshot, 1),
		pollInterval: defaultStatusPollInterval,
		now:          time.Now,
		snapshot: StatusSnapshot{
			State:     StatusStateStarting,
			UpdatedAt: time.Now().UTC(),
		},
	}
}

func (s *statusService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	s.running = true
	s.stopCh = stopCh
	s.doneCh = doneCh
	pollInterval := s.pollInterval
	if pollInterval <= 0 {
		pollInterval = defaultStatusPollInterval
	}
	s.mu.Unlock()

	go s.run(stopCh, doneCh, pollInterval)
}

func (s *statusService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	stopCh := s.stopCh
	doneCh := s.doneCh
	s.running = false
	s.stopCh = nil
	s.doneCh = nil
	s.mu.Unlock()

	close(stopCh)
	<-doneCh
}

func (s *statusService) Snapshot() StatusSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

func (s *statusService) Updates() <-chan StatusSnapshot {
	return s.updates
}

func (s *statusService) run(stopCh <-chan struct{}, doneCh chan<- struct{}, pollInterval time.Duration) {
	defer close(doneCh)

	var sub statusEventSubscription
	var eventCh <-chan events.Event
	if s.eventSource != nil {
		sub = s.eventSource.Subscribe(statusRefreshEventMask)
		if sub != nil {
			eventCh = sub.C()
		}
	}
	if sub != nil {
		defer sub.Unsubscribe()
	}

	s.refresh()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			s.refresh()
		case _, ok := <-eventCh:
			if !ok {
				eventCh = nil
				continue
			}
			s.refresh()
		}
	}
}

func (s *statusService) refresh() {
	next := s.computeSnapshot()

	stuck := next.FolderState == "idle" && next.NeedTotalItems > 0
	if stuck && !s.loggedStuck {
		log.Printf("sync: folder is idle but %d items can't sync", next.NeedTotalItems)
		for _, p := range next.NeedPaths {
			log.Printf("sync:   stuck: %s", p)
		}
		log.Printf("sync: to fix, delete stuck paths from your sync folder or add matching .stignore patterns")
		s.loggedStuck = true
	} else if !stuck {
		s.loggedStuck = false
	}

	stored, changed := s.setSnapshot(next)
	if changed {
		s.publish(stored)
	}
}

func (s *statusService) setSnapshot(next StatusSnapshot) (StatusSnapshot, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev := s.snapshot
	if !next.UpdatedAt.After(prev.UpdatedAt) {
		next.UpdatedAt = prev.UpdatedAt.Add(time.Nanosecond)
	}

	changed := !s.published || !snapshotEqualIgnoringUpdatedAt(prev, next)
	s.snapshot = next
	s.published = true
	return next, changed
}

func (s *statusService) publish(snapshot StatusSnapshot) {
	select {
	case s.updates <- snapshot:
		return
	default:
	}

	select {
	case <-s.updates:
	default:
	}

	select {
	case s.updates <- snapshot:
	default:
	}
}

func (s *statusService) computeSnapshot() StatusSnapshot {
	snapshot := StatusSnapshot{
		State:     StatusStateStarting,
		UpdatedAt: s.now().UTC(),
	}

	if s.cfg == nil || s.runtime == nil {
		snapshot.State = StatusStateUnavailable
		snapshot.Error = "status runtime unavailable"
		return snapshot
	}

	cfg := s.cfg.RawCopy()
	if len(cfg.Folders) == 0 {
		snapshot.State = StatusStateUnavailable
		snapshot.Error = "no folders configured"
		return snapshot
	}
	if s.localID != protocol.EmptyDeviceID {
		snapshot.DeviceID = s.localID.String()
	}

	folderID := cfg.Folders[0].ID
	snapshot.FolderID = folderID

	folderState, err := s.runtime.FolderState(folderID)
	if err != nil {
		snapshot.State = StatusStateUnavailable
		snapshot.Error = err.Error()
		return snapshot
	}
	snapshot.FolderState = folderState

	needTotalItems, err := s.runtime.NeedTotalItems(folderID)
	if err != nil {
		snapshot.State = StatusStateUnavailable
		snapshot.Error = err.Error()
		return snapshot
	}
	snapshot.NeedTotalItems = needTotalItems

	snapshot.Peers = buildPeerStatuses(cfg.Devices, s.localID, folderID, s.runtime)
	snapshot.ConnectedPeers, snapshot.TotalPeers = countPeers(snapshot.Peers)
	snapshot.State = deriveStatusState(folderState, needTotalItems, snapshot.ConnectedPeers, snapshot.TotalPeers)

	if folderState == "idle" && needTotalItems > 0 {
		if paths, err := s.runtime.NeedFolderFiles(folderID, 10); err == nil && len(paths) > 0 {
			snapshot.NeedPaths = paths
		}
	}

	return snapshot
}

func buildPeerStatuses(devices []config.DeviceConfiguration, localID protocol.DeviceID, folderID string, rt statusRuntime) []PeerStatus {
	var peers []PeerStatus
	for _, device := range devices {
		if device.DeviceID == localID {
			continue
		}
		peers = append(peers, PeerStatus{
			ShortID:   shortDeviceID(device.DeviceID),
			Name:      device.Name,
			Connected: rt.IsConnectedTo(device.DeviceID),
			NeedBytes: rt.DeviceNeedBytes(folderID, device.DeviceID),
			LastSeen:  rt.DeviceLastSeen(device.DeviceID),
		})
	}
	sort.Slice(peers, func(i, j int) bool {
		li := peers[i].Name
		if li == "" {
			li = peers[i].ShortID
		}
		lj := peers[j].Name
		if lj == "" {
			lj = peers[j].ShortID
		}
		return li < lj
	})
	return peers
}

func countPeers(peers []PeerStatus) (connected, total int) {
	total = len(peers)
	for _, p := range peers {
		if p.Connected {
			connected++
		}
	}
	return
}

func shortDeviceID(id protocol.DeviceID) string {
	s := id.String()
	if i := strings.Index(s, "-"); i > 0 {
		return s[:i]
	}
	return s
}

func deriveStatusState(folderState string, needTotalItems, connectedPeers, totalPeers int) StatusState {
	state := strings.ToLower(strings.TrimSpace(folderState))

	switch {
	case state == "":
		return StatusStateStarting
	case strings.Contains(state, "error"), state == "unknown":
		return StatusStateError
	case state == "idle":
		if needTotalItems > 0 {
			return StatusStateSyncing
		}
		if totalPeers > 0 && connectedPeers == 0 {
			return StatusStateWaitingPeers
		}
		return StatusStateSynced
	default:
		return StatusStateSyncing
	}
}

func snapshotEqualIgnoringUpdatedAt(a, b StatusSnapshot) bool {
	return a.State == b.State &&
		a.DeviceID == b.DeviceID &&
		a.FolderID == b.FolderID &&
		a.FolderState == b.FolderState &&
		a.NeedTotalItems == b.NeedTotalItems &&
		a.ConnectedPeers == b.ConnectedPeers &&
		a.TotalPeers == b.TotalPeers &&
		a.Error == b.Error &&
		peersEqual(a.Peers, b.Peers)
}

func peersEqual(a, b []PeerStatus) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
