package postgres

import (
	"context"

	"github.com/echlebek/migration"
	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/core/v2"
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
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), EventsDDL)
		return err
	},
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
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), MigrateAddSortHistoryFunc)
		return err
	},
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), MigrateAddSelectorColumn)
		return err
	},
	migrateUpdateSelectors,
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), ringSchema)
		return err
	},
	fixMissingStateSelector,
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(context.Background(), migrateEntityState)
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
