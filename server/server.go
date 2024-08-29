package main

import (
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	pb "github.com/jumaniyozov/goerpc/proto/gen/todo/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"net/http"
	"os"
)

type server struct {
	d db
	pb.UnimplementedTodoServiceServer
}

func newMetricsServer(httpAddr string, reg *prometheus.Registry) *http.Server {
	httpSrv := &http.Server{Addr: httpAddr}
	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	httpSrv.Handler = m
	return httpSrv
}

func newGrpcServer(lis net.Listener, srvMetrics *grpcprom.ServerMetrics) (*grpc.Server, error) {
	creds, err := credentials.NewServerTLSFromFile("./certs/server_cert.pem", "./certs/server_key.pem")
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime)

	limiter := &simpleLimiter{
		limiter: rate.NewLimiter(5, 10),
	}

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			ratelimit.UnaryServerInterceptor(limiter),
			srvMetrics.UnaryServerInterceptor(),
			auth.UnaryServerInterceptor(validateAuthToken),
			UnaryLogInterceptor,
			logging.UnaryServerInterceptor(logCalls(logger)),
		),
		grpc.ChainStreamInterceptor(
			ratelimit.StreamServerInterceptor(limiter),
			srvMetrics.StreamServerInterceptor(),
			auth.StreamServerInterceptor(validateAuthToken),
			StreamLogInterceptor,
			logging.StreamServerInterceptor(logCalls(logger)),
		),
	}

	s := grpc.NewServer(opts...)

	//registration of endpoints
	pb.RegisterTodoServiceServer(s, &server{
		d: New(),
	})

	return s, nil
}
