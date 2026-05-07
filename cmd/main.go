package main

import (
	"log"
	"os"
	"time"

	"cash-flow/internal/config"
	"cash-flow/internal/handler"
	"cash-flow/internal/middleware"
	"cash-flow/internal/repository"
	"cash-flow/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env")
	}

	db := config.NewDB()
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authSvc)

	txRepo := repository.NewTransactionRepository(db)
	txSvc := service.NewTransactionService(txRepo, userRepo, db)
	txHandler := handler.NewTransactionHandler(txSvc)

	r := gin.Default()

	cors_origin := os.Getenv("CORS_ORIGIN_URL")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cors_origin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/transactions", txHandler.Create)
			protected.GET("/transactions", txHandler.GetAll)
			protected.GET("/transactions/:id", txHandler.GetByID)
			protected.PUT("/transactions/:id", txHandler.Update)
			protected.DELETE("/transactions/:id", txHandler.Delete)
		}
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
