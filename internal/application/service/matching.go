package service

import (
	"context"
	"fmt"
	"log"
	"orderMatching/internal/domain/entities"
	"orderMatching/internal/domain/ports"
	"orderMatching/internal/infrastructure/orderbook"
	"sync"
	"time"
)

type MatchingService struct {
	orderBook     *orderbook.OrderBook
	messageBroker ports.MessageBroker
	isRunning     bool
	mutex         sync.RWMutex
	stopChan      chan struct{}
	errorChan     chan error
	startTime     time.Time
}

func NewMatchingService(orderBook *orderbook.OrderBook, broker ports.MessageBroker) *MatchingService {
	return &MatchingService{
		orderBook:     orderBook,
		messageBroker: broker,
		stopChan:      make(chan struct{}),
		errorChan:     make(chan error, 1),
	}
}

func (s *MatchingService) Start(ctx context.Context) error {
	s.mutex.Lock()
	if s.isRunning {
		s.mutex.Unlock()
		return fmt.Errorf("matching service is already running")
	}

	if err := s.messageBroker.Start(ctx); err != nil {
		s.mutex.Unlock()
		return fmt.Errorf("failed to start message broker: %w", err)
	}

	s.startTime = time.Now()
	s.isRunning = true
	log.Printf("Matching service started")
	s.mutex.Unlock()

	go s.consumeOrders(ctx)

	select {
	case err := <-s.errorChan:
		log.Printf("Service error: %v", err)
		return fmt.Errorf("matching service error: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *MatchingService) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isRunning {
		return nil
	}

	log.Printf("Stopping matching service")
	close(s.stopChan)
	s.isRunning = false

	if err := s.messageBroker.Stop(); err != nil {
		return fmt.Errorf("failed to stop message broker: %w", err)
	}

	return nil
}

func (s *MatchingService) consumeOrders(ctx context.Context) {
	orderChan, err := s.messageBroker.ConsumeOrders(ctx)
	if err != nil {
		s.errorChan <- fmt.Errorf("failed to start consuming orders: %w", err)
		return
	}

	for {
		select {
		case order := <-orderChan:
			if err := s.processOrder(ctx, order); err != nil {
				log.Printf("Error processing order %s: %v", order.ID, err)
			}

		case <-s.stopChan:
			return

		case <-ctx.Done():
			return
		}
	}
}

func (s *MatchingService) processOrder(ctx context.Context, order *entities.Order) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil")
	}

	if err := s.validateOrder(order); err != nil {
		return fmt.Errorf("invalid order: %w", err)
	}

	if err := s.orderBook.ProcessOrder(ctx, order); err != nil {
		return fmt.Errorf("order book processing failed: %w", err)
	}

	return nil
}

func (s *MatchingService) validateOrder(order *entities.Order) error {
	validationErrors := make([]string, 0)

	if order.ID == "" {
		validationErrors = append(validationErrors, "order ID is required")
	}
	if order.Quantity <= 0 {
		validationErrors = append(validationErrors,
			fmt.Sprintf("invalid quantity: %f, must be positive", order.Quantity))
	}
	if order.Price <= 0 {
		validationErrors = append(validationErrors,
			fmt.Sprintf("invalid price: %f, must be positive", order.Price))
	}
	if order.Side != "BUY" && order.Side != "SELL" {
		validationErrors = append(validationErrors,
			fmt.Sprintf("invalid side: %s, must be BUY or SELL", order.Side))
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %v", validationErrors)
	}

	return nil
}
