package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/syncthing/syncthing/lib/protocol"
)

// validID is a well-formed Syncthing device ID string for use in tests.
const validID1 = "NVUIHRB-CAIDSJU-NVEW4J4-GYJG5UC-MRLKTYJ-TTKD5MN-AWS7PXD-7EPRYA6"
const validID2 = "L4ASN6X-XR7BHYQ-AC5HEAE-62HV7QA-CAG2GR4-SJ5BFQ2-RP2RBBP-K5HUEAH"

func mustParseID(t *testing.T, s string) protocol.DeviceID {
	t.Helper()
	id, err := protocol.DeviceIDFromString(s)
	if err != nil {
		t.Fatalf("mustParseID(%q): %v", s, err)
	}
	return id
}

func TestParsePeersFile(t *testing.T) {
	t.Parallel()

	id1 := mustParseID(t, validID1)
	id2 := mustParseID(t, validID2)

	tests := []struct {
		name    string
		content string
		want    []PeerEntry
		wantErr bool
	}{
		{
			name:    "empty file",
			content: "",
			want:    nil,
		},
		{
			name:    "only blank lines",
			content: "\n\n\n",
			want:    nil,
		},
		{
			name:    "plain comment only",
			content: "# just a comment\n",
			want:    nil,
		},
		{
			name:    "single ID no name",
			content: validID1 + "\n",
			want:    []PeerEntry{{ID: id1}},
		},
		{
			name:    "comment before ID gives name",
			content: "# MacBook\n" + validID1 + "\n",
			want:    []PeerEntry{{ID: id1, Name: "MacBook"}},
		},
		{
			name:    "inline name after ID",
			content: validID1 + " Poco\n",
			want:    []PeerEntry{{ID: id1, Name: "Poco"}},
		},
		{
			name:    "inline name with spaces",
			content: validID1 + " Alex MacBook\n",
			want:    []PeerEntry{{ID: id1, Name: "Alex MacBook"}},
		},
		{
			name:    "inline name overrides preceding comment",
			content: "# Comment\n" + validID1 + " Inline\n",
			want:    []PeerEntry{{ID: id1, Name: "Inline"}},
		},
		{
			name: "blank line resets pending comment",
			content: "# MacBook\n" +
				"\n" +
				validID1 + "\n",
			want: []PeerEntry{{ID: id1, Name: ""}},
		},
		{
			name: "multiple consecutive comments last one wins",
			content: "# First\n" +
				"# Second\n" +
				validID1 + "\n",
			want: []PeerEntry{{ID: id1, Name: "Second"}},
		},
		{
			name:    "comment at EOF with no following ID is discarded",
			content: validID1 + "\n# orphan comment\n",
			want:    []PeerEntry{{ID: id1}},
		},
		{
			name:    "malformed ID is skipped",
			content: "not-a-device-id\n" + validID1 + "\n",
			want:    []PeerEntry{{ID: id1}},
		},
		{
			name: "malformed ID resets pending comment",
			content: "# MacBook\n" +
				"not-a-device-id\n" +
				validID1 + "\n",
			want: []PeerEntry{{ID: id1, Name: ""}},
		},
		{
			name: "two peers comment-before format",
			content: "# MacBook\n" +
				validID1 + "\n" +
				"# Poco\n" +
				validID2 + "\n",
			want: []PeerEntry{
				{ID: id1, Name: "MacBook"},
				{ID: id2, Name: "Poco"},
			},
		},
		{
			name: "two peers inline format",
			content: validID1 + " MacBook\n" +
				validID2 + " Poco\n",
			want: []PeerEntry{
				{ID: id1, Name: "MacBook"},
				{ID: id2, Name: "Poco"},
			},
		},
		{
			name: "mixed formats",
			content: "# MacBook\n" +
				validID1 + "\n" +
				validID2 + " Poco\n",
			want: []PeerEntry{
				{ID: id1, Name: "MacBook"},
				{ID: id2, Name: "Poco"},
			},
		},
		{
			name:    "whitespace-only lines treated as blank",
			content: "  \n# MacBook\n  \t  \n" + validID1 + "\n",
			want:    []PeerEntry{{ID: id1, Name: ""}},
		},
		{
			name:    "comment name is trimmed",
			content: "#   MacBook   \n" + validID1 + "\n",
			want:    []PeerEntry{{ID: id1, Name: "MacBook"}},
		},
		{
			name:    "ID line is trimmed",
			content: "  " + validID1 + "  \n",
			want:    []PeerEntry{{ID: id1}},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "peers.txt")
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatalf("setup: %v", err)
			}

			got, err := ParsePeersFile(path)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParsePeersFile() error = %v, wantErr %v", err, tc.wantErr)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("ParsePeersFile() returned %d entries, want %d\ngot:  %v\nwant: %v", len(got), len(tc.want), got, tc.want)
			}
			for i := range tc.want {
				if got[i].ID != tc.want[i].ID {
					t.Errorf("entry[%d].ID = %v, want %v", i, got[i].ID, tc.want[i].ID)
				}
				if got[i].Name != tc.want[i].Name {
					t.Errorf("entry[%d].Name = %q, want %q", i, got[i].Name, tc.want[i].Name)
				}
			}
		})
	}
}

func TestParsePeersFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := ParsePeersFile(filepath.Join(t.TempDir(), "nonexistent.txt"))
	if !os.IsNotExist(err) {
		t.Fatalf("expected not-exist error, got %v", err)
	}
}

func TestWritePeersFileRoundTrip(t *testing.T) {
	t.Parallel()

	id1 := mustParseID(t, validID1)
	id2 := mustParseID(t, validID2)

	original := []PeerEntry{
		{ID: id1, Name: "MacBook"},
		{ID: id2, Name: "Poco"},
	}

	path := filepath.Join(t.TempDir(), "peers.txt")
	if err := WritePeersFile(path, original); err != nil {
		t.Fatalf("WritePeersFile: %v", err)
	}

	got, err := ParsePeersFile(path)
	if err != nil {
		t.Fatalf("ParsePeersFile: %v", err)
	}
	if len(got) != len(original) {
		t.Fatalf("round-trip: got %d entries, want %d", len(got), len(original))
	}
	for i := range original {
		if got[i].ID != original[i].ID {
			t.Errorf("entry[%d].ID = %v, want %v", i, got[i].ID, original[i].ID)
		}
		if got[i].Name != original[i].Name {
			t.Errorf("entry[%d].Name = %q, want %q", i, got[i].Name, original[i].Name)
		}
	}
}
