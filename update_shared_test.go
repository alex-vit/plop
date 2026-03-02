package main

import "testing"

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"1.2.0", "1.1.0", true},
		{"1.1.1", "1.1.0", true},
		{"2.0.0", "1.9.9", true},
		{"1.10.0", "1.9.0", true},
		{"1.1.0", "1.1.0", false},
		{"1.0.0", "1.1.0", false},
		{"1.1.0", "2.0.0", false},
		{"1.1.0", "", false},
		{"1.1.0", "dev", false},
		{"bad", "1.0.0", false},
		{"1.0.0", "bad", false},
		{"1.0", "1.0.0", false},
	}
	for _, tt := range tests {
		got := isNewer(tt.latest, tt.current)
		if got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}
