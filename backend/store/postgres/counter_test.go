// test resources for integration testing watcher correctness.
// defines a Counter store resource compatible with sensu-go/backend/poll
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/poll"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/types"
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
	return &v2.ObjectMeta{
		Name:        fmt.Sprint(c.Id),
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
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

func (c *counter) GetTypeMeta() v2.TypeMeta {
	return v2.TypeMeta{
		APIVersion: "counter_fixture/v2",
		Type:       "Counter",
	}
}

func counterResolver(name string) (interface{}, error) {
	switch name {
	case "Counter":
		return &counter{}, nil
	default:
		return nil, errors.New("type does not exist")
	}
}

func init() {
	types.RegisterResolver("counter_fixture/v2", counterResolver)

	registerWatchStoreOverride("testing::counter", func(req storev2.ResourceRequest, db *pgxpool.Pool) (poll.Table, error) {
		return &counterIndex{db: db}, nil
	})

}

// counterIndex implements poll.Table for counter resources
type counterIndex struct {
	db *pgxpool.Pool
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
		wr, err := wrapper.WrapResource(resource)
		if err != nil {
			return results, err
		}
		result := cr.Row(fmt.Sprint(cr.Id), wr)
		results = append(results, result)
	}
	return results, nil
}
