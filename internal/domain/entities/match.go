package entities

import (
	"fmt"
	"time"
)

type MatchStatus string

const (
	MatchStatusPending   MatchStatus = "PENDING"
	MatchStatusComplete  MatchStatus = "COMPLETE"
	MatchStatusFailed    MatchStatus = "FAILED"
	MatchStatusCancelled MatchStatus = "CANCELLED"
)

type MatchEvent struct {
	ID          string      `json:"id"`
	BuyOrderID  string      `json:"buy_order_id"`
	SellOrderID string      `json:"sell_order_id"`
	Price       float64     `json:"price"`
	Quantity    int64       `json:"quantity"`
	Time        time.Time   `json:"time"`
	Status      MatchStatus `json:"status"`
}

func NewMatchEvent(buyOrder, sellOrder *Order) (*MatchEvent, error) {
	if err := validateMatchOrders(buyOrder, sellOrder); err != nil {
		return nil, err
	}

	return &MatchEvent{
		ID:          generateMatchID(buyOrder.ID, sellOrder.ID),
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		Price:       calculateMatchPrice(buyOrder, sellOrder),
		Quantity:    calculateMatchQuantity(buyOrder, sellOrder),
		Time:        time.Now(),
		Status:      MatchStatusPending,
	}, nil
}

type MatchResult struct {
	Event         *MatchEvent
	BuyRemaining  int64
	SellRemaining int64
	Success       bool
	Error         error
}

func IsMatchable(buyOrder, sellOrder *Order) bool {
	return buyOrder != nil &&
		sellOrder != nil &&
		buyOrder.Side == Buy &&
		sellOrder.Side == Sell &&
		buyOrder.Price >= sellOrder.Price
}

func validateMatchOrders(buyOrder, sellOrder *Order) error {
	if buyOrder == nil {
		return fmt.Errorf("buy order cannot be nil")
	}
	if sellOrder == nil {
		return fmt.Errorf("sell order cannot be nil")
	}
	if buyOrder.Side != Buy {
		return fmt.Errorf("invalid buy order side: %s", buyOrder.Side)
	}
	if sellOrder.Side != Sell {
		return fmt.Errorf("invalid sell order side: %s", sellOrder.Side)
	}
	if buyOrder.Price < sellOrder.Price {
		return fmt.Errorf("buy price %.2f is less than sell price %.2f", buyOrder.Price, sellOrder.Price)
	}
	if buyOrder.Quantity <= 0 || sellOrder.Quantity <= 0 {
		return fmt.Errorf("order quantities must be positive")
	}
	return nil
}

func calculateMatchPrice(buyOrder, sellOrder *Order) float64 {
	if buyOrder.Time.Before(sellOrder.Time) {
		return buyOrder.Price
	}
	return sellOrder.Price
}

func calculateMatchQuantity(buyOrder, sellOrder *Order) int64 {
	if buyOrder.Quantity < sellOrder.Quantity {
		return buyOrder.Quantity
	}
	return sellOrder.Quantity
}

func generateMatchID(buyOrderID, sellOrderID string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("match_%s_%s_%d", buyOrderID, sellOrderID, timestamp)
}

func (m *MatchEvent) Complete() {
	m.Status = MatchStatusComplete
}

func (m *MatchEvent) Fail() {
	m.Status = MatchStatusFailed
}

func (m *MatchEvent) Cancel() {
	m.Status = MatchStatusCancelled
}

func (m *MatchEvent) IsComplete() bool {
	return m.Status == MatchStatusComplete
}

func (m *MatchEvent) IsFailed() bool {
	return m.Status == MatchStatusFailed
}

func (m *MatchEvent) IsCancelled() bool {
	return m.Status == MatchStatusCancelled
}

func (m *MatchEvent) Validate() error {
	if m.BuyOrderID == "" {
		return fmt.Errorf("buy order ID is required")
	}
	if m.SellOrderID == "" {
		return fmt.Errorf("sell order ID is required")
	}
	if m.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if m.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if m.Time.IsZero() {
		return fmt.Errorf("time is required")
	}
	return nil
}

func (m *MatchEvent) String() string {
	return fmt.Sprintf("Match{ID: %s, Buy: %s, Sell: %s, Price: %.2f, Quantity: %d, Status: %s}",
		m.ID, m.BuyOrderID, m.SellOrderID, m.Price, m.Quantity, m.Status)
}
