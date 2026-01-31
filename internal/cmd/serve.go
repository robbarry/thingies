package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"thingies/internal/cmd/shared"
	"thingies/internal/db"
	"thingies/internal/server"
)

var (
	servePort int
	serveHost string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	Long:  `Start an HTTP server providing REST API access to Things 3 data.`,
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8484, "Port to listen on")
	serveCmd.Flags().StringVar(&serveHost, "host", "0.0.0.0", "Host to bind to")

	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// Open database
	thingsDB, err := db.Open(shared.GetDBPath(cmd))
	if err != nil {
		return err
	}
	defer thingsDB.Close()

	// Create server
	cfg := server.Config{
		Host: serveHost,
		Port: servePort,
	}
	srv := server.New(cfg, thingsDB)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Wait for signal or error
	select {
	case err := <-errChan:
		return err
	case <-sigChan:
		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return srv.Shutdown(ctx)
	}
}
