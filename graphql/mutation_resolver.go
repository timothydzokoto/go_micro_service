package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/timothydzokoto/grpc_graphql_microservice/order"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
)

type mutationResolver struct {
	server *Server
}

func (r *mutationResolver) CreateAccount(ctx context.Context, in AccountInput) (*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	a, err := r.server.accountClient.PostAccount(ctx, in.Name)
	if err != nil {
		log.Println("Error creating account: ", err)
		return nil, err
	}

	return &Account{
		ID:   a.ID,
		Name: a.Name,
	}, nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, in ProductInput) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	p, err := r.server.catalogClient.PostProduct(ctx, in.Name, in.Description, in.Price)
	if err != nil {
		log.Println("Error creating product: ", err)
		return nil, err
	}
	return &Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}, nil
}

func (r *mutationResolver) CreateOrder(ctx context.Context, in OrderInput) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	var products []order.OrderedProduct
	for _, p := range in.Products {
		if p.Quantity <= 0 {
			return nil, ErrInvalidParameter
		}
		products = append(products, order.OrderedProduct{
			ID:       p.ID,
			Quantity: uint64(p.Quantity),
		})

	}

	o, err := r.server.orderClient.PostOrder(ctx, in.AccountID, products)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var ops []*OrderedProduct

	for _, op := range o.Products {
		ops = append(ops, &OrderedProduct{
			ID:          op.ID,
			Name:        op.Name,
			Description: op.Description,
			Quantity:    int(op.Quantity),
		})
	}

	return &Order{
		ID:         o.ID,
		Products:   ops,
		TotalPrice: o.TotalPrice,
		CreatedAt:  o.CreatedAt,
	}, nil
}
