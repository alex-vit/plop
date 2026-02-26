//go:build darwin

package autostart

import (
	"bytes"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	launchAgentLabel = "com.alexvit.plop"
	launchAgentName  = launchAgentLabel + ".plist"
)

func MenuLabel() string {
	return "Start on Login"
}

func IsEnabled(_ string) bool {
	path, err := launchAgentPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func Enable(homeDir string) error {
	path, err := launchAgentPath()
	if err != nil {
		return err
	}
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	plist := launchAgentPlist(exePath, cleanHomeDir(homeDir))
	return os.WriteFile(path, []byte(plist), 0o644)
}

func Disable(_ string) error {
	path, err := launchAgentPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func launchAgentPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", launchAgentName), nil
}

func launchAgentPlist(exePath, homeDir string) string {
	args := []string{exePath}
	if homeDir != "" {
		args = append(args, "--home", homeDir)
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n")
	b.WriteString(`<plist version="1.0">` + "\n")
	b.WriteString("<dict>\n")
	b.WriteString("  <key>Label</key>\n")
	b.WriteString("  <string>" + xmlEscape(launchAgentLabel) + "</string>\n")
	b.WriteString("  <key>ProgramArguments</key>\n")
	b.WriteString("  <array>\n")
	for _, arg := range args {
		b.WriteString("    <string>" + xmlEscape(arg) + "</string>\n")
	}
	b.WriteString("  </array>\n")
	b.WriteString("  <key>RunAtLoad</key>\n")
	b.WriteString("  <true/>\n")
	b.WriteString("</dict>\n")
	b.WriteString("</plist>\n")
	return b.String()
}

func xmlEscape(v string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(v))
	return b.String()
}
