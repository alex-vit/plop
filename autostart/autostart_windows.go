//go:build windows

package autostart

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	registryKey  = `Software\Microsoft\Windows\CurrentVersion\Run`
	registryName = "Plop"
)

func MenuLabel() string {
	return "Start with Windows"
}

func IsEnabled(_ string) bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(registryName)
	return err == nil
}

func Enable(homeDir string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	k, _, err := registry.CreateKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue(registryName, runKeyValue(exePath, cleanHomeDir(homeDir)))
}

func Disable(_ string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return err
	}
	defer k.Close()
	err = k.DeleteValue(registryName)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	return err
}

func runKeyValue(exePath, homeDir string) string {
	value := quoteRunArg(exePath)
	if homeDir != "" {
		value += " --home " + quoteRunArg(homeDir)
	}
	return value
}

func quoteRunArg(s string) string {
	if s == "" {
		return `""`
	}
	if !strings.ContainsAny(s, " \t\"") {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
