package repository

import (
	"context"
	"sync"

	"github.com/mehdivijeh/matching-engine/internal/domain/entities"
)

//this is a sample stateful database that handle duplicate orders, also we can use redis or other distributed tools

type InMemoryOrderRepository struct {
	orders map[string]*entities.Order
	mutex  sync.RWMutex
}

func NewOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[string]*entities.Order),
	}
}

func (r *InMemoryOrderRepository) Add(ctx context.Context, order *entities.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepository) Remove(ctx context.Context, orderID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.orders, orderID)
	return nil
}

func (r *InMemoryOrderRepository) Exists(ctx context.Context, orderID string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.orders[orderID]
	return exists
}
