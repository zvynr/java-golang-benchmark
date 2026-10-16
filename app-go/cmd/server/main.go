package main

import (
	"benchmark-go/internal/config"
	"benchmark-go/internal/controller"
	"benchmark-go/internal/database"
	"benchmark-go/internal/middleware"
	"benchmark-go/internal/repository"
	"benchmark-go/internal/service"
	"benchmark-go/internal/worker"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	bootStart := time.Now()

	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Connect Database (using pgxpool configured identical to Java HikariCP for Apple-to-Apple benchmark)
	pool, err := database.NewPostgresPool(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}
	defer pool.Close()

	// 3. Initialize Centralized Worker Pool Manager
	poolMgr := worker.NewPoolManager()

	// 4. Initialize Repositories, Services, and Controllers
	ticketRepo := repository.NewTicketRepository(pool)
	benchmarkService := service.NewBenchmarkService(ticketRepo, poolMgr)
	ticketService := service.NewTicketService(ticketRepo)

	healthController := controller.NewHealthController()
	benchmarkController := controller.NewBenchmarkController(benchmarkService)
	ticketController := controller.NewTicketController(ticketService)

	// 5. Initialize Gin Router & Middleware
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Initialize Prometheus metrics registry & middleware with pgxpool stats matching HikariCP metrics
	middleware.InitMetrics(pool)
	r.Use(middleware.PrometheusMiddleware())

	// Register Routes
	healthController.RegisterRoutes(r)

	apiGroup := r.Group("/api")
	{
		benchmarkGroup := apiGroup.Group("/benchmark")
		benchmarkController.RegisterRoutes(benchmarkGroup)

		ticketGroup := apiGroup.Group("/tickets")
		ticketController.RegisterRoutes(ticketGroup)
	}

	// 6. Start HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Started Go Benchmark Application in %v (server running on :%s)", time.Since(bootStart), cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
