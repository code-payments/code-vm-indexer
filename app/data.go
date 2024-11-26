package app

import (
	"context"
	"fmt"

	"github.com/code-payments/code-server/pkg/config/env"
	pg "github.com/code-payments/code-server/pkg/database/postgres"
	"github.com/pkg/errors"

	"github.com/code-payments/code-vm-indexer/data/ram"
	memory_ram_store "github.com/code-payments/code-vm-indexer/data/ram/memory"
	postgres_ram_store "github.com/code-payments/code-vm-indexer/data/ram/postgres"
)

type dataStorageType string

const (
	dataStorateTypeMemory   dataStorageType = "memory"
	dataStorateTypePostgres dataStorageType = "postgres"
)

const (
	dataStorageTypeConfigEnvName = "DATA_STORAGE_TYPE"
	defaultDataStorageType       = dataStorateTypeMemory

	postgresUserConfigEnvName     = "POSTGRES_USER"
	postgresPasswordConfigEnvName = "POSTGRES_PASSWORD"
	postgresHostConfigEnvName     = "POSTGRES_HOST"
	postgresPortConfigEnvName     = "POSTGRES_PORT"
	postgresDbConfigEnvName       = "POSTGRES_DB_NAME"
	postgresMaxOpenConnsEnvName   = "POSTGRES_MAX_OPEN_CONNS"

	ramTableNameEnvName = "RAM_TABLE_NAME"
)

type DataProvider struct {
	Ram ram.Store
}

// NewDataProvider dynamically selects storage implementations for use within
// a Code VM indexer application using environment configs.
func NewDataProvider() (*DataProvider, error) {
	ctx := context.TODO()

	dataStoreTypeConfig := env.NewStringConfig(dataStorageTypeConfigEnvName, string(defaultDataStorageType))
	selectedDataStorageType := dataStorageType(dataStoreTypeConfig.Get(ctx))

	dp := &DataProvider{}

	switch selectedDataStorageType {
	case dataStorateTypeMemory:
		dp.Ram = memory_ram_store.New()
	case dataStorateTypePostgres:
		userConfig := env.NewStringConfig(postgresUserConfigEnvName, "")
		passwordConfig := env.NewStringConfig(postgresPasswordConfigEnvName, "")
		hostConfig := env.NewStringConfig(postgresHostConfigEnvName, "")
		portConfig := env.NewInt64Config(postgresPortConfigEnvName, 0)
		dbConfig := env.NewStringConfig(postgresDbConfigEnvName, "")
		maxOpenConnsConfig := env.NewUint64Config(postgresMaxOpenConnsEnvName, 64)
		ramTableNameConfig := env.NewStringConfig(ramTableNameEnvName, "")

		db, err := pg.NewWithUsernameAndPassword(
			userConfig.Get(ctx),
			passwordConfig.Get(ctx),
			hostConfig.Get(ctx),
			fmt.Sprint(portConfig.Get(ctx)),
			dbConfig.Get(ctx),
		)
		if err != nil {
			return nil, err
		}

		db.SetMaxOpenConns(int(maxOpenConnsConfig.Get(ctx)))

		dp.Ram = postgres_ram_store.New(db, ramTableNameConfig.Get(ctx))
	default:
		return nil, errors.Errorf("invalid data storage type: %s", selectedDataStorageType)
	}

	return dp, nil
}
