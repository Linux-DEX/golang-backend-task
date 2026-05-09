package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sarabjeet/golang-backend-task/internal/api"
	"github.com/sarabjeet/golang-backend-task/internal/config"
	"github.com/sarabjeet/golang-backend-task/internal/queue"
	"github.com/sarabjeet/golang-backend-task/internal/storage"
)

func main() {
	log.Println("Starting EDI Processing API Server")

	// Load configuration
	cfg := config.Load()
	log.Printf("Configuration loaded - Port: %s, MongoDB: %s\n", cfg.Server.Port, cfg.MongoDB.URI)

	// Initialize MongoDB
	db, err := storage.NewMongoDB(&cfg.MongoDB)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Close(ctx); err != nil {
			log.Printf("Failed to close MongoDB: %v", err)
		}
	}()

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := db.CreateIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to create MongoDB indexes: %v", err)
	}
	cancel()

	// Initialize Redis queue
	redisQueue, err := queue.NewRedisQueue(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis queue: %v", err)
	}
	defer func() {
		if err := redisQueue.Close(); err != nil {
			log.Printf("Failed to close Redis: %v", err)
		}
	}()

	// Setup router
	router := api.SetupRouter(db, redisQueue)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Configure HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server
	go func() {
		log.Printf("Server started on port %s\n", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
