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

type PeerStatus struct {
	ShortID   string `json:"shortID"`
	Connected bool   `json:"connected"`
}

type StatusSnapshot struct {
	DeviceID       string       `json:"deviceID,omitempty"`
	State          StatusState  `json:"state"`
	FolderID       string       `json:"folderID,omitempty"`
	FolderState    string       `json:"folderState,omitempty"`
	NeedTotalItems int          `json:"needTotalItems"`
	ConnectedPeers int          `json:"connectedPeers"`
	TotalPeers     int          `json:"totalPeers"`
	Peers          []PeerStatus `json:"peers,omitempty"`
	Error          string       `json:"error,omitempty"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}
