package postgres

import (
	"context"

	"github.com/echlebek/migration"
)

// migrations are a log of database migrations for the config store.
// They are applied by the migration library. From the docs:
//
// Once a migration is created, it should never be changed. Every time a
// database is opened with this package, all necessary migrations are executed
// in a single transaction. If any part of the process fails, an error is
// returned and the transaction is rolled back so that the database is left
// untouched.
//
// Add new migrations by adding to the migrations slice. Do not disturb the
// ordering of existing migrations!
var configMigrations = []migration.Migrator{
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), ConfigurationDDL)
		return err
	},
}
