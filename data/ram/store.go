package ram

import (
	"context"
	"errors"

	"github.com/code-payments/code-server/pkg/solana/cvm"
)

var (
	ErrAccountNotFound = errors.New("memory account not found")
	ErrItemNotFound    = errors.New("memory item not found")
	ErrStaleState      = errors.New("memory item state is stale")
)

type Store interface {
	// Save updates the database record for a piece of allocated memory
	Save(ctx context.Context, record *Record) error

	// GetAllMemoryAccounts gets all unique memory account addresses
	GetAllMemoryAccounts(ctx context.Context, vm string) ([]string, error)

	// GetAllByMemoryAccount gets all database records for a given memory account
	GetAllByMemoryAccount(ctx context.Context, memoryAccount string) ([]*Record, error)

	// GetAllVirtualAccountsByAddressAndType gets all database records for
	// allocated memory with the provided address and account type
	GetAllVirtualAccountsByAddressAndType(ctx context.Context, address string, accountType cvm.VirtualAccountType) ([]*Record, error)
}
