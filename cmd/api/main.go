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

	cfg := config.Load()
	log.Printf("Configuration loaded - Port: %s, MongoDB: %s\n", cfg.Server.Port, cfg.MongoDB.URI)

	// init mongo
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := db.CreateIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to create MongoDB indexes: %v", err)
	}
	cancel()

	// init redis
	redisQueue, err := queue.NewRedisQueue(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis queue: %v", err)
	}
	defer func() {
		if err := redisQueue.Close(); err != nil {
			log.Printf("Failed to close Redis: %v", err)
		}
	}()

	router := api.SetupRouter(db, redisQueue)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Server started on port %s\n", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
