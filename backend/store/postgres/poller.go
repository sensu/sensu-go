package postgres

import (
	"context"
	"fmt"
	"time"

	corev2 "github.com/sensu/core/v2"
	apitools "github.com/sensu/sensu-api-tools"
	"github.com/sensu/sensu-go/backend/poll"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

type entityConfigPoller struct {
	db  DBI
	req storev2.ResourceRequest
}

func (e *entityConfigPoller) Now(ctx context.Context) (time.Time, error) {
	var now time.Time
	row := e.db.QueryRow(ctx, "SELECT NOW();")
	if err := row.Scan(&now); err != nil {
		return now, &store.ErrInternal{Message: err.Error()}
	}
	return now, nil
}

func (e *entityConfigPoller) Since(ctx context.Context, updatedSince time.Time) ([]poll.Row, error) {
	wrapper := &EntityConfigWrapper{
		Namespace: e.req.Namespace,
		Name:      e.req.Name,
		UpdatedAt: updatedSince,
	}
	queryParams := wrapper.SQLParams()
	rows, rerr := e.db.Query(ctx, pollEntityConfigQuery, queryParams...)
	if rerr != nil {
		logger.Errorf("entity config since query failed with error %v", rerr)
		return nil, &store.ErrInternal{Message: rerr.Error()}
	}
	defer rows.Close()
	var since []poll.Row
	for rows.Next() {
		var resultW EntityConfigWrapper
		if err := rows.Scan(resultW.SQLParams()...); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		if err := rows.Err(); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		id := fmt.Sprintf("%s/%s", resultW.Namespace, resultW.Name)
		pollResult := poll.Row{
			Id:        id,
			Resource:  &resultW,
			CreatedAt: resultW.CreatedAt,
			UpdatedAt: resultW.UpdatedAt,
		}
		if resultW.DeletedAt.Valid {
			pollResult.DeletedAt = &wrapper.DeletedAt.Time
		}
		since = append(since, pollResult)
	}
	return since, nil
}

func newEntityConfigPoller(req storev2.ResourceRequest, db DBI) (poll.Table, error) {
	return &entityConfigPoller{db: db, req: req}, nil
}

type configurationPoller struct {
	db  DBI
	req storev2.ResourceRequest
}

func (e *configurationPoller) Now(ctx context.Context) (time.Time, error) {
	var now time.Time
	row := e.db.QueryRow(ctx, "SELECT NOW();")
	if err := row.Scan(&now); err != nil {
		return now, &store.ErrInternal{Message: err.Error()}
	}
	return now, nil
}

func (e *configurationPoller) Since(ctx context.Context, updatedSince time.Time) ([]poll.Row, error) {
	queryParams := []interface{}{&e.req.APIVersion, &e.req.Type, &updatedSince}
	rows, err := e.db.Query(ctx, ConfigNotificationQuery, queryParams...)
	if err != nil {
		logger.Errorf("entity config since query failed with error %v", err)
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	defer rows.Close()
	var since []poll.Row
	for rows.Next() {
		var record configRecord
		if err := rows.Scan(&record.id, &record.apiVersion, &record.apiType, &record.namespace, &record.name,
			&record.resource, &record.createdAt, &record.updatedAt, &record.deletedAt); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		if err := rows.Err(); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		id := fmt.Sprint(record.id)

		resource, err := apitools.Resolve(e.req.APIVersion, e.req.Type)
		if err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		pollResult := poll.Row{
			Id: id,
			Resource: &wrap.Wrapper{
				TypeMeta:    &corev2.TypeMeta{APIVersion: e.req.APIVersion, Type: e.req.Type},
				Encoding:    wrap.Encoding_json,
				Compression: wrap.Compression_none,
				Value:       []byte(record.resource),
			},
			CreatedAt: record.createdAt.Time,
			UpdatedAt: record.updatedAt.Time,
		}
		if record.deletedAt.Valid {
			pollResult.DeletedAt = &record.deletedAt.Time
		}
		since = append(since, pollResult)
	}
	return since, nil
}

func newConfigurationPoller(req storev2.ResourceRequest, db DBI) (poll.Table, error) {
	return &configurationPoller{db: db, req: req}, nil
}
