package engine

import (
	"errors"
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
	IsConnectedTo(deviceID protocol.DeviceID) bool
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
	snap, err := r.internals.DBSnapshot(folderID)
	if err != nil {
		return 0, err
	}
	defer snap.Release()

	return snap.NeedSize(protocol.LocalDeviceID).TotalItems(), nil
}

func (r *syncthingStatusRuntime) IsConnectedTo(deviceID protocol.DeviceID) bool {
	if r.internals == nil {
		return false
	}
	return r.internals.IsConnectedTo(deviceID)
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

	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}
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

	snapshot.ConnectedPeers, snapshot.TotalPeers = peerConnectionCounts(cfg.Devices, s.localID, s.runtime)
	snapshot.State = deriveStatusState(folderState, needTotalItems, snapshot.ConnectedPeers, snapshot.TotalPeers)

	return snapshot
}

func peerConnectionCounts(devices []config.DeviceConfiguration, localID protocol.DeviceID, runtime statusRuntime) (connected int, total int) {
	for _, device := range devices {
		if device.DeviceID == localID {
			continue
		}
		total++
		if runtime.IsConnectedTo(device.DeviceID) {
			connected++
		}
	}
	return connected, total
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
		a.FolderID == b.FolderID &&
		a.FolderState == b.FolderState &&
		a.NeedTotalItems == b.NeedTotalItems &&
		a.ConnectedPeers == b.ConnectedPeers &&
		a.TotalPeers == b.TotalPeers &&
		a.Error == b.Error
}
