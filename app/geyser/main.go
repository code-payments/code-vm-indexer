package main

import (
	"context"
	"sync"

	"github.com/code-payments/code-server/pkg/config/env"
	grpcapp "github.com/code-payments/code-server/pkg/grpc/app"
	"github.com/code-payments/code-server/pkg/solana"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
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
func (a *app) Init(_ grpcapp.Config, metricsProvider *newrelic.Application) error {
	ctx, cancel := context.WithCancel(context.Background())
	a.workerCancelFunc = cancel
	a.shutdownCh = make(chan struct{})

	solanaRpcEndpointConfig := env.NewStringConfig(solanaRpcEndpointConfigEnvName, defaultSolanaRpcEndpoint)

	solanaClient := solana.New(solanaRpcEndpointConfig.Get(ctx))

	dataProvider, err := indexerapp.NewDataProvider()
	if err != nil {
		return err
	}

	a.geyserWorker = geyser.NewWorker(ctx, solanaClient, dataProvider.Ram, geyser.WithEnvConfigs())
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
		logrus.WithError(err).Fatal("error running indexer geyser worker")
	}
}
