package orderbook

import (
	"container/heap"
	"context"
	"sync"

	"github.com/mehdivijeh/matching-engine/internal/domain/entities"
	"github.com/mehdivijeh/matching-engine/internal/domain/ports"
	"github.com/mehdivijeh/matching-engine/pkg/common/errors"
)

type OrderBook struct {
	buyOrders     PriorityQueue
	sellOrders    PriorityQueue
	mutex         sync.RWMutex
	orderRepo     ports.OrderRepository
	messageBroker ports.MessageBroker
}

func NewOrderBook(repo ports.OrderRepository, broker ports.MessageBroker) *OrderBook {
	return &OrderBook{
		buyOrders:     make(PriorityQueue, 0),
		sellOrders:    make(PriorityQueue, 0),
		orderRepo:     repo,
		messageBroker: broker,
	}
}

func (ob *OrderBook) ProcessOrder(ctx context.Context, order *entities.Order) error {
	if err := ob.validateAndStoreOrder(ctx, order); err != nil {
		return err
	}

	ob.mutex.Lock()
	matched := ob.attemptMatch(ctx, order)
	if !matched {
		ob.addToOrderBook(order)
	}
	ob.mutex.Unlock()

	return nil
}

func (ob *OrderBook) validateAndStoreOrder(ctx context.Context, order *entities.Order) error {
	if ob.orderRepo.Exists(ctx, order.ID) {
		return errors.NewDuplicateOrderError(order.ID)
	}
	return ob.orderRepo.Add(ctx, order)
}

func (ob *OrderBook) attemptMatch(ctx context.Context, order *entities.Order) bool {
	oppositeQueue := ob.getOppositeQueue(order.Side)
	if oppositeQueue.Len() == 0 {
		return false
	}

	bestMatch := ob.findBestMatch(oppositeQueue)
	if !entities.IsMatchable(order, bestMatch) {
		return false
	}

	ob.mutex.Unlock()
	matchEvent, err := entities.NewMatchEvent(order, bestMatch)
	if err != nil {
		ob.mutex.Lock()
		heap.Push(oppositeQueue, &OrderItem{order: bestMatch})
		return false
	}

	if err = ob.messageBroker.PublishMatch(ctx, matchEvent); err != nil {
		ob.mutex.Lock()
		heap.Push(oppositeQueue, &OrderItem{order: bestMatch})
		return false
	}
	ob.mutex.Lock()

	heap.Pop(oppositeQueue)
	ob.handleRemainingQuantities(order, bestMatch, matchEvent, oppositeQueue)
	return true
}

func (ob *OrderBook) getOppositeQueue(side entities.OrderSide) *PriorityQueue {
	if side == entities.Buy {
		return &ob.sellOrders
	}
	return &ob.buyOrders
}

func (ob *OrderBook) findBestMatch(queue *PriorityQueue) *entities.Order {
	if item, ok := queue.Peek(); ok {
		return item.order
	}
	return nil
}

func (ob *OrderBook) handleRemainingQuantities(order, matchedOrder *entities.Order, matchEvent *entities.MatchEvent, queue *PriorityQueue) {
	if remainingQty := order.Quantity - matchEvent.Quantity; remainingQty > 0 {
		order.Quantity = remainingQty
		ob.addToOrderBook(order)
	}

	if remainingQty := matchedOrder.Quantity - matchEvent.Quantity; remainingQty > 0 {
		matchedOrder.Quantity = remainingQty
		heap.Push(queue, &OrderItem{order: matchedOrder})
	}
}

func (ob *OrderBook) addToOrderBook(order *entities.Order) {
	item := &OrderItem{order: order}
	if order.Side == entities.Buy {
		heap.Push(&ob.buyOrders, item)
	} else {
		heap.Push(&ob.sellOrders, item)
	}
}

func (ob *OrderBook) GetBestBid() (float64, bool) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	if item, ok := ob.buyOrders.Peek(); ok {
		return item.order.Price, true
	}
	return 0, false
}

func (ob *OrderBook) GetBestAsk() (float64, bool) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	if item, ok := ob.sellOrders.Peek(); ok {
		return item.order.Price, true
	}
	return 0, false
}

func (ob *OrderBook) GetOrderBookDepth() (int, int) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	return ob.buyOrders.Len(), ob.sellOrders.Len()
}
