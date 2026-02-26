package engine

import "time"

type StatusState string

const (
	StatusStateStarting     StatusState = "starting"
	StatusStateSyncing      StatusState = "syncing"
	StatusStateSynced       StatusState = "synced"
	StatusStateWaitingPeers StatusState = "waiting_peers"
	StatusStateError        StatusState = "error"
	StatusStateUnavailable  StatusState = "unavailable"
)

type StatusSnapshot struct {
	State          StatusState `json:"state"`
	FolderID       string      `json:"folderID,omitempty"`
	FolderState    string      `json:"folderState,omitempty"`
	NeedTotalItems int         `json:"needTotalItems"`
	ConnectedPeers int         `json:"connectedPeers"`
	TotalPeers     int         `json:"totalPeers"`
	Error          string      `json:"error,omitempty"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}
