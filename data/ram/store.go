package ram

import (
	"context"
	"errors"

	"github.com/code-payments/code-server/pkg/solana/cvm"
)

var (
	ErrStaleState = errors.New("memory item state is stale")
	ErrNotFound   = errors.New("memory item not found")
)

type Store interface {
	// Save updates the database record for a piece of allocated memory
	Save(ctx context.Context, record *Record) error

	// GetAllByMemoryAccount gets all database records for a given memory account
	GetAllByMemoryAccount(ctx context.Context, memoryAccount string) ([]*Record, error)

	// GetAllVirtualAccountsByAddressAndType gets all database records for
	// allocated memory with the provided address and account type
	GetAllVirtualAccountsByAddressAndType(ctx context.Context, address string, accountType cvm.VirtualAccountType) ([]*Record, error)
}
