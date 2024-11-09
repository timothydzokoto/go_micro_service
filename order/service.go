package order

import (
	"context"
	"time"

	"github.com/segmentio/ksuid"
)

type Service interface {
	PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error)
	GetOrdersForAccount(ctx context.Context, id string) ([]*Order, error)
	GetOrders(ctx context.Context, skip uint64, take uint64) ([]Order, error)
}

type Order struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	AccountID  string    `json:"account_id"`
	TotalPrice float64   `json:"total_price"`
	Products   []OrderedProduct
}

type OrderedProduct struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    uint64  `json:"quantity"`
}

type orderService struct {
	repository Repository
}

func NewService(r Repository) *orderService {
	return &orderService{r}
}

func (s *orderService) PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {
	o := &Order{
		ID:         ksuid.New().String(),
		AccountID:  accountID,
		CreatedAt:  time.Now().UTC(),
		TotalPrice: 0.0,
		Products:   products,
	}
	for _, p := range products {
		o.TotalPrice += p.Price * float64(p.Quantity)
	}
	if err := s.repository.PutOrder(ctx, *o); err != nil {
		return nil, err
	}

	return o, nil
}

func (s *orderService) GetOrdersForAccount(ctx context.Context, id string) ([]*Order, error) {
	return s.repository.GetOrderForAccount(ctx, id)
}

func (s *orderService) GetOrders(ctx context.Context, skip uint64, take uint64) ([]Order, error) {
	if skip > 100 || (take == 0 && skip == 0) {
		take = 100
	}
	return s.repository.ListOrders(ctx, skip, take)
}
