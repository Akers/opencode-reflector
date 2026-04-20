package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/akers/opencode-reflector/internal/config"
	"github.com/akers/opencode-reflector/internal/server"
	"github.com/akers/opencode-reflector/internal/store"
)

var (
	version = "dev"
)

func main() {
	configPath := flag.String("config", "reflector.yaml", "path to configuration file")
	reflectorDir := flag.String("data-dir", ".reflector", "path to .reflector data directory")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("opencode-reflector %s\n", version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure data directory layout
	if err := config.EnsureDefaultLayout(*reflectorDir); err != nil {
		log.Fatalf("Failed to create directory layout: %v", err)
	}

	// Initialize SQLite store
	dbPath := filepath.Join(*reflectorDir, "data", "reflector.db")
	ctx := context.Background()
	s, err := store.NewStore(ctx, dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer s.Close()

	// Create server (sentiment and L2 classification are read from session metadata)
	srv := server.NewServer(cfg, s, *reflectorDir, version)

	// Setup HTTP server
	addr := fmt.Sprintf(":%d", cfg.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      srv.Routes(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("opencode-reflector %s starting on %s", version, addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
