package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alex-vit/plop/engine"
)

func init() {
	rootCmd.AddCommand(idCmd)
}

var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Print this device's ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		certFile := filepath.Join(homeDir, "cert.pem")
		keyFile := filepath.Join(homeDir, "key.pem")

		cert, err := engine.LoadCert(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("loading certificate: %w", err)
		}

		fmt.Println(engine.DeviceID(cert))
		return nil
	},
}
