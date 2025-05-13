package geyser

import (
	"context"
	"time"

	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	geyserpb "github.com/code-payments/code-vm-indexer/generated/geyser/v1"

	"github.com/code-payments/code-server/pkg/solana/cvm"
)

const (
	defaultStreamSubscriptionTimeout = time.Minute
)

var (
	ErrTimeoutReceivingUpdate = errors.New("timed out receiving update")
)

func newGeyserClient(endpoint, xToken string) (geyserpb.GeyserClient, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if len(xToken) > 0 {
		opts = append(
			opts,
			grpc.WithUnaryInterceptor(newXTokenUnaryClientInterceptor(xToken)),
			grpc.WithStreamInterceptor(newXTokenStreamClientInterceptor(xToken)),
		)
	}

	conn, err := grpc.NewClient(endpoint, opts...)
	if err != nil {
		return nil, err
	}

	client := geyserpb.NewGeyserClient(conn)

	return client, nil
}

func boundedRecv(ctx context.Context, streamer geyserpb.Geyser_SubscribeClient, timeout time.Duration) (update *geyserpb.SubscribeUpdate, err error) {
	done := make(chan struct{})
	go func() {
		update, err = streamer.Recv()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, ErrTimeoutReceivingUpdate
	case <-done:
		return update, err
	}
}

func (w *Worker) subscribeToProgramUpdatesFromGeyser(ctx context.Context, endpoint, xToken string) error {
	log := w.log.WithField("method", "subscribeToProgramUpdatesFromGeyser")
	log.Debug("subscription started")

	defer func() {
		log.Debug("subscription stopped")
	}()

	client, err := newGeyserClient(endpoint, xToken)
	if err != nil {
		return errors.Wrap(err, "error creating client")
	}

	streamer, err := client.Subscribe(ctx)
	if err != nil {
		return errors.Wrap(err, "error opening subscription stream")
	}

	req := &geyserpb.SubscribeRequest{
		Accounts: make(map[string]*geyserpb.SubscribeRequestFilterAccounts),
	}
	req.Accounts["accounts_subscription"] = &geyserpb.SubscribeRequestFilterAccounts{
		Owner: []string{base58.Encode(cvm.PROGRAM_ID)},
	}
	finalizedCommitmentLevel := geyserpb.CommitmentLevel_FINALIZED
	req.Commitment = &finalizedCommitmentLevel
	err = streamer.Send(req)
	if err != nil {
		return errors.Wrap(err, "error sending subscription request")
	}

	for {
		update, err := boundedRecv(ctx, streamer, defaultStreamSubscriptionTimeout)
		if err != nil {
			return errors.Wrap(err, "error recieving update")
		}

		accountUpdate := update.GetAccount()
		if accountUpdate == nil {
			continue
		}

		// Ignore startup updates. We only care about real-time updates due to
		// transactions.
		if accountUpdate.IsStartup {
			continue
		}

		// Queue program updates for async processing. Most importantly, we need to
		// process messages from the gRPC subscription as fast as possible to avoid
		// backing up the Geyser plugin, which kills this subscription and we end up
		// missing updates.
		select {
		case w.programUpdatesChan <- accountUpdate:
		default:
			log.Warn("dropping update because queue is full")
		}
	}
}

func newXTokenUnaryClientInterceptor(xToken string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = withXToken(ctx, xToken)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func newXTokenStreamClientInterceptor(xToken string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx = withXToken(ctx, xToken)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func withXToken(ctx context.Context, xToken string) context.Context {
	md := metadata.Pairs("x-token", xToken)
	return metadata.NewOutgoingContext(ctx, md)
}
