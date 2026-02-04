package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/config"
	"github.com/DevAnuragT/context_keeper/internal/database"
	"github.com/DevAnuragT/context_keeper/internal/logger"
	"github.com/DevAnuragT/context_keeper/internal/server"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize structured logging
	logger.Init(cfg.LogLevel)

	// Log startup information
	logger.Info("Starting ContextKeeper server", map[string]interface{}{
		"environment": cfg.Environment,
		"port":        cfg.Port,
		"log_level":   cfg.LogLevel,
	})

	// Initialize database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	logger.Info("Database connection established")

	// Run migrations
	if err := database.Migrate(db); err != nil {
		logger.Error("Failed to run migrations", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	logger.Info("Database migrations completed")

	// Initialize server
	srv := server.New(db, cfg)

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: srv,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Server starting", map[string]interface{}{
			"port": cfg.Port,
		})
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	logger.Info("Server exited")
}
