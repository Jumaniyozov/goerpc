package main

import (
	"context"
	pb "github.com/jumaniyozov/goerpc/proto/gen/todo/v1"
	"io"
	"time"
)

func (s *server) AddTask(_ context.Context, in *pb.AddTaskRequest) (*pb.AddTaskResponse, error) {
	id, _ := s.d.addTask(in.Description, in.DueDate.AsTime())
	return &pb.AddTaskResponse{Id: id}, nil
}

func (s *server) ListTasks(req *pb.ListTasksRequest, stream pb.TodoService_ListTasksServer) error {
	return s.d.getTasks(func(t any) error {
		task := t.(*pb.Task)
		overdue := task.DueDate != nil && !task.Done &&
			task.DueDate.AsTime().Before(time.Now().UTC())
		err := stream.Send(&pb.ListTasksResponse{
			Task:    task,
			Overdue: overdue,
		})
		return err
	})
}

func (s *server) UpdateTasks(stream pb.TodoService_UpdateTasksServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.UpdateTasksResponse{})
		}
		if err != nil {
			return err
		}
		s.d.updateTask(
			req.Task.Id,
			req.Task.Description,
			req.Task.DueDate.AsTime(),
			req.Task.Done,
		)
	}
}
