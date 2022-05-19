package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/poll"
)

const (
	nowQuery   = `SELECT COALESCE (MAX(updated_at) + '1 microsecond'::interval, NOW()) FROM counters;`
	sinceQuery = `SELECT id, c, created_at, updated_at, deleted_at FROM counters WHERE updated_at >= $1;`
	counterDDL = `DROP TABLE IF EXISTS counters;

	DROP FUNCTION IF EXISTS trigger_set_updated_at();
	CREATE FUNCTION trigger_set_updated_at()
	RETURNS TRIGGER AS $$
	BEGIN
	  NEW.updated_at = NOW();
	  RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;


	CREATE TABLE counters (
		id                  bigserial primary key,
		c                   bigint NOT NULL DEFAULT 0,
		created_at          timestamptz NOT NULL DEFAULT NOW(),
		updated_at          timestamptz NOT NULL DEFAULT NOW(),
		deleted_at          timestamptz
	);

	DROP TRIGGER IF EXISTS set_updated_at ON counters;
	CREATE TRIGGER set_updated_at
	BEFORE UPDATE ON counters
	FOR EACH ROW
	EXECUTE PROCEDURE trigger_set_updated_at();`
)

// counter core/v3.Resource
type counter struct {
	Id int64
	C  int64
}

func (c *counter) GetMetadata() *v2.ObjectMeta {
	return nil
}
func (c *counter) SetMetadata(*v2.ObjectMeta) {
}
func (c *counter) RBACName() string {
	return "ctr"
}
func (c *counter) StoreName() string {
	return "ctr"
}
func (c *counter) URIPath() string {
	return "ctr"
}
func (c *counter) Validate() error {
	return nil
}

// counterIndex watcher.Table and WatchQueryBuilder
type counterIndex struct {
	db *pgxpool.Pool
}

func (p *counterIndex) queryFor(s, n string) poll.Table {
	return p
}

func (p *counterIndex) Now(ctx context.Context) (time.Time, error) {
	var ts time.Time
	row := p.db.QueryRow(ctx, nowQuery)
	err := row.Scan(&ts)
	return ts, err
}

func (p *counterIndex) Since(ctx context.Context, ts time.Time) ([]poll.Row, error) {

	type counterRecord struct {
		recordStatus
		counter
	}

	rows, err := p.db.Query(ctx, sinceQuery, ts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []poll.Row
	for rows.Next() {
		var cr counterRecord

		if err := rows.Scan(&cr.Id, &cr.C, &cr.CreatedAt, &cr.UpdatedAt, &cr.DeletedAt); err != nil {
			return results, err
		}
		// unmarshal resource
		resource := &counter{C: cr.C, Id: cr.Id}
		result := cr.Row(fmt.Sprint(cr.Id), resource)
		results = append(results, result)
	}
	return results, nil
}
