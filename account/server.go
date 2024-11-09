//go:generate protoc --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative ./account.proto

package account

import (
	"context"
	"fmt"
	"net"

	"github.com/timothydzokoto/grpc_graphql_microservice/account/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedAccountServiceServer
	service Service
}

func ListenGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	// Create a GRPC server to handle requests
	serv := grpc.NewServer()
	pb.RegisterAccountServiceServer(serv, &grpcServer{
		UnimplementedAccountServiceServer: pb.UnimplementedAccountServiceServer{},
		service:                           s,
	})
	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) PostAccount(ctx context.Context, req *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	a, err := s.service.PostAccount(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &pb.PostAccountResponse{Account: &pb.Account{
		Id:   a.ID,
		Name: a.Name,
	}}, nil
}

func (s *grpcServer) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	a, err := s.service.GetAccount(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetAccountResponse{Account: &pb.Account{
		Id:   a.ID,
		Name: a.Name,
	}}, nil
}

func (s *grpcServer) GetAccounts(ctx context.Context, req *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	res, err := s.service.GetAccounts(ctx, req.Skip, req.Take)
	if err != nil {
		return nil, err
	}
	accounts := []*pb.Account{}
	for _, a := range res {
		accounts = append(accounts, &pb.Account{
			Id:   a.ID,
			Name: a.Name,
		})
	}
	return &pb.GetAccountsResponse{Accounts: accounts}, nil
}
