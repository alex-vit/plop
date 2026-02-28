package engine

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/syncthing/syncthing/lib/protocol"
)

// PeerEntry is a device ID with an optional display name from peers.txt.
type PeerEntry struct {
	ID   protocol.DeviceID
	Name string
}

// ParsePeersFile reads a peers.txt file and returns the peer entries found.
// Blank lines and malformed lines are ignored.
//
// Two name formats are supported:
//   - Comment before ID:  "# MacBook\nDEVICE_ID"
//   - Inline after ID:    "DEVICE_ID  MacBook"
//
// A comment line directly preceding an ID line is used as its name.
// A blank line resets the pending comment (treating it as a plain comment).
// Inline name takes precedence over a preceding comment.
func ParsePeersFile(path string) ([]PeerEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []PeerEntry
	var pendingName string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			pendingName = ""
			continue
		}
		if strings.HasPrefix(line, "#") {
			pendingName = strings.TrimSpace(line[1:])
			continue
		}
		// First field is the device ID; remaining fields are an optional inline name.
		fields := strings.Fields(line)
		id, err := protocol.DeviceIDFromString(fields[0])
		if err != nil {
			pendingName = ""
			continue
		}
		name := pendingName
		if len(fields) > 1 {
			name = strings.Join(fields[1:], " ")
		}
		entries = append(entries, PeerEntry{ID: id, Name: name})
		pendingName = ""
	}
	return entries, scanner.Err()
}

// WritePeersFile writes peer entries to a peers.txt file, one per line.
// Names are written as a comment on the line before the device ID.
// Existing content is replaced.
func WritePeersFile(path string, peers []PeerEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, p := range peers {
		if p.Name != "" {
			if _, err := fmt.Fprintf(w, "# %s\n", p.Name); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "%s\n", p.ID); err != nil {
			return err
		}
	}
	return w.Flush()
}

// AppendPeersFile appends a single device ID to a peers.txt file.
// Creates the file if it doesn't exist.
func AppendPeersFile(path string, id protocol.DeviceID) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(id.String() + "\n")
	return err
}
