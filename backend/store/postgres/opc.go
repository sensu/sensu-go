package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sensu/sensu-go/backend/store"
)

type OPC struct {
	db DBI
}

func NewOPC(db DBI) *OPC {
	return &OPC{db: db}
}

func (o *OPC) MonitorOperators(ctx context.Context, req store.MonitorOperatorsRequest) <-chan []store.OperatorState {
	results := make(chan []store.OperatorState, 1)
	if req.Every == 0 {
		req.Every = time.Second
	}
	if req.ErrorHandler == nil {
		req.ErrorHandler = func(error) {}
	}
	go func() {
		ticker := time.NewTicker(req.Every)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				o.doMonitorTx(ctx, req, results)
			}
		}
	}()
	return results
}

func (o *OPC) reassignAbsentControllers(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, opcReassignAbsentControllers)
	if err != nil {
		return err
	}
	return nil
}

func (o *OPC) doMonitorTx(ctx context.Context, req store.MonitorOperatorsRequest, results chan []store.OperatorState) {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		req.ErrorHandler(err)
		return
	}
	var rollback bool
	defer func() {
		if rollback {
			if err := tx.Rollback(ctx); err != nil {
				req.ErrorHandler(err)
			}
		} else {
			if err := tx.Commit(ctx); err != nil {
				req.ErrorHandler(err)
			}
		}
	}()

	if err := o.reassignAbsentControllers(ctx, tx); err != nil {
		req.ErrorHandler(fmt.Errorf("couldn't reassign deleted controllers: %s", err))
		rollback = true
		return
	}

	operators, controllerIDs, err := o.getMonitorResults(ctx, tx, req)
	if err != nil {
		rollback = true
		req.ErrorHandler(err)
		return
	}

	if len(operators) == 0 {
		rollback = true
		results <- operators
		return
	}

	ctlIDs := make([]sql.NullInt64, 0, len(controllerIDs))
	for k := range controllerIDs {
		ctlIDs = append(ctlIDs, k)
	}

	if err := o.updateMonitorResults(ctx, tx, req, ctlIDs); err != nil {
		rollback = true
		req.ErrorHandler(err)
		return
	}

	results <- operators
}

type postgresOpState store.OperatorState

func (p *postgresOpState) SQLParams(ctlID *sql.NullInt64) []interface{} {
	return []interface{}{
		&p.Namespace,
		&p.Type,
		&p.Name,
		&p.Present,
		&p.LastUpdate,
		&p.CheckInTimeout,
		&p.Metadata,
		ctlID,
	}
}

type controllerGetter struct {
	tx          pgx.Tx
	controllers map[int64]*store.OperatorKey
}

func newControllerGetter(tx pgx.Tx) *controllerGetter {
	return &controllerGetter{
		tx:          tx,
		controllers: make(map[int64]*store.OperatorKey),
	}
}

func (c *controllerGetter) GetController(ctx context.Context, id sql.NullInt64) (*store.OperatorKey, error) {
	if !id.Valid {
		return nil, nil
	}
	key, ok := c.controllers[id.Int64]
	if ok {
		return key, nil
	}
	row := c.tx.QueryRow(ctx, opcGetOperatorByID, id.Int64)
	var controller postgresOpState
	if err := row.Scan(controller.SQLParams(new(sql.NullInt64))...); err != nil {
		return nil, err
	}
	result := &store.OperatorKey{
		Namespace: controller.Namespace,
		Type:      controller.Type,
		Name:      controller.Name,
	}
	c.controllers[id.Int64] = result
	return result, nil
}

