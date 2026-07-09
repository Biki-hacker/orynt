package main

import (
	"fmt"
	"log"
	"orynt/internal/api"
	"orynt/internal/config"
	"orynt/internal/repository"
	"orynt/internal/service"
)

func main() {
	log.Println("[ORYNT] Starting Smart Stadium & Tournament Operations Platform...")

	// 1. Load Configurations
	cfg := config.LoadConfig()

	// 2. Initialize Repositories (with automatic in-memory fallbacks)
	dbRepo := repository.NewFirestoreRepository(cfg.UseFirestore, cfg.FirebaseProjectID, cfg.FirebaseCredentialsFile)
	pubSubRepo := repository.NewRedisRepository(cfg.UseRedis, cfg.RedisURL)

	// 3. Initialize Services (injecting repositories)
	authSvc := service.NewAuthService(dbRepo, cfg.JWTSecret)
	tournSvc := service.NewTournamentService(dbRepo, pubSubRepo)
	stadSvc := service.NewStadiumService(dbRepo, pubSubRepo)
	opsSvc := service.NewOperationsService(dbRepo, pubSubRepo)
	aiSvc := service.NewAIService(dbRepo, pubSubRepo) // reads GEMINI_API_KEY internally

	// 4. Initialize WebSockets Hub
	wsHub := api.NewWSHub()
	go wsHub.Run()

	// 5. Connect PubSub events to WebSockets hub
	api.StartPubSubListener(pubSubRepo, wsHub)

	// 6. Setup Gin handler delivery layer
	handler := api.NewAPIHandler(authSvc, tournSvc, stadSvc, opsSvc, aiSvc)
	router := api.SetupRouter(handler, wsHub, pubSubRepo)

	// 7. Start Gin Web Server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("[ORYNT] Server running at http://localhost%s", addr)
	log.Printf("[ORYNT] Real-time WebSockets hub at ws://localhost%s/api/ws", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("[ORYNT] Server failed to start: %v", err)
	}
}
