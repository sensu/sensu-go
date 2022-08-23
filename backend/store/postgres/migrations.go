package postgres

import (
	"context"

	"github.com/echlebek/migration"
	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// migrations are a log of database migrations for the event store.
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
var migrations = []migration.Migrator{
	// Migration 0
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), EventsDDL)
		return err
	},
	// Migration 1
	func(tx migration.LimitedTx) error {
		var version int64
		row := tx.QueryRow(context.Background(), `SELECT to_number(substring(Version(), 'PostgreSQL (\d+)\.'), '99');`)
		if err := row.Scan(&version); err != nil {
			goto TRYMIGRATE
		}
		if version <= 9 {
			return nil
		}
	TRYMIGRATE:
		_, err := tx.Exec(context.Background(), MigrateEventsID)
		return err
	},
	// Migration 2
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), MigrateAddSortHistoryFunc)
		return err
	},
	// Migration 3
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), MigrateAddSelectorColumn)
		return err
	},
	// Migration 4
	migrateUpdateSelectors,
	// Migration 5
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), ringSchema)
		return err
	},
	// Migration 6
	fixMissingStateSelector,
	// Migration 7
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), migrateEntityState)
		return err
	},
	// Migration 8
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), migrateDropEntitiesNetworkUnique)
		return err
	},
	// Migration 9
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), migrateRefreshUpdatedAtProcedure)
		return err
	},
	// Migration 10
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), migrateRenameEntitiesTable)
		return err
	},
	// Migration 11
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), migrateRenameEntityStateUniqueConstraint)
		return err
	},
	// Migration 12
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), entityConfigSchema)
		return err
	},
	// Migration 13
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), migrateAddEntityConfigIdToEntityState)
		return err
	},
	// Migration 14
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), namespaceSchema)
		return err
	},
	// Migration 15
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), addNamespaceForeignKeys)
		return err
	},
	// Migration 16
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), addTimestampColumns)
		return err
	},
	// Migration 17
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), addInitializedTable)
		return err
	},
	// Migration 18
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), addSilencesTable)
		return err
	},
}

type eventRecord struct {
	Namespace  string
	Entity     string
	Check      string
	Serialized []byte
}

func migrateUpdateSelectors(tx migration.LimitedTx) error {
	const limit = 1000
	offset := 0
	for {
		events, err := getEvents(tx, limit, offset)
		if err != nil {
			return err
		}
		if err := updateSelectors(tx, events); err != nil {
			return err
		}
		if len(events) < limit {
			break
		}
		offset += limit
	}
	return nil
}

func getEvents(tx migration.LimitedTx, limit, offset int) ([]eventRecord, error) {
	result := make([]eventRecord, 0, limit)
	rows, err := tx.Query(context.Background(), `SELECT sensu_namespace, sensu_entity, sensu_check, serialized FROM events ORDER BY sensu_namespace, sensu_entity, sensu_check LIMIT $1 OFFSET $2;`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var rec eventRecord
		if err := rows.Scan(&rec.Namespace, &rec.Entity, &rec.Check, &rec.Serialized); err != nil {
			return nil, err
		}
		result = append(result, rec)
	}
	return result, nil
}

func updateSelectors(tx migration.LimitedTx, events []eventRecord) error {
	for _, e := range events {
		var event corev2.Event
		if err := proto.Unmarshal(e.Serialized, &event); err != nil {
			return err
		}
		selectors := marshalSelectors(&event)
		_, err := tx.Exec(context.Background(), `UPDATE events SET selectors = $1 WHERE sensu_namespace = $2 AND sensu_entity = $3 AND sensu_check = $4`, selectors, e.Namespace, e.Entity, e.Check)
		if err != nil {
			return err
		}
	}
	return nil
}

func fixMissingStateSelector(tx migration.LimitedTx) error {
	_, err := tx.Exec(context.Background(), `UPDATE events SET selectors = jsonb_set(selectors, '{event.check.state}', '"passing"') WHERE selectors->'event.check.state' = '""' AND events.status = 0;`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), `UPDATE events SET selectors = jsonb_set(selectors, '{event.check.state}', '"failing"') WHERE selectors->'event.check.state' = '""' AND events.status > 0;`)
	if err != nil {
		return err
	}
	return nil
}
