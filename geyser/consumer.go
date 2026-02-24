package geyser

import (
	"context"
	"errors"
	"time"

	"github.com/mr-tron/base58"
	"go.uber.org/zap"
)

func (w *Worker) consumeGeyserProgramUpdateEvents(ctx context.Context) error {
	log := w.log.With(zap.String("method", "consumeGeyserProgramUpdateEvents"))

	for {
		// Is the service stopped?
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := w.subscribeToProgramUpdatesFromGeyser(ctx, w.conf.grpcPluginEndpoint.Get(ctx), w.conf.grpcPluginXToken.Get(ctx))
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Warn("program update consumer unexpectedly terminated", zap.Error(err))
		}

		// Avoid spamming new connections when something is wrong
		time.Sleep(time.Second)
	}
}

func (w *Worker) programUpdateWorker(ctx context.Context, id int) {
	log := w.log.With(
		zap.String("method", "programUpdateWorker"),
		zap.Int("worker_id", id),
	)

	log.Debug("worker started")

	defer func() {
		log.Debug("worker stopped")
	}()

	for update := range w.programUpdatesChan {
		func() {
			base58PublicKey := base58.Encode(update.Account.Pubkey)
			base58ProgramAddress := base58.Encode(update.Account.Owner)

			log := log.With(
				zap.String("account", base58PublicKey),
				zap.String("program", base58ProgramAddress),
				zap.Uint64("slot", update.Slot),
			)
			if len(update.Account.TxnSignature) > 0 {
				log = log.With(zap.String("transaction", base58.Encode(update.Account.TxnSignature)))
			}

			handler, ok := w.programUpdateHandlers[base58ProgramAddress]
			if !ok {
				log.Debug("not handling update from program")
				return
			}

			if err := handler.Handle(ctx, update); err != nil {
				log.Warn("failed to process program account update", zap.Error(err))
			}
		}()
	}
}
