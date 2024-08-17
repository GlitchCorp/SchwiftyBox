package server

import (
	"context"
	"log"
	"net/http"

	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module provides server dependency injection
var Module = fx.Module("server",
	fx.Provide(NewServer),
)

// Server represents the HTTP server
type Server struct {
	engine *gin.Engine
	config *config.ServerConfig
}

// NewServer creates a new HTTP server
func NewServer(lc fx.Lifecycle, cfg *config.Config, handlers *handlers.Handlers) *Server {
	// Set Gin mode based on environment (you can make this configurable)
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	// Add middleware
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	// Setup routes
	api := engine.Group("/api")
	{
		// Authentication endpoints (no auth required)
		api.POST("/users", handlers.RegisterUser)
		api.POST("/token", handlers.Login)
		api.POST("/refresh", handlers.RefreshToken)
		api.POST("/token/verify", handlers.VerifyToken)

		// User statistics (no auth required)
		api.GET("/users", handlers.GetUserStatistics)

		// Password reset endpoints (no auth required)
		api.POST("/users/reset-password", handlers.RequestPasswordReset)
		api.POST("/users/send-password", handlers.SetNewPassword)

		// Protected routes (require JWT authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(handlers.GetJWTService()))
		{
			// User management
			protected.POST("/users/deactivate", handlers.DeactivateUser)
			protected.GET("/users/:user_id", handlers.GetUserDetails)
			protected.PUT("/users/:user_id", handlers.UpdateUserDetails)
			protected.DELETE("/users/:user_id", handlers.DeleteUser)

			// Items management
			protected.GET("/items", handlers.GetItems)
			protected.GET("/items/:item_id", handlers.GetItem)
			protected.POST("/items", handlers.CreateItem)
			protected.PATCH("/items/:item_id", handlers.UpdateItem)
			protected.DELETE("/items/:item_id", handlers.DeleteItem)

			// Tags management
			protected.GET("/tags", handlers.GetTags)
			protected.POST("/tags", handlers.CreateTag)
		}
	}

	server := &Server{
		engine: engine,
		config: &cfg.Server,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("Starting HTTP server on %s", server.config.Port)
			go func() {
				if err := server.engine.Run(server.config.Port); err != nil && err != http.ErrServerClosed {
					log.Fatalf("Failed to start server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Stopping HTTP server")
			return nil
		},
	})

	return server
}

// GetEngine returns the Gin engine (useful for testing)
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}
