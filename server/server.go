package main

import (
	pb "github.com/jumaniyozov/goerpc/proto/gen/todo/v1"
)

type server struct {
	d db
	pb.UnimplementedTodoServiceServer
}
