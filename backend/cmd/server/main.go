package main

import (
	"log"
	"os"

	"this-is-pun/backend/internal/data"
	"this-is-pun/backend/internal/router"
	"this-is-pun/backend/internal/service"
)

func main() {
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	seed := data.Seed()
	var repo service.Repository
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		postgresRepo, err := data.NewPostgresRepository(databaseURL, seed)
		if err != nil {
			log.Printf("database unavailable, falling back to in-memory seed data: %v", err)
			repo = data.NewMemoryRepository(seed)
		} else {
			defer postgresRepo.Close()
			repo = postgresRepo
			log.Printf("database connected")
		}
	} else {
		repo = data.NewMemoryRepository(seed)
		log.Printf("DATABASE_URL not set, using in-memory seed data")
	}
	gameService := service.NewGameService(repo)
	engine := router.New(gameService)

	log.Printf("this-is-pun backend listening on :%s", port)
	if err := engine.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
