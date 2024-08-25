package main

import (
	"context"
	"fmt"
	"gRPC_gateway/proto"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	grpcAddress = "localhost:50051"
	httpAddress = "localhost:8080"
)

type server struct {
	proto.UnimplementedGatewayServer
}

func (s *server) GetExample(ctx context.Context, in *proto.Message) (*proto.Message, error) {
	fmt.Println(in)
	return &proto.Message{Id: in.Id}, nil
}

func startGrpcServer() error {
	grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))

	reflection.Register(grpcServer)

	proto.RegisterGatewayServer(grpcServer, &server{})

	list, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return err
	}

	return grpcServer.Serve(list)
}

func startHttpServer(ctx context.Context) error {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := proto.RegisterGatewayHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		panic(err)
	}

	log.Printf("http server listening on: %s", httpAddress)

	return http.ListenAndServe(httpAddress, mux)
}

func main() {
	ctx := context.Background()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		if err := startGrpcServer(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()

		if err := startHttpServer(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}
