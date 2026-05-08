package handler

import (
	"net/http"
	"os"
	"sync"

	"time"

	"cash-flow/internal/config"
	"cash-flow/internal/handler"
	"cash-flow/internal/middleware"
	"cash-flow/internal/repository"
	"cash-flow/internal/service"

	cors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	app  *gin.Engine
	once sync.Once
)

// once.Do memastikan inisialisasi hanya terjadi sekali
// meskipun ada multiple request — mengurangi dampak cold start
func setupApp() {
	once.Do(func() {
		godotenv.Load()

		db := config.NewDB()

		userRepo := repository.NewUserRepository(db)
		txRepo := repository.NewTransactionRepository(db)

		authSvc := service.NewAuthService(userRepo)
		txSvc := service.NewTransactionService(txRepo, userRepo, db)

		authHandler := handler.NewAuthHandler(authSvc)
		txHandler := handler.NewTransactionHandler(txSvc)

		gin.SetMode(gin.ReleaseMode)
		app = gin.New()
		app.Use(gin.Recovery())

		app.Use(cors.New(cors.Config{
			AllowOrigins:     []string{os.Getenv("FRONTEND_URL")},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))

		api := app.Group("/api/v1")
		{
			auth := api.Group("/auth")
			{
				auth.POST("/register", authHandler.Register)
				auth.POST("/login", authHandler.Login)
			}

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
	})
}

// Handler adalah entry point yang dipanggil Vercel setiap request
func Handler(w http.ResponseWriter, r *http.Request) {
	setupApp()
	app.ServeHTTP(w, r)
}
