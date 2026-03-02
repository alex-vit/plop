package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/syncthing/syncthing/lib/protocol"

	"github.com/alex-vit/plop/engine"
	"github.com/alex-vit/plop/paths"
)

// stringSlice implements flag.Value for a repeatable string flag.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func runRun(args []string) error {
	var peerStrs stringSlice

	fs := flag.NewFlagSet("plop run", flag.ContinueOnError)
	fs.Var(&peerStrs, "peer", "device ID of a peer to sync with (repeatable)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("run accepts at most 1 argument (folder path)")
	}

	home := homeDir
	folderPath := ""

	if fs.NArg() == 1 {
		abs, err := filepath.Abs(fs.Arg(0))
		if err != nil {
			return fmt.Errorf("resolving folder path: %w", err)
		}
		folderPath = abs

		configDir, err := paths.ConfigDir()
		if err != nil {
			return fmt.Errorf("config dir: %w", err)
		}
		hash := sha256.Sum256([]byte(abs))
		home = filepath.Join(configDir, "instances", hex.EncodeToString(hash[:4]))
	}

	var peers []protocol.DeviceID
	for _, s := range peerStrs {
		id, err := protocol.DeviceIDFromString(s)
		if err != nil {
			return fmt.Errorf("invalid peer device ID %q: %w", s, err)
		}
		peers = append(peers, id)
	}

	eng, err := engine.New(home, folderPath, peers)
	if err != nil {
		return fmt.Errorf("creating engine: %w", err)
	}

	if err := eng.Start(); err != nil {
		return fmt.Errorf("starting engine: %w", err)
	}

	fmt.Printf("plop running as %s\n", eng.DeviceID())
	fmt.Printf("Syncing: %s\n", eng.SyncFolder())
	fmt.Printf("Config: %s\n", filepath.Join(home, "config.xml"))
	fmt.Println("Press Ctrl-C to stop.")

	// Handle shutdown signals.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		eng.Wait()
		close(done)
	}()

	select {
	case <-sig:
		fmt.Println("\nShutting down...")
		eng.Stop()
		<-done
	case <-done:
	}

	return nil
}
