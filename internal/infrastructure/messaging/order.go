package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mehdivijeh/matching-engine/internal/domain/entities"
	"github.com/mehdivijeh/matching-engine/internal/domain/ports"
	"github.com/mehdivijeh/matching-engine/pkg/common/config"
	"github.com/mehdivijeh/matching-engine/pkg/messaging/kafka"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

type OrderHandler struct {
	broker    *kafka.Broker
	orderChan chan *entities.Order
	config    *config.KafkaConfig
	isRunning bool
	mutex     sync.RWMutex
}

type kafkaOrderHandler struct {
	orderChan chan *entities.Order
	ctx       context.Context
}

var _ ports.MessageBroker = (*OrderHandler)(nil)

func NewOrderHandler(cfg *config.KafkaConfig) (*OrderHandler, error) {
	broker, err := kafka.NewBroker(&kafka.Config{
		Brokers: cfg.Brokers,
		GroupID: cfg.GroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create broker: %w", err)
	}

	return &OrderHandler{
		broker:    broker,
		orderChan: make(chan *entities.Order),
		config:    cfg,
	}, nil
}

func (h *OrderHandler) Start(ctx context.Context) error {
	h.mutex.Lock()
	if h.isRunning {
		h.mutex.Unlock()
		return fmt.Errorf("broker is already running")
	}
	h.isRunning = true
	h.mutex.Unlock()

	handler := &kafkaOrderHandler{
		orderChan: h.orderChan,
		ctx:       ctx,
	}

	go func() {
		if err := h.broker.Subscribe(ctx, h.config.OrdersTopic, handler); err != nil {
			log.Printf("[OrderHandler] subscription ended: %v", err)
		}
	}()

	return nil
}

func (h *OrderHandler) Stop() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.isRunning {
		return nil
	}

	h.isRunning = false
	return h.broker.Close()
}

func (h *OrderHandler) PublishMatch(ctx context.Context, match *entities.MatchEvent) error {
	data, err := json.Marshal(match)
	if err != nil {
		return fmt.Errorf("match marshal failed: %w", err)
	}

	msg := &kafka.Message{
		Topic: h.config.MatchesTopic,
		Key:   match.ID,
		Value: data,
	}

	if err := h.broker.Publish(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish match: %w", err)
	}

	log.Printf("[OrderHandler] Published match event for order %s", match.ID)
	return nil
}

func (h *OrderHandler) ConsumeOrders(ctx context.Context) (<-chan *entities.Order, error) {
	return h.orderChan, nil
}

func (h *kafkaOrderHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *kafkaOrderHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *kafkaOrderHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("[OrderHandler] Starting to consume messages")
	for {
		select {
		case msg := <-claim.Messages():
			var order entities.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("[OrderHandler] Failed to unmarshal order: %v", err)
				continue
			}

			log.Printf("[OrderHandler] Received order: %s", order.ID)

			select {
			case h.orderChan <- &order:
				session.MarkMessage(msg, "")
				log.Printf("[OrderHandler] Processed order: %s", order.ID)
			case <-h.ctx.Done():
				return nil
			}

		case <-h.ctx.Done():
			return nil
		}
	}
}
