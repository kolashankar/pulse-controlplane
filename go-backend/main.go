package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pulse-control-plane/config"
	"pulse-control-plane/database"
	"pulse-control-plane/routes"
	"pulse-control-plane/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	utils.InitLogger(cfg.LogLevel)

	log.Info().Msg("🚀 Starting Pulse Control Plane...")

	// Connect to MongoDB
	if err := database.ConnectMongoDB(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer database.DisconnectMongoDB()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes
	routes.SetupRoutes(router, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("port", cfg.Port).Msg("🌟 Pulse Control Plane is running")
		log.Info().Str("environment", cfg.Environment).Msg("Environment")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("🛑 Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("✅ Server exited gracefully")
}
