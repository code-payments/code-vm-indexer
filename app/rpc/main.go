package main

import (
	"sync"

	grpcapp "github.com/code-payments/code-server/pkg/grpc/app"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	indexerpb "github.com/code-payments/code-vm-indexer/generated/indexer/v1"

	indexerapp "github.com/code-payments/code-vm-indexer/app"
	"github.com/code-payments/code-vm-indexer/rpc"
)

type app struct {
	indexerServer indexerpb.IndexerServer

	shutdown   sync.Once
	shutdownCh chan struct{}
}

// Init implements grpcapp.App.Init
//
// todo: Cleanup gRPC app package (ie. no hardcoded metrics provider)
func (a *app) Init(_ grpcapp.Config, metricsProvider *newrelic.Application) error {
	a.shutdownCh = make(chan struct{})

	dataProvider, err := indexerapp.NewDataProvider()
	if err != nil {
		return err
	}

	a.indexerServer = rpc.NewServer(dataProvider.Ram)

	return nil
}

// RegisterWithGRPC implements grpcapp.App.RegisterWithGRPC
func (a *app) RegisterWithGRPC(server *grpc.Server) {
	indexerpb.RegisterIndexerServer(server, a.indexerServer)
}

// ShutdownChan implements grpcapp.App.ShutdownChan
func (a *app) ShutdownChan() <-chan struct{} {
	return a.shutdownCh
}

// Stop implements grpcapp.App.Stop
func (a *app) Stop() {
	a.shutdown.Do(func() {
		close(a.shutdownCh)
	})
}

func main() {
	if err := grpcapp.Run(
		&app{},
	); err != nil {
		logrus.WithError(err).Fatal("error running indexer rpc")
	}
}
