package geyser

import (
	"context"
	"errors"
	"time"

	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
)

func (w *Worker) consumeGeyserProgramUpdateEvents(ctx context.Context) error {
	log := w.log.WithField("method", "consumeGeyserProgramUpdateEvents")

	for {
		// Is the service stopped?
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := w.subscribeToProgramUpdatesFromGeyser(ctx, w.conf.grpcPluginEndpoint.Get(ctx))
		if err != nil && !errors.Is(err, context.Canceled) {
			log.WithError(err).Warn("program update consumer unexpectedly terminated")
		}

		// Avoid spamming new connections when something is wrong
		time.Sleep(time.Second)
	}
}

func (w *Worker) programUpdateWorker(ctx context.Context, id int) {
	log := w.log.WithFields(logrus.Fields{
		"method":    "programUpdateWorker",
		"worker_id": id,
	})

	log.Debug("worker started")

	defer func() {
		log.Debug("worker stopped")
	}()

	for update := range w.programUpdatesChan {
		func() {
			base58PublicKey := base58.Encode(update.Pubkey)
			base58ProgramAddress := base58.Encode(update.Owner)

			log := log.WithFields(logrus.Fields{
				"account": base58PublicKey,
				"program": base58ProgramAddress,
				"slot":    update.Slot,
			})
			if update.TxSignature != nil {
				log = log.WithField("transaction", *update.TxSignature)
			}

			handler, ok := w.programUpdateHandlers[base58ProgramAddress]
			if !ok {
				log.Debug("not handling update from program")
				return
			}

			if err := handler.Handle(ctx, update); err != nil {
				log.WithError(err).Warn("failed to process program account update")
			}
		}()
	}
}
