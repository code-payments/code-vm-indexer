package geyser

import (
	"context"
	"sync"

	"github.com/code-payments/ocp-server/solana"
	"go.uber.org/zap"

	"github.com/code-payments/code-vm-indexer/data/ram"
	geyserpb "github.com/code-payments/code-vm-indexer/generated/geyser/v1"
)

// todo: Generalize the Geyser worker so we can reuse code in this repo and code-server

type Worker struct {
	log  *zap.Logger
	conf *conf

	programUpdatesChan    chan *geyserpb.SubscribeUpdateAccount
	programUpdateHandlers map[string]ProgramAccountUpdateHandler
}

func NewWorker(ctx context.Context, log *zap.Logger, solanaClient solana.Client, ramStore ram.Store, configProvider ConfigProvider) *Worker {
	conf := configProvider()

	return &Worker{
		log:  log,
		conf: conf,

		programUpdatesChan:    make(chan *geyserpb.SubscribeUpdateAccount, conf.programUpdateQueueSize.Get(context.Background())),
		programUpdateHandlers: initializeProgramAccountUpdateHandlers(log, conf, solanaClient, ramStore),
	}
}

func (w *Worker) Run(ctx context.Context) error {
	// Setup event worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < int(w.conf.programUpdateWorkerCount.Get(ctx)); i++ {
		wg.Add(1)
		go func(id int) {
			w.programUpdateWorker(ctx, id)
			wg.Done()
		}(i)
	}

	// Main event loops to consume updates from subscriptions to Geyser that
	// will be processed async
	go func() {
		err := w.consumeGeyserProgramUpdateEvents(ctx)
		if err != nil && err != context.Canceled {
			w.log.Warn("geyser program update consumer terminated unexpectedly", zap.Error(err))
		}
	}()

	// Start backup workers to catch missed events
	for _, handler := range w.programUpdateHandlers {
		go func(handler ProgramAccountUpdateHandler) {
			err := handler.RunBackupWorker(ctx)
			if err != nil && err != context.Canceled {
				w.log.Warn("backup worker terminated unexpectedly", zap.Error(err))
			}
		}(handler)
	}

	// Wait for the service to stop
	select {
	case <-ctx.Done():
	}

	// Gracefully shutdown
	close(w.programUpdatesChan)
	wg.Wait()

	return nil
}
