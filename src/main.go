package main

import (
	"context"
	"log"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handlers"
	"backend/internal/jwt"
	"backend/internal/middleware"
	"backend/internal/server"
	"backend/internal/user"

	"go.uber.org/fx"
)

func main() {
	log.Println("Starting SchwiftyBox application!!...")

	app := fx.New(
		// Provide configuration
		fx.Provide(config.NewConfig),

		// Include all modules
		database.Module,
		jwt.Module,
		user.Module,
		handlers.Module,
		middleware.Module,
		server.Module,

		// Add lifecycle hooks
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					log.Println("Application started successfully")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					log.Println("Application stopped")
					return nil
				},
			})
		}),

		// Ensure server is started
		fx.Invoke(func(s *server.Server) {
			log.Println("Server dependency injected and will be started")
		}),
	)

	app.Run()
}
