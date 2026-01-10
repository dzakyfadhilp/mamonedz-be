package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mamonedz/internal/config"
	"mamonedz/internal/database"
	"mamonedz/internal/handlers"
	"mamonedz/internal/middleware"
	"mamonedz/internal/models"
	"mamonedz/internal/repository"
	"mamonedz/internal/services"
	"mamonedz/pkg/response"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&models.User{}, &models.Expense{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Setup repositories
	userRepo := repository.NewUserRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)

	// Setup services
	authService := services.NewAuthService(userRepo, cfg)
	expenseService := services.NewExpenseService(expenseRepo)

	// Setup handlers
	authHandler := handlers.NewAuthHandler(authService)
	expenseHandler := handlers.NewExpenseHandler(expenseService)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS(cfg.CORSOrigins))
	router.Use(gin.Logger())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "healthy"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.Auth(authService))
		{
			// Get current user
			protected.GET("/auth/me", authHandler.Me)

			// Expenses
			expenses := protected.Group("/expenses")
			{
				expenses.GET("", expenseHandler.GetAll)
				expenses.GET("/stats", expenseHandler.GetStats)
				expenses.GET("/:id", expenseHandler.GetByID)
				expenses.POST("", expenseHandler.Create)
				expenses.PUT("/:id", expenseHandler.Update)
				expenses.DELETE("/:id", expenseHandler.Delete)
			}
		}
	}

	// Server setup with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
