package ports

import (
	"context"
	"github.com/mehdivijeh/matching-engine/internal/domain/entities"
)

type OrderRepository interface {
	Add(ctx context.Context, order *entities.Order) error
	Remove(ctx context.Context, orderID string) error
	Exists(ctx context.Context, orderID string) bool
}

type MessageBroker interface {
	Start(ctx context.Context) error
	Stop() error
	PublishMatch(ctx context.Context, event *entities.MatchEvent) error
	ConsumeOrders(ctx context.Context) (<-chan *entities.Order, error)
}

type MatchingEngine interface {
	ProcessOrder(ctx context.Context, order *entities.Order) error
}
