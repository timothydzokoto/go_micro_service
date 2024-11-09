package order

import (
	"context"
	"log"
	"time"

	"github.com/timothydzokoto/grpc_graphql_microservice/order/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := pb.NewOrderServiceClient(conn)
	return &Client{conn, c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {
	postProducts := []*pb.PostOrderRequest_OrderProduct{}
	for _, p := range products {
		postProducts = append(postProducts, &pb.PostOrderRequest_OrderProduct{
			ProductId: p.ID,
			Quantity:  p.Quantity,
		})
	}
	r, err := c.service.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: accountID,
		Products:  postProducts,
	})
	if err != nil {
		return nil, err
	}

	newOrder := r.Order
	newOrderCreatedAt := time.Time{}
	if err := newOrderCreatedAt.UnmarshalBinary(newOrder.CreatedAt); err != nil {
		return nil, err
	}

	return &Order{
		ID:         newOrder.Id,
		CreatedAt:  newOrderCreatedAt,
		AccountID:  newOrder.AccountId,
		TotalPrice: newOrder.TotalPrice,
		Products:   products,
	}, nil

}

func (c *Client) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	r, err := c.service.GetOrderForAccount(ctx, &pb.GetOrderForAccountRequest{AccountId: accountID})
	if err != nil {
		log.Fatal("Error getting order: ", err)
		return nil, err
	}

	orders := []Order{}
	for _, orderProto := range r.Orders {
		newOrder := Order{
			ID:         orderProto.Id,
			AccountID:  orderProto.AccountId,
			TotalPrice: orderProto.TotalPrice,
		}
		newOrder.CreatedAt = time.Time{}
		newOrder.CreatedAt.UnmarshalBinary(orderProto.CreatedAt)

		products := []OrderedProduct{}
		for _, productProto := range orderProto.Products {

			products = append(products, OrderedProduct{
				ID:          productProto.Id,
				Name:        productProto.Name,
				Description: productProto.Description,
				Price:       productProto.Price,
				Quantity:    productProto.Quantity,
			})
		}
		newOrder.Products = products
		orders = append(orders, newOrder)
	}
	return orders, nil
}
