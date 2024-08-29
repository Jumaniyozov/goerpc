package main

import (
	"context"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	_ "google.golang.org/grpc/encoding/gzip"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		log.Fatalln("usage: server [GRPC_IP_ADDR] [METRICS_IP_ADDR]")
	}

	grpcAddr := args[0]
	httpAddr := args[1]

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}

	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01,
				0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)

	g, ctx := errgroup.WithContext(ctx)
	grpcServer, err := newGrpcServer(lis, srvMetrics)
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
	g.Go(func() error {
		log.Printf("gRPC server listening at %s\n", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to gRPC server: %v\n", err)
			return err
		}
		log.Println("gRPC server shutdown")
		return nil
	})

	metricsServer := newMetricsServer(httpAddr, reg)
	g.Go(func() error {
		log.Printf("metrics server listening at %s\n", httpAddr)
		if err := metricsServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("failed to serve metrics: %v\n", err)
			return err
		}
		log.Println("metrics server shutdown")
		return nil
	})

	select {
	case <-quit:
		break
	case <-ctx.Done():
		break
	}
	cancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer timeoutCancel()
	log.Println("shutting down servers, please wait...")
	grpcServer.GracefulStop()
	metricsServer.Shutdown(timeoutCtx)
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
