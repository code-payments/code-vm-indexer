package geyser

import (
	"context"

	"github.com/code-payments/code-server/pkg/solana"
	"github.com/code-payments/code-server/pkg/solana/cvm"
	"github.com/mr-tron/base58"

	geyserpb "github.com/code-payments/code-vm-indexer/generated/geyser/v1"

	"github.com/code-payments/code-vm-indexer/data/ram"
)

type ProgramAccountUpdateHandler interface {
	// Handle handles account updates from Geyser. Updates are not guaranteed
	// to come in order. Implementations must be idempotent and should not
	// trust the account data passed in. Always refer to finalized blockchain
	// state from another RPC provider.
	Handle(ctx context.Context, update *geyserpb.SubscribeUpdateAccount) error

	// RunBackupWorker runs the backup worker for the handler, which should
	// periodically fill any gaps of data due to missed real-time events from
	// Geyser.
	RunBackupWorker(ctx context.Context) error
}

func initializeProgramAccountUpdateHandlers(conf *conf, solanaClient solana.Client, ramStore ram.Store) map[string]ProgramAccountUpdateHandler {
	ctx := context.TODO()
	return map[string]ProgramAccountUpdateHandler{
		base58.Encode(cvm.PROGRAM_ADDRESS): NewMemoryAccountWithDataUpdateHandler(
			solanaClient,
			ramStore,
			conf.memoryAccountBackkupWorkerInterval.Get(ctx),
			conf.vmAccount.Get(ctx),
		),
	}
}
