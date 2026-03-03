package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/alex-vit/plop/engine"
)

func runID(args []string) error {
	fs := flag.NewFlagSet("plop id", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("id takes no arguments")
	}

	certFile := filepath.Join(homeDir, "cert.pem")
	keyFile := filepath.Join(homeDir, "key.pem")

	cert, err := engine.LoadOrGenerateCert(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("loading certificate: %w", err)
	}

	fmt.Println(engine.DeviceID(cert))
	return nil
}
