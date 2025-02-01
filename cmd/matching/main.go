package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"orderMatching/internal/application/service"
	"orderMatching/internal/infrastructure/messaging"
	"orderMatching/internal/infrastructure/orderbook"
	"orderMatching/internal/infrastructure/repository"
	"orderMatching/pkg/common/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("starting service with config: %+v", cfg)

	orderRepo := repository.NewOrderRepository()
	messageBroker, err := messaging.NewOrderHandler(cfg.Kafka)
	if err != nil {
		log.Fatalf("Failed to create message broker: %v", err)
	}

	orderBook := orderbook.NewOrderBook(orderRepo, messageBroker)
	matchingService := service.NewMatchingService(orderBook, messageBroker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := matchingService.Start(ctx); err != nil {
			log.Printf("Matching service error: %v", err)
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case <-ctx.Done():
		log.Printf("Context cancelled")
	}

	log.Println("Shutting down...")
	if err := matchingService.Stop(); err != nil {
		log.Printf("Error stopping matching service: %v", err)
	}

}
