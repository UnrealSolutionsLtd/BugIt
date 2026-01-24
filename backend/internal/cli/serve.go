// Package cli provides CLI commands for BugIt.
package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/unrealsolutions/bugit/internal/api"
	"github.com/unrealsolutions/bugit/internal/db"
	"github.com/unrealsolutions/bugit/internal/storage"
)

// ServeCmd returns the serve command.
func ServeCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the BugIt HTTP server",
		Long:  "Starts the BugIt HTTP API server for receiving repro bundle uploads.",
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir, _ := cmd.Flags().GetString("data-dir")
			logLevel, _ := cmd.Flags().GetString("log-level")

			// Setup logging
			setupLogging(logLevel)

			// Initialize storage
			store, err := storage.New(dataDir)
			if err != nil {
				return fmt.Errorf("init storage: %w", err)
			}

			// Cleanup old temp directories on startup
			if removed, err := store.CleanupOldTempDirs(time.Hour); err == nil && removed > 0 {
				slog.Info("cleaned up old temp directories", "count", removed)
			}

			// Initialize database
			database, err := db.Open(store.DBPath())
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer database.Close()

			// Create server
			version := cmd.Root().Version
			server := api.NewServer(database, store, version)

			// Setup HTTP server
			httpServer := &http.Server{
				Addr:         fmt.Sprintf(":%d", port),
				Handler:      server.Handler(),
				ReadTimeout:  30 * time.Minute, // Long timeout for large uploads
				WriteTimeout: 30 * time.Minute,
				IdleTimeout:  60 * time.Second,
			}

			// Graceful shutdown
			done := make(chan os.Signal, 1)
			signal.Notify(done, os.Interrupt, syscall.SIGTERM)

			go func() {
				slog.Info("starting server", "port", port, "data_dir", dataDir)
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					slog.Error("server error", "error", err)
					os.Exit(1)
				}
			}()

			<-done
			slog.Info("shutting down...")

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := httpServer.Shutdown(ctx); err != nil {
				return fmt.Errorf("shutdown: %w", err)
			}

			slog.Info("server stopped")
			return nil
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "HTTP port")

	return cmd
}

func setupLogging(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(handler))
}
