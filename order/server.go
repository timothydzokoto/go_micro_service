//go:generate protoc --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative order.proto

package order

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/timothydzokoto/grpc_graphql_microservice/account"
	"github.com/timothydzokoto/grpc_graphql_microservice/catalog"
	"github.com/timothydzokoto/grpc_graphql_microservice/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service       Service
	accountClient *account.Client
	catalogClient *catalog.Client
}

func ListenGRPC(s Service, accountURL, catalogURL string, port int) error {
	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		return err
	}

	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		accountClient.Close()
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		return err
	}

	serv := grpc.NewServer()
	pb.RegisterOrderServiceServer(serv, &grpcServer{
		UnimplementedOrderServiceServer: pb.UnimplementedOrderServiceServer{},
		service:                         s,
		accountClient:                   accountClient,
		catalogClient:                   catalogClient,
	})
	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) PostProduct(ctx context.Context, r *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	_, err := s.accountClient.GetAccount(ctx, r.AccountId)
	if err != nil {
		log.Println("Error getting account: ", err)
		return nil, err
	}

	productsID := []string{}

	orderedProducts, err := s.catalogClient.GetProducts(ctx, 0, 0, "", productsID)
	if err != nil {
		log.Println("Error getting products: ", err)
		return nil, errors.New("product not found")
	}

	products := []OrderedProduct{}
	for _, p := range orderedProducts {
		product := OrderedProduct{
			ID:          p.ID,
			Quantity:    0,
			Price:       p.Price,
			Name:        p.Name,
			Description: p.Description,
		}

		for _, rp := range r.Products {
			if rp.ProductId == p.ID {
				product.Quantity = rp.Quantity
				break
			}
		}

		if product.Quantity == 0 {
			products = append(products, product)
		}
	}

	order, err := s.service.PostOrder(ctx, r.AccountId, products)
	if err != nil {
		log.Println("Error creating order: ", err)
		return nil, errors.New("error creating order")
	}
	orderProto := &pb.Order{
		Id:         order.ID,
		AccountId:  order.AccountID,
		TotalPrice: order.TotalPrice,
		Products:   []*pb.Order_OrderedProduct{},
	}
	orderProto.CreatedAt, err = order.CreatedAt.MarshalBinary()
	if err != nil {
		log.Println("Error marshalling time: ", err)
		return nil, errors.New("error marshalling time")
	}
	for _, p := range order.Products {
		orderProto.Products = append(orderProto.Products, &pb.Order_OrderedProduct{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		})
	}
	return &pb.PostOrderResponse{Order: orderProto}, nil

}

func (s *grpcServer) GetOrderForAccount(ctx context.Context, r *pb.GetOrderForAccountRequest) (*pb.GetOrderForAccountResponse, error) {
	accountOrders, err := s.service.GetOrdersForAccount(ctx, r.AccountId)
	if err != nil {
		log.Println("Error getting order: ", err)
		return nil, errors.New("error getting order")
	}

	productIDMap := map[string]bool{}
	for _, o := range accountOrders {
		for _, p := range o.Products {
			productIDMap[p.ID] = true
		}
	}

	productIDs := []string{}
	for k := range productIDMap {
		productIDs = append(productIDs, k)
	}

	products, err := s.catalogClient.GetProducts(ctx, 0, 0, "", productIDs)
	if err != nil {
		log.Println("Error getting products: ", err)
		return nil, errors.New("error getting products")
	}

	orders := []*pb.Order{}
	for _, o := range accountOrders {
		op := &pb.Order{
			Id:         o.ID,
			AccountId:  o.AccountID,
			TotalPrice: o.TotalPrice,
			Products:   []*pb.Order_OrderedProduct{},
		}
		op.CreatedAt, err = o.CreatedAt.MarshalBinary()
		if err != nil {
			log.Println("Error marshalling time: ", err)
			return nil, errors.New("error marshalling time")
		}

		for _, product := range o.Products {
			for _, p := range products {
				if p.ID == product.ID {
					product.Name = p.Name
					product.Description = p.Description
					product.Price = p.Price
					break
				}

			}
			op.Products = append(op.Products, &pb.Order_OrderedProduct{
				Id:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Quantity:    product.Quantity,
			})
		}
		orders = append(orders, op)
	}
	return &pb.GetOrderForAccountResponse{Orders: orders}, nil
}
