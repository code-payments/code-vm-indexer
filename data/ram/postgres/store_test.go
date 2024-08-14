package postgres

import (
	"database/sql"
	"os"
	"testing"

	postgrestest "github.com/code-payments/code-server/pkg/database/postgres/test"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/code-payments/code-vm-indexer/data/ram"
	"github.com/code-payments/code-vm-indexer/data/ram/tests"
)

var (
	testStore ram.Store
	teardown  func()
)

const (
	// Used for testing ONLY, the table and migrations are external to this repository
	tableCreate = `
	CREATE TABLE vm_ram_test_table (
		id SERIAL NOT NULL PRIMARY KEY,

		vm TEXT NOT NULL,

		memory_account TEXT NOT NULL,
		index INTEGER NOT NULL,
		is_allocated BOOL NOT NULL,

		address TEXT NULL,
		item_type SMALLINT NULL,
		data BYTEA NULL,

		slot BIGINT NOT NULL,

		last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

		CONSTRAINT vm_ram_test_table__uniq__memory_account__and__index UNIQUE (memory_account, index)
	);
	`

	// Used for testing ONLY, the table and migrations are external to this repository
	tableDestroy = `
		DROP TABLE vm_ram_test_table;
	`
)

func TestMain(m *testing.M) {
	log := logrus.StandardLogger()

	testPool, err := dockertest.NewPool("")
	if err != nil {
		log.WithError(err).Error("Error creating docker pool")
		os.Exit(1)
	}

	var cleanUpFunc func()
	db, cleanUpFunc, err := postgrestest.StartPostgresDB(testPool)
	if err != nil {
		log.WithError(err).Error("Error starting postgres image")
		os.Exit(1)
	}
	defer db.Close()

	if err := createTestTables(db); err != nil {
		logrus.StandardLogger().WithError(err).Error("Error creating test tables")
		cleanUpFunc()
		os.Exit(1)
	}

	testStore = New(db, "vm_ram_test_table")
	teardown = func() {
		if pc := recover(); pc != nil {
			cleanUpFunc()
			panic(pc)
		}

		if err := resetTestTables(db); err != nil {
			logrus.StandardLogger().WithError(err).Error("Error resetting test tables")
			cleanUpFunc()
			os.Exit(1)
		}
	}

	code := m.Run()
	cleanUpFunc()
	os.Exit(code)
}

func TestRamPostgresStore(t *testing.T) {
	tests.RunTests(t, testStore, teardown)
}

func createTestTables(db *sql.DB) error {
	_, err := db.Exec(tableCreate)
	if err != nil {
		logrus.StandardLogger().WithError(err).Error("could not create test tables")
		return err
	}
	return nil
}

func resetTestTables(db *sql.DB) error {
	_, err := db.Exec(tableDestroy)
	if err != nil {
		logrus.StandardLogger().WithError(err).Error("could not drop test tables")
		return err
	}

	return createTestTables(db)
}
