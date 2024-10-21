package postgres

import (
	"context"
	"database/sql"

	"github.com/code-payments/code-server/pkg/solana/cvm"
	"github.com/jmoiron/sqlx"

	"github.com/code-payments/code-vm-indexer/data/ram"
)

type store struct {
	db        *sqlx.DB
	tableName string
}

// New returns a new postgres-backed ram.Store
func New(db *sql.DB, tableName string) ram.Store {
	return &store{
		db:        sqlx.NewDb(db, "pgx"),
		tableName: tableName,
	}
}

// Save implements ram.Store.Save
func (s *store) Save(ctx context.Context, record *ram.Record) error {
	model, err := toModel(record)
	if err != nil {
		return err
	}

	err = model.dbSave(ctx, s.tableName, s.db)
	if err != nil {
		return err
	}

	fromModel(model).CopyTo(record)

	return nil
}

// GetAllMemoryAccounts implements ram.Store.GetAllMemoryAccounts
func (s *store) GetAllMemoryAccounts(ctx context.Context, vm string) ([]string, error) {
	return dbGetAllMemoryAccounts(ctx, s.tableName, s.db, vm)
}

// GetAllByMemoryAccount implements ram.Store.GetAllByMemoryAccount
func (s *store) GetAllByMemoryAccount(ctx context.Context, memoryAccount string) ([]*ram.Record, error) {
	models, err := dbGetAllByMemoryAccount(ctx, s.tableName, s.db, memoryAccount)
	if err != nil {
		return nil, err
	}

	res := make([]*ram.Record, len(models))
	for i, model := range models {
		res[i] = fromModel(model)
	}
	return res, nil
}

// GetAllVirtualAccountsByAddressAndType implements ram.Store.GetAllVirtualAccountsByAddressAndType
func (s *store) GetAllVirtualAccountsByAddressAndType(ctx context.Context, address string, accountType cvm.VirtualAccountType) ([]*ram.Record, error) {
	models, err := dbGetAllVirtualAccountsByAddressAndType(ctx, s.tableName, s.db, address, accountType)
	if err != nil {
		return nil, err
	}

	res := make([]*ram.Record, len(models))
	for i, model := range models {
		res[i] = fromModel(model)
	}
	return res, nil
}
