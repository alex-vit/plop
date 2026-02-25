package engine

import (
	"bufio"
	"os"
	"strings"

	"github.com/syncthing/syncthing/lib/protocol"
)

// ParsePeersFile reads a peers.txt file and returns the device IDs found.
// Blank lines and lines starting with # are ignored. Whitespace is trimmed.
func ParsePeersFile(path string) ([]protocol.DeviceID, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var peers []protocol.DeviceID
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		id, err := protocol.DeviceIDFromString(line)
		if err != nil {
			continue // skip malformed lines
		}
		peers = append(peers, id)
	}
	return peers, scanner.Err()
}

// WritePeersFile writes device IDs to a peers.txt file, one per line.
// Existing content is replaced.
func WritePeersFile(path string, peers []protocol.DeviceID) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, p := range peers {
		if _, err := w.WriteString(p.String() + "\n"); err != nil {
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
