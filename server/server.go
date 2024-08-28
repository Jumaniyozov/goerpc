package main

import (
	pb "github.com/jumaniyozov/goerpc/proto/gen/todo/v2"
)

type server struct {
	d db
	pb.UnimplementedTodoServiceServer
}
