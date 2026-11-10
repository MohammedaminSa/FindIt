package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/config"
	"github.com/MohammedaminSa/FindIt/backend/internal/database"
	"github.com/MohammedaminSa/FindIt/backend/internal/handler"
	"github.com/MohammedaminSa/FindIt/backend/internal/middleware"
	"github.com/MohammedaminSa/FindIt/backend/internal/repository"
	"github.com/MohammedaminSa/FindIt/backend/internal/service"
	"github.com/MohammedaminSa/FindIt/backend/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// setupRouter initializes the Gin engine with all global middlewares and v1 API routes.
func setupRouter(db handler.Pinger, authHandler *handler.AuthHandler) *gin.Engine {
	router := gin.New()

	// Global Middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Handlers
	healthHandler := handler.NewHealthHandler(db)

	// API v1 Route Group
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)
		if authHandler != nil {
			authHandler.RegisterRoutes(v1)
		}
	}

	return router
}

func main() {
	// 1. Load configuration from environment / .env
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load application configuration: %v", err)
	}

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. Initialize PostgreSQL connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var dbPool *pgxpool.Pool
	dbPool, err = database.NewPool(ctx, cfg)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v (running in degraded mode)", err)
	} else {
		defer dbPool.Close()
		log.Println("PostgreSQL connection pool established successfully")
	}

	// 3. Setup dependencies, router, and routes
	var pinger handler.Pinger
	if dbPool != nil {
		pinger = dbPool
	}

	jwtManager, err := token.NewJWTManager(cfg.JWTSecret, 24*time.Hour)
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	userRepo := repository.NewUserRepository(dbPool)
	authService := service.NewAuthService(userRepo, jwtManager)
	authHandler := handler.NewAuthHandler(authService)

	router := setupRouter(pinger, authHandler)

	// 4. Configure HTTP server with production-ready timeouts
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Start HTTP server in a background Goroutine
	go func() {
		log.Printf("FindIt API server listening on http://localhost:%s (%s mode)", cfg.ServerPort, cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to listen: %v", err)
		}
	}()

	// 6. Graceful Shutdown listener (handles SIGINT / SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown signal received, shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("FindIt API server gracefully terminated")
}
