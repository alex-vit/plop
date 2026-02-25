package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/alex-vit/plop/engine"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the sync daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := engine.New(homeDir)
		if err != nil {
			return fmt.Errorf("creating engine: %w", err)
		}

		if err := eng.Start(); err != nil {
			return fmt.Errorf("starting engine: %w", err)
		}

		fmt.Printf("plop running as %s\n", eng.DeviceID())
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
	},
}
