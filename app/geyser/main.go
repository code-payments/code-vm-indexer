package main

import (
	"context"
	"sync"

	"github.com/code-payments/ocp-server/config/env"
	grpcapp "github.com/code-payments/ocp-server/grpc/app"
	"github.com/code-payments/ocp-server/metrics"
	"github.com/code-payments/ocp-server/solana"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	indexerapp "github.com/code-payments/code-vm-indexer/app"
	"github.com/code-payments/code-vm-indexer/geyser"
)

const (
	solanaRpcEndpointConfigEnvName = "SOLANA_RPC_ENDPOINT"
	defaultSolanaRpcEndpoint       = "http://localhost:8899"
)

type app struct {
	geyserWorker *geyser.Worker

	shutdown         sync.Once
	shutdownCh       chan struct{}
	workerCancelFunc context.CancelFunc
}

// Init implements grpcapp.App.Init
//
// todo: Cleanup gRPC app package (ie. no hardcoded metrics provider)
func (a *app) Init(logger *zap.Logger, _ metrics.Provider, _ grpcapp.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	a.workerCancelFunc = cancel
	a.shutdownCh = make(chan struct{})

	solanaRpcEndpointConfig := env.NewStringConfig(solanaRpcEndpointConfigEnvName, defaultSolanaRpcEndpoint)

	solanaClient := solana.New(solanaRpcEndpointConfig.Get(ctx))

	dataProvider, err := indexerapp.NewDataProvider()
	if err != nil {
		return err
	}

	a.geyserWorker = geyser.NewWorker(ctx, logger, solanaClient, dataProvider.Ram, geyser.WithEnvConfigs())
	go a.geyserWorker.Run(ctx)

	return nil
}

// RegisterWithGRPC implements grpcapp.App.RegisterWithGRPC
func (a *app) RegisterWithGRPC(server *grpc.Server) {
}

// ShutdownChan implements grpcapp.App.ShutdownChan
func (a *app) ShutdownChan() <-chan struct{} {
	return a.shutdownCh
}

// Stop implements grpcapp.App.Stop
func (a *app) Stop() {
	a.shutdown.Do(func() {
		close(a.shutdownCh)
		a.workerCancelFunc()
	})
}

func main() {
	if err := grpcapp.Run(
		&app{},
	); err != nil {
		log, _ := zap.NewProduction()
		log.Fatal("error running indexer geyser worker", zap.Error(err))
	}
}
