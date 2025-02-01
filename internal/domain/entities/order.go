package entities

import (
	"time"
)

type OrderSide string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

type Order struct {
	ID       string    `json:"id"`
	Price    float64   `json:"price"`
	Quantity int64     `json:"quantity"`
	Side     OrderSide `json:"side"`
	Time     time.Time `json:"time"`
}