func (o *OPC) getMonitorResults(ctx context.Context, tx pgx.Tx, req store.MonitorOperatorsRequest) ([]store.OperatorState, map[sql.NullInt64]struct{}, error) {
	results := make([]store.OperatorState, 0)
	rows, err := tx.Query(ctx, opcGetNotifications, req.Type, req.ControllerName, req.ControllerType, req.ControllerNamespace)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer rows.Close()
	controllerGetter := newControllerGetter(tx)
	controllerIDs := []sql.NullInt64{}
	for rows.Next() {
		var controllerID sql.NullInt64
		var result store.OperatorState
		ptr := (*postgresOpState)(&result)
		if err := rows.Scan(ptr.SQLParams(&controllerID)...); err != nil {
			return nil, nil, err
		}
		results = append(results, result)
		controllerIDs = append(controllerIDs, controllerID)
	}
	ctlIDs := make(map[sql.NullInt64]struct{})
	for i := range results {
		ctl, err := controllerGetter.GetController(ctx, controllerIDs[i])
		if err != nil {
			return nil, nil, err
		}
		results[i].Controller = ctl
		ctlIDs[controllerIDs[i]] = struct{}{}
	}

	return results, ctlIDs, nil
}

func (o *OPC) updateMonitorResults(ctx context.Context, tx pgx.Tx, req store.MonitorOperatorsRequest, ids []sql.NullInt64) error {
	int64IDs := make([]int64, len(ids))
	for i := range ids {
		int64IDs[i] = ids[i].Int64
	}
	_, err := tx.Exec(ctx, opcUpdateNotifications, int64IDs, req.Type)
	return err
}

func (o *OPC) CheckIn(ctx context.Context, state store.OperatorState) error {
	ctl := state.Controller
	var ctlNamespace, ctlType, ctlName interface{}
	if ctl != nil {
		ctlNamespace = ctl.Namespace
		ctlType = ctl.Type
		ctlName = ctl.Name
	}
	var meta = []byte("{}")
	if state.Metadata != nil {
		meta = *state.Metadata
	}
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return err
	}
	var rollback bool
	defer func() {
		if rollback {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()
	row := tx.QueryRow(ctx, getOperatorID, state.Namespace, state.Type, state.Name)
	var id int64
	timeout := int64(state.CheckInTimeout / time.Microsecond)
	if err := row.Scan(&id); err != nil {
		if err != pgx.ErrNoRows {
			rollback = true
			return err
		}
		// insert case
		result, err := tx.Exec(ctx, opcCheckInInsert, state.Namespace, state.Type,
			state.Name, timeout, state.Present, meta, ctlNamespace, ctlType, ctlName)
		if err != nil {
			rollback = true
			return fmt.Errorf("couldn't insert operator record: %s", err)
		}
		if result.RowsAffected() != 1 {
			return fmt.Errorf("%s: wrong number of rows inserted: want 1, got %d", state.Name, result.RowsAffected())
		}
		return nil
	}
	// update case
	_, err = tx.Exec(ctx, opcCheckInUpdate, id, timeout, state.Present, state.Metadata, ctlNamespace, ctlType, ctlName)
	if err != nil {
		return fmt.Errorf("couldn't update operator record: %s", err)
	}
	return nil
}

func (o *OPC) CheckOut(ctx context.Context, key store.OperatorKey) error {
	result, err := o.db.Exec(ctx, opcCheckOut, key.Namespace, key.Type, key.Name)
	if err != nil {
		return err
	}
	if result.RowsAffected() < 1 {
		return &store.ErrNotFound{Key: fmt.Sprintf("%#v", key)}
	}
	return nil
}

func (o *OPC) QueryOperator(ctx context.Context, key store.OperatorKey) (store.OperatorState, error) {
	row := o.db.QueryRow(ctx, opcGetOperator, key.Namespace, key.Type, key.Name)
	var controllerID sql.NullInt64
	var operator postgresOpState
	if err := row.Scan(operator.SQLParams(&controllerID)...); err != nil {
		if err == pgx.ErrNoRows {
			return store.OperatorState{}, &store.ErrNotFound{Key: fmt.Sprintf("%#v\n", key)}
		}
		return store.OperatorState{}, err
	}
	if controllerID.Valid {
		row := o.db.QueryRow(ctx, opcGetOperatorByID, controllerID)
		var controller postgresOpState
		if err := row.Scan(controller.SQLParams(new(sql.NullInt64))...); err != nil {
			return store.OperatorState{}, err
		}
		operator.Controller = &store.OperatorKey{
			Namespace: controller.Namespace,
			Type:      controller.Type,
			Name:      controller.Name,
		}
	}
	return store.OperatorState(operator), nil
}
