package errors

import "fmt"

type DuplicateOrderError struct {
	OrderID string
}

func (e *DuplicateOrderError) Error() string {
	return fmt.Sprintf("duplicate order ID: %s", e.OrderID)
}

func NewDuplicateOrderError(orderID string) error {
	return &DuplicateOrderError{OrderID: orderID}
}
