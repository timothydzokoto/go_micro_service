package main

import (
	"context"
	"log"
	"time"
)

type accountResolver struct {
	server *Server
}

func (ar *accountResolver) Orders(ctx context.Context, obj *Account) ([]*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	orderList, err := ar.server.orderClient.GetOrdersForAccount(ctx, obj.ID)
	if err != nil {
		log.Panicln("Error getting order: ", err)
		return nil, err
	}

	var orders []*Order
	for _, o := range orderList {
		var products []*OrderedProduct
		for _, p := range o.Products {
			products = append(products, &OrderedProduct{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
				Quantity:    int(p.Quantity),
			})
		}

		orders = append(orders, &Order{
			ID:         o.ID,
			CreatedAt:  o.CreatedAt,
			TotalPrice: o.TotalPrice,
			Products:   products,
		})

	}

	return orders, nil
}
