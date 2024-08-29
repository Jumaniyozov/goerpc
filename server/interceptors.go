package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
)

const authTokenKey string = "auth_token"
const authTokenValue string = "authd"

type AuthFunc func(ctx context.Context) (context.Context, error)

func validateAuthToken(ctx context.Context) (context.Context, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if t, ok := md["auth_token"]; ok {
		switch {
		case len(t) != 1:
			return nil, status.Errorf(
				codes.InvalidArgument,
				fmt.Sprintf("%s should contain only 1 value", authTokenKey),
			)
		case t[0] != "authd":
			return nil, status.Errorf(
				codes.Unauthenticated,
				fmt.Sprintf("incorrect %s", authTokenKey),
			)
		}
	} else {
		return nil, status.Errorf(
			codes.Unauthenticated,
			fmt.Sprintf("failed to get %s", authTokenKey),
		)
	}
	return ctx, nil
}

func UnaryLogInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	log.Println(info.FullMethod, "called")
	return handler(ctx, req)
}

func StreamLogInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println(info.FullMethod, "called")
	return handler(srv, ss)
}
