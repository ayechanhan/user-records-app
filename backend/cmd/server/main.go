package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/ayechanhan/user-records-app/backend/internal/config"
	"github.com/ayechanhan/user-records-app/backend/internal/handler"
	"github.com/ayechanhan/user-records-app/backend/internal/logging"
	"github.com/ayechanhan/user-records-app/backend/internal/repository/mongo"
	"github.com/ayechanhan/user-records-app/backend/internal/repository/postgres"
	"github.com/ayechanhan/user-records-app/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := postgres.Migrate(cfg.PostgresDSN); err != nil {
		log.Fatalf("postgres migrate: %v", err)
	}
	db, err := postgres.Connect(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres connect: %v", err)
	}
	userRepo := postgres.NewUserRepository(db)

	ctx := context.Background()
	mongoClient, err := mongo.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	if err := mongo.EnsureIndexes(ctx, mongoClient, cfg.MongoDBName); err != nil {
		log.Fatalf("mongo ensure indexes: %v", err)
	}
	logRepo := mongo.NewLogRepository(mongoClient, cfg.MongoDBName)

	eventBus := logging.NewBus(logging.DefaultBufferSize)
	logWorker := logging.NewWorker(eventBus, logRepo)
	logWorker.Start()

	authService := service.NewAuthService(userRepo, eventBus, cfg.HMACSecret, cfg.JWTSecret, cfg.AdminEmail, cfg.AdminPassword)
	authHandler := handler.NewAuthHandler(authService)

	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	v1.POST("/auth/login", authHandler.Login)

	log.Printf("listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
