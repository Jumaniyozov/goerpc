package main

import (
	"context"
	"fmt"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"slices"
	"time"

	pb "github.com/jumaniyozov/goerpc/proto/gen/todo/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Filter applies a mask (FieldMask) to a msg.
func Filter(msg proto.Message, mask *fieldmaskpb.FieldMask) {
	if mask == nil || len(mask.Paths) == 0 {
		return
	}

	// creates a object to apply reflection on msg
	rft := msg.ProtoReflect()

	// loop over all the fields in rft
	rft.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		if !slices.Contains(mask.Paths, string(fd.Name())) {
			rft.Clear(fd) // clear all the fields that are not contained in mask
		}
		return true
	})
}

// AddTask adds a Task to the database.
// It returns the id of the newly inserted Task or an error.
// If description is empty or if dueDate is in the past,
// it will return an InvalidArgument error.
func (s *server) AddTask(_ context.Context, in *pb.AddTaskRequest) (*pb.AddTaskResponse, error) {
	v, err := protovalidate.New()
	if err != nil {
		fmt.Println("failed to initialize validator:", err)
	}

	if err = v.Validate(in); err != nil {
		return nil, err
	}

	id, err := s.d.addTask(in.Description, in.DueDate.AsTime())

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"unexpected error: %s",
			err.Error(),
		)
	}

	return &pb.AddTaskResponse{Id: id}, nil
}

// ListTasks streams the Tasks present in the database.
// It optionally returns an error if anything went wrong.
// It is cancellable and deadline aware.
func (s *server) ListTasks(req *pb.ListTasksRequest, stream pb.TodoService_ListTasksServer) error {
	ctx := stream.Context()

	return s.d.getTasks(func(t interface{}) error {
		select {
		case <-ctx.Done():
			switch ctx.Err() {
			case context.Canceled:
				log.Printf("request canceled: %s", ctx.Err())
			case context.DeadlineExceeded:
				log.Printf("request deadline exceeded: %s", ctx.Err())
			}
			return ctx.Err()
		// TODO: replace following case by 'default:' replace by 'default:' on production APIs.
		default:
		}

		task := t.(*pb.Task)

		Filter(task, req.Mask)

		overdue := task.DueDate != nil && !task.Done && task.DueDate.AsTime().Before(time.Now().UTC())
		err := stream.Send(&pb.ListTasksResponse{
			Task:    task,
			Overdue: overdue,
		})

		return err
	})
}

// UpdateTasks apply the updates needed to be made.
// It reads the changes to be made through stream.
// It optionally returns an error if anything went wrong.
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
			req.Id,
			req.Description,
			req.DueDate.AsTime(),
			req.Done,
		)
	}
}

// DeleteTasks deletes Tasks in the database.
// It reads the changes to be made through stream.
// For each change being applied it sends back an acknowledgement.
// It optionally returns an error if anything went wrong.
func (s *server) DeleteTasks(stream pb.TodoService_DeleteTasksServer) error {
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		s.d.deleteTask(req.Id)
		stream.Send(&pb.DeleteTasksResponse{})
	}
}
