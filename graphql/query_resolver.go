package main

import (
	"context"
	"errors"
	"log"
	"time"
)

type queryResolver struct {
	server *Server
}

func (qr *queryResolver) Accounts(ctx context.Context, pagination *PaginationInput, id *string) ([]*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	if id != nil {
		r, err := qr.server.accountClient.GetAccount(ctx, *id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return []*Account{{
			ID:   r.ID,
			Name: r.Name,
		}}, nil
	}

	skip, take := uint64(0), uint64(10)
	if pagination != nil {
		skip, take = pagination.bounds()
	}

	accountList, err := qr.server.accountClient.GetAccounts(ctx, skip, take)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var accounts []*Account
	for _, a := range accountList {
		accounts = append(accounts, &Account{
			ID:   a.ID,
			Name: a.Name,
		})
	}

	return accounts, nil
}

func (qr *queryResolver) Products(ctx context.Context, pagination *PaginationInput, query *string, id *string) ([]*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	if id != nil {
		r, err := qr.server.catalogClient.GetProduct(ctx, *id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return []*Product{{
			ID:          r.ID,
			Name:        r.Name,
			Price:       r.Price,
			Description: r.Description,
		}}, nil
	}

	skip, take := uint64(0), uint64(10)
	if pagination != nil {
		skip, take = pagination.bounds()
	}

	q := ""
	if query != nil {
		q = *query
	}

	productList, err := qr.server.catalogClient.GetProducts(ctx, skip, take, q, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var products []*Product
	for _, p := range productList {
		products = append(products, &Product{
			ID:          p.ID,
			Name:        p.Name,
			Price:       p.Price,
			Description: p.Description,
		})
	}

	return products, nil

}

func (p PaginationInput) bounds() (uint64, uint64) {
	skip := uint64(0)
	take := uint64(0)
	if p.Skip != 0 {
		skip = uint64(p.Skip)
	}
	if p.Take != 0 {
		take = uint64(p.Take)
	}

	return skip, take
}

func (qr *queryResolver) Orders(ctx context.Context, pagination *PaginationInput, id *string) ([]*Order, error) {
	return nil, errors.New("No order")
}
