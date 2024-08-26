package main

import (
	"context"
	"fmt"
	pb "github.com/jumaniyozov/goerpc/proto/gen/todo/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatalln("usage: client [IP_ADDR]")
	}
	addr := args[0]
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c := pb.NewTodoServiceClient(conn)

	fmt.Println("--------ADD--------")
	dueDate := time.Now().Add(5 * time.Second)
	addTask(c, "This is a task", dueDate)
	fmt.Println("-------------------")

	fmt.Println("--------LIST-------")
	printTasks(c)
	fmt.Println("-------------------")

	fmt.Println("-------UPDATE------")
	updateTasks(c, []*pb.UpdateTasksRequest{
		{Task: &pb.Task{Id: 1, Description: "A better name for the task"}},
		{Task: &pb.Task{Id: 2, DueDate: timestamppb.New(dueDate.Add(5 * time.Hour))}},
		{Task: &pb.Task{Id: 3, Done: true}},
	}...)
	printTasks(c)
	fmt.Println("-------------------")

	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Fatalf("unexpected error: %v", err)
		}
	}(conn)
}

func addTask(c pb.TodoServiceClient, description string, dueDate time.Time) uint64 {
	req := &pb.AddTaskRequest{
		Description: description,
		DueDate:     timestamppb.New(dueDate),
	}

	res, err := c.AddTask(context.Background(), req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("added task: %d\n", res.Id)
	return res.Id
}

func printTasks(c pb.TodoServiceClient) {
	req := &pb.ListTasksRequest{}
	stream, err := c.ListTasks(context.Background(), req)

	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("unexpected error: %v", err)
		}

		fmt.Println(res.Task.String(), "overdue: ", res.Overdue)
	}

}

func updateTasks(c pb.TodoServiceClient, reqs ...*pb.UpdateTasksRequest) {
	stream, err := c.UpdateTasks(context.Background())
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
	for _, req := range reqs {
		err := stream.Send(req)
		if err != nil {
			return
		}
		if err != nil {
			log.Fatalf("unexpected error: %v", err)
		}
		if req.Task != nil {
			fmt.Printf("updated task with id: %d\n", req.Task.Id)
		}
	}
	if _, err = stream.CloseAndRecv(); err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
}
