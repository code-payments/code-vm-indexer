package postgres

import (
	"context"
	"database/sql"
	"time"

	pgutil "github.com/code-payments/code-server/pkg/database/postgres"
	"github.com/code-payments/code-server/pkg/solana/cvm"
	"github.com/jmoiron/sqlx"

	"github.com/code-payments/code-vm-indexer/data/ram"
)

type model struct {
	Id sql.NullInt64 `db:"id"`

	Vm string `db:"vm"`

	MemoryAccount string `db:"memory_account"`
	Index         uint16 `db:"index"`
	IsAllocated   bool   `db:"is_allocated"`

	Address sql.NullString `db:"address"`
	Type    sql.NullInt16  `db:"item_type"`
	Data    []byte         `db:"data"`

	Slot uint64 `db:"slot"`

	LastUpdatedAt time.Time `db:"last_updated_at"`
}

func toModel(obj *ram.Record) (*model, error) {
	if err := obj.Validate(); err != nil {
		return nil, err
	}

	var address sql.NullString
	if obj.Address != nil {
		address.Valid = true
		address.String = *obj.Address
	}

	var itemType sql.NullInt16
	if obj.Type != nil {
		itemType.Valid = true
		itemType.Int16 = int16(*obj.Type)
	}

	return &model{
		Vm: obj.Vm,

		MemoryAccount: obj.MemoryAccount,
		Index:         obj.Index,
		IsAllocated:   obj.IsAllocated,

		Address: address,
		Data:    obj.Data,
		Type:    itemType,

		Slot: obj.Slot,

		LastUpdatedAt: obj.LastUpdatedAt,
	}, nil
}

func fromModel(obj *model) *ram.Record {
	var address *string
	if obj.Address.Valid {
		address = &obj.Address.String
	}

	var itemType *cvm.VirtualAccountType
	if obj.Type.Valid {
		casted := cvm.VirtualAccountType(obj.Type.Int16)
		itemType = &casted
	}

	return &ram.Record{
		Id: uint64(obj.Id.Int64),

		Vm: obj.Vm,

		MemoryAccount: obj.MemoryAccount,
		Index:         obj.Index,
		IsAllocated:   obj.IsAllocated,

		Address: address,
		Type:    itemType,
		Data:    obj.Data,

		Slot: obj.Slot,

		LastUpdatedAt: obj.LastUpdatedAt,
	}
}

func (m *model) dbSave(ctx context.Context, tableName string, db *sqlx.DB) error {
	err := pgutil.ExecuteInTx(ctx, db, sql.LevelDefault, func(tx *sqlx.Tx) error {
		query := `INSERT INTO ` + tableName + `
			(vm, memory_account, index, is_allocated, address, item_type, data, slot, last_updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)

			ON CONFLICT (memory_account, index)
			DO UPDATE
				SET is_allocated = $4, address = $5, item_type = $6, data = $7, slot = $8, last_updated_at = $9
				WHERE ` + tableName + `.memory_account = $2 AND ` + tableName + `.index = $3 AND ` + tableName + `.slot < $8

			RETURNING
				id, vm, memory_account, index, is_allocated, address, item_type, data, slot, last_updated_at`

		m.LastUpdatedAt = time.Now()

		return tx.QueryRowxContext(
			ctx,
			query,

			m.Vm,

			m.MemoryAccount,
			m.Index,
			m.IsAllocated,

			m.Address,
			m.Type,
			m.Data,

			m.Slot,

			m.LastUpdatedAt.UTC(),
		).StructScan(m)
	})
	return pgutil.CheckNoRows(err, ram.ErrStaleState)
}

func dbGetAllMemoryAccounts(ctx context.Context, tableName string, db *sqlx.DB, vm string) ([]string, error) {
	res := []string{}

	query := `SELECT DISTINCT(memory_account) FROM ` + tableName + `
		WHERE vm = $1`

	err := db.SelectContext(
		ctx,
		&res,
		query,
		vm,
	)
	if err != nil {
		return nil, pgutil.CheckNoRows(err, ram.ErrAccountNotFound)
	} else if len(res) == 0 {
		return nil, ram.ErrAccountNotFound
	}
	return res, nil
}

func dbGetAllByMemoryAccount(ctx context.Context, tableName string, db *sqlx.DB, memoryAccount string) ([]*model, error) {
	res := []*model{}

	query := `SELECT id, vm, memory_account, index, is_allocated, address, item_type, data, slot, last_updated_at FROM ` + tableName + `
		WHERE memory_account = $1`

	err := db.SelectContext(
		ctx,
		&res,
		query,
		memoryAccount,
	)
	if err != nil {
		return nil, pgutil.CheckNoRows(err, ram.ErrItemNotFound)
	} else if len(res) == 0 {
		return nil, ram.ErrItemNotFound
	}
	return res, nil
}

func dbGetAllVirtualAccountsByAddressAndType(ctx context.Context, tableName string, db *sqlx.DB, address string, accountType cvm.VirtualAccountType) ([]*model, error) {
	res := []*model{}

	query := `SELECT id, vm, memory_account, index, is_allocated, address, item_type, data, slot, last_updated_at FROM ` + tableName + `
		WHERE address = $1 AND item_type = $2`

	err := db.SelectContext(
		ctx,
		&res,
		query,
		address,
		accountType,
	)
	if err != nil {
		return nil, pgutil.CheckNoRows(err, ram.ErrItemNotFound)
	} else if len(res) == 0 {
		return nil, ram.ErrItemNotFound
	}
	return res, nil
}
