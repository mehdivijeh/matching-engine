package orderbook

import (
	"github.com/mehdivijeh/matching-engine/internal/domain/entities"
)

type OrderItem struct {
	order *entities.Order
	index int
}

type PriorityQueue []*OrderItem

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].order.Side == entities.Buy {
		return pq[i].order.Price > pq[j].order.Price
	}
	return pq[i].order.Price < pq[j].order.Price
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*OrderItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq PriorityQueue) Peek() (*OrderItem, bool) {
	if len(pq) == 0 {
		return nil, false
	}
	return pq[0], true
}
