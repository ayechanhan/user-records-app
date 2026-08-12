package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/ayechanhan/user-records-app/backend/internal/config"
	mongorepo "github.com/ayechanhan/user-records-app/backend/internal/repository/mongo"
	"github.com/ayechanhan/user-records-app/backend/internal/repository/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := postgres.Migrate(cfg.PostgresDSN); err != nil {
		log.Fatalf("postgres migrate: %v", err)
	}
	if _, err := postgres.Connect(cfg.PostgresDSN); err != nil {
		log.Fatalf("postgres connect: %v", err)
	}

	ctx := context.Background()
	mongoClient, err := mongorepo.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	if err := mongorepo.EnsureIndexes(ctx, mongoClient, cfg.MongoDBName); err != nil {
		log.Fatalf("mongo ensure indexes: %v", err)
	}

	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
