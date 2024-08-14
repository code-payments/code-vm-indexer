package geyser

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/code-payments/code-server/pkg/solana/cvm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	geyserpb "github.com/code-payments/code-vm-indexer/generated/geyser/v1"
)

const (
	defaultStreamSubscriptionTimeout = time.Minute
)

var (
	ErrSubscriptionFallenBehind = errors.New("subscription stream fell behind")
	ErrTimeoutReceivingUpdate   = errors.New("timed out receiving update")
)

func newGeyserClient(ctx context.Context, endpoint string) (geyserpb.GeyserClient, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := geyserpb.NewGeyserClient(conn)

	// Unfortunately the RPCs we use no longer support hearbeats. We'll let each
	// individual subscriber determine what an appropriate timeout to receive a
	// message should be.
	/*
		heartbeatResp, err := client.GetHeartbeatInterval(ctx, &geyserpb.EmptyRequest{})
		if err != nil {
			return nil, 0, errors.Wrap(err, "error getting heartbeat interval")
		}

		heartbeatTimeout := time.Duration(2 * heartbeatResp.HeartbeatIntervalMs * uint64(time.Millisecond))
	*/

	return client, nil
}

func boundedProgramUpdateRecv(ctx context.Context, streamer geyserpb.Geyser_SubscribeProgramUpdatesClient, timeout time.Duration) (update *geyserpb.TimestampedAccountUpdate, err error) {
	done := make(chan struct{})
	go func() {
		update, err = streamer.Recv()
		close(done)
	}()

	select {
	case <-time.After(timeout):
		return nil, ErrTimeoutReceivingUpdate
	case <-done:
		return update, err
	}
}

func (w *Worker) subscribeToProgramUpdatesFromGeyser(ctx context.Context, endpoint string) error {
	log := w.log.WithField("method", "subscribeToProgramUpdatesFromGeyser")
	log.Debug("subscription started")

	defer func() {
		log.Debug("subscription stopped")
	}()

	client, err := newGeyserClient(ctx, endpoint)
	if err != nil {
		return errors.Wrap(err, "error creating client")
	}

	streamer, err := client.SubscribeProgramUpdates(ctx, &geyserpb.SubscribeProgramsUpdatesRequest{
		Programs: [][]byte{cvm.PROGRAM_ID},
	})
	if err != nil {
		return errors.Wrap(err, "error opening subscription stream")
	}

	for {
		update, err := boundedProgramUpdateRecv(ctx, streamer, defaultStreamSubscriptionTimeout)
		if err != nil {
			return errors.Wrap(err, "error receiving update")
		}

		messageAge := time.Since(update.Ts.AsTime())
		if messageAge > defaultStreamSubscriptionTimeout {
			log.WithField("message_age", messageAge).Warn(ErrSubscriptionFallenBehind.Error())
			return ErrSubscriptionFallenBehind
		}

		// Ignore startup updates. We only care about real-time updates due to
		// transactions.
		if update.AccountUpdate.IsStartup {
			continue
		}

		// Queue program updates for async processing. Most importantly, we need to
		// process messages from the gRPC subscription as fast as possible to avoid
		// backing up the Geyser plugin, which kills this subscription and we end up
		// missing updates.
		select {
		case w.programUpdatesChan <- update.AccountUpdate:
		default:
			log.Warn("dropping update because queue is full")
		}
	}
}
