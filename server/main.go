package main

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	pb "github.com/jumaniyozov/goerpc/proto/gen/todo/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
	"log"
	"net"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatalln("usage: server [IP_ADDR]")
	}

	addr := args[0]
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	defer func(lis net.Listener) {
		if err := lis.Close(); err != nil {
			log.Fatalf("unexpected error: %v", err)
		}
	}(lis)

	log.Printf("listening at %s\n", addr)

	creds, err := credentials.NewServerTLSFromFile("./certs/server_cert.pem", "./certs/server_key.pem")
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime)
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(auth.UnaryServerInterceptor(validateAuthToken), UnaryLogInterceptor, logging.UnaryServerInterceptor(logCalls(logger))),
		grpc.ChainStreamInterceptor(auth.StreamServerInterceptor(validateAuthToken), StreamLogInterceptor, logging.StreamServerInterceptor(logCalls(logger))),
	}
	s := grpc.NewServer(opts...)

	//registration of endpoints
	pb.RegisterTodoServiceServer(s, &server{
		d: New(),
	})

	defer s.Stop()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
