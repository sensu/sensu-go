package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v4/pgxpool"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
)

func TestOPCControllerMigrationTimedOut(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		//ctx, cancel := context.WithCancel(ctx)
		//defer cancel()
		nsStore := NewNamespaceStore(db)
		ns := corev3.FixtureNamespace("default")
		if err := nsStore.CreateIfNotExists(ctx, ns); err != nil {
			t.Fatal(err)
		}
		backendState1 := store.OperatorState{
			Type:           store.BackendOperator,
			Name:           "backend1",
			CheckInTimeout: time.Millisecond,
			Present:        true,
		}
		backendState2 := store.OperatorState{
			Type:           store.BackendOperator,
			Name:           "backend2",
			CheckInTimeout: time.Hour,
			Present:        true,
		}
		agentState := store.OperatorState{
			Namespace: "default",
			Type:      store.AgentOperator,
			Name:      "agent1",
			Controller: &store.OperatorKey{
				Namespace: "",
				Type:      store.BackendOperator,
				Name:      "backend1",
			},
			CheckInTimeout: 100 * time.Millisecond,
			Present:        true,
		}
		opc := NewOPC(db)
		if err := opc.CheckIn(ctx, backendState1); err != nil {
			t.Fatal(err)
		}
		if err := opc.CheckIn(ctx, backendState2); err != nil {
			t.Fatal(err)
		}
		if err := opc.CheckIn(ctx, agentState); err != nil {
			t.Fatal(err)
		}

		req := store.MonitorOperatorsRequest{
			Type:           store.AgentOperator,
			ControllerType: store.BackendOperator,
			ControllerName: "backend1",
			Every:          100 * time.Millisecond,
			ErrorHandler: func(err error) {
				// can get closed pool error here due to a race between the test
				// terminating and the pool closing, so don't handle errors
			},
		}

		notifications := opc.MonitorOperators(ctx, req)

		bop, err := opc.QueryOperator(ctx, store.OperatorKey{Namespace: "", Type: store.BackendOperator, Name: "backend1"})
		if err != nil {
			t.Error(err)
		}
		if !bop.Present {
			t.Error("backend1 not present")
		}

		// drain the buffer in case it filled before backend1 became absent
		_ = <-notifications

		if notes := <-notifications; len(notes) > 0 {
			t.Error("shouldn't have got a notification")
		}

		req = store.MonitorOperatorsRequest{
			Type:           store.AgentOperator,
			ControllerType: store.BackendOperator,
			ControllerName: "backend2",
			Every:          100 * time.Millisecond,
			ErrorHandler: func(err error) {
				// can get closed pool error here due to a race between the test
				// terminating and the pool closing, so don't handle errors
			},
		}

		notifications = opc.MonitorOperators(ctx, req)
		opStates := <-notifications

		if len(opStates) == 0 {
			t.Error("no opstates")
			return
		}

		if opStates[0].Controller == nil {
			t.Error("nil controller")
			return
		}

		if got, want := opStates[0].Controller.Name, "backend2"; got != want {
			t.Errorf("bad controller: got %s, want %s", got, want)
		}

	})
}

func TestOPCControllerMigrationCheckOut(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		nsStore := NewNamespaceStore(db)
		ns := corev3.FixtureNamespace("default")
		if err := nsStore.CreateIfNotExists(ctx, ns); err != nil {
			t.Fatal(err)
		}
		backendState1 := store.OperatorState{
			Type:           store.BackendOperator,
			Name:           "backend1",
			CheckInTimeout: time.Hour,
			Present:        true,
		}
		backendState2 := store.OperatorState{
			Type:           store.BackendOperator,
			Name:           "backend2",
			CheckInTimeout: time.Hour,
			Present:        true,
		}
		agentState := store.OperatorState{
			Namespace: "default",
			Type:      store.AgentOperator,
			Name:      "agent1",
			Controller: &store.OperatorKey{
				Namespace: "",
				Type:      store.BackendOperator,
				Name:      "backend1",
			},
			CheckInTimeout: 100 * time.Millisecond,
			Present:        true,
		}
		opc := NewOPC(db)
		if err := opc.CheckIn(ctx, backendState1); err != nil {
			t.Fatal(err)
		}
		if err := opc.CheckIn(ctx, backendState2); err != nil {
			t.Fatal(err)
		}
		if err := opc.CheckIn(ctx, agentState); err != nil {
			t.Fatal(err)
		}

		req := store.MonitorOperatorsRequest{
			Type:           store.AgentOperator,
			ControllerType: store.BackendOperator,
			ControllerName: "backend1",
			Every:          100 * time.Millisecond,
			ErrorHandler: func(err error) {
				// can get closed pool error here due to a race between the test
				// terminating and the pool closing, so don't handle errors
			},
		}

		notifications := opc.MonitorOperators(ctx, req)
		opStates := <-notifications
		if len(opStates) == 0 {
			for {
				opStates = <-notifications
				if len(opStates) != 0 {
					break
				}
			}
		}

		if opStates[0].Present {
			t.Error("agent1 present")
		}

		if err := opc.CheckOut(ctx, store.OperatorKey{Namespace: "", Type: store.BackendOperator, Name: "backend1"}); err != nil {
			t.Error(err)
			return
		}

		if _, err := opc.QueryOperator(ctx, store.OperatorKey{Namespace: "", Type: store.BackendOperator, Name: "backend1"}); err == nil {
			t.Error("expected non-nil error")
		} else if _, ok := err.(*store.ErrNotFound); !ok {
			t.Error(err)
		}

		// drain the buffer in case it filled before checkout
		_ = <-notifications

		if len(<-notifications) > 0 {
			t.Error("shouldn't have got a notification")
		}

		req = store.MonitorOperatorsRequest{
			Type:           store.AgentOperator,
			ControllerType: store.BackendOperator,
			ControllerName: "backend2",
			Every:          100 * time.Millisecond,
			ErrorHandler: func(err error) {
				// can get closed pool error here due to a race between the test
				// terminating and the pool closing, so don't handle errors
			},
		}

		notifications = opc.MonitorOperators(ctx, req)

		opStates = <-notifications

		if len(opStates) == 0 {
			t.Error("no opstates")
			return
		}

		if opStates[0].Controller == nil {
			t.Error("nil controller")
			return
		}

		if got, want := opStates[0].Controller.Name, "backend2"; got != want {
			t.Errorf("bad controller: got %s, want %s", got, want)
		}

	})
}

func TestOPCQueryOperator(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		nsStore := NewNamespaceStore(db)
		ns := corev3.FixtureNamespace("default")
		if err := nsStore.CreateIfNotExists(ctx, ns); err != nil {
			t.Fatal(err)
		}
		backendState := store.OperatorState{
			Type:           store.BackendOperator,
			Name:           "backend1",
			CheckInTimeout: time.Hour,
			Present:        true,
		}
		agentState := store.OperatorState{
			Namespace: "default",
			Type:      store.AgentOperator,
			Name:      "agent1",
			Controller: &store.OperatorKey{
				Namespace: "",
				Type:      store.BackendOperator,
				Name:      "backend1",
			},
			CheckInTimeout: time.Millisecond * 100,
			Present:        true,
		}
		opc := NewOPC(db)
		if err := opc.CheckIn(ctx, backendState); err != nil {
			t.Fatal(err)
		}
		if err := opc.CheckIn(ctx, agentState); err != nil {
			t.Fatal(err)
		}

		agentOp, err := opc.QueryOperator(ctx, store.OperatorKey{Namespace: "default", Type: store.AgentOperator, Name: "agent1"})
		if err != nil {
			t.Fatal(err)
		}
		got, want := agentOp, agentState
		if !cmp.Equal(got.Namespace, want.Namespace) {
			t.Error("unequal namespace", cmp.Diff(got.Namespace, want.Namespace))
		}
		if !cmp.Equal(got.Name, want.Name) {
			t.Error("unequal name", cmp.Diff(got.Name, want.Name))
		}
		if !cmp.Equal(got.Type, want.Type) {
			t.Error("unequal type", cmp.Diff(got.Type, want.Type))
		}
		if !cmp.Equal(got.CheckInTimeout, want.CheckInTimeout) {
			t.Error("unequal check in timeout", got.CheckInTimeout, want.CheckInTimeout)
		}
		if !got.Present {
			t.Error("operator is not present")
		}
		if !cmp.Equal(got.Controller, want.Controller) {
			t.Error("unequal controller", cmp.Diff(got.Controller, want.Controller))
		}
	})
}

func TestOPCQueryControllerOperator(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		nsStore := NewNamespaceStore(db)
		ns := corev3.FixtureNamespace("default")
		if err := nsStore.CreateIfNotExists(ctx, ns); err != nil {
			t.Fatal(err)
		}
		backendState := store.OperatorState{
			Type:           store.BackendOperator,
			Name:           "backend1",
			CheckInTimeout: time.Hour,
			Present:        true,
		}
		opc := NewOPC(db)
		if err := opc.CheckIn(ctx, backendState); err != nil {
			t.Fatal(err)
		}

		backendOp, err := opc.QueryOperator(ctx, store.OperatorKey{Namespace: "", Type: store.BackendOperator, Name: "backend1"})
		if err != nil {
			t.Fatal(err)
		}
		got, want := backendOp, backendState
		if !cmp.Equal(got.Namespace, want.Namespace) {
			t.Error("unequal namespace", cmp.Diff(got.Namespace, want.Namespace))
		}
		if !cmp.Equal(got.Name, want.Name) {
			t.Error("unequal name", cmp.Diff(got.Name, want.Name))
		}
		if !cmp.Equal(got.Type, want.Type) {
			t.Error("unequal type", cmp.Diff(got.Type, want.Type))
		}
		if !cmp.Equal(got.CheckInTimeout, want.CheckInTimeout) {
			t.Error("unequal check in timeout", got.CheckInTimeout, want.CheckInTimeout)
		}
		if !cmp.Equal(got.Controller, want.Controller) {
			t.Error("unequal controller", cmp.Diff(got.Controller, want.Controller))
		}
		if !got.Present {
			t.Error("operator not present")
		}
	})
}

func TestOPCIntegration(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		nsStore := NewNamespaceStore(db)
		ns := corev3.FixtureNamespace("default")
		if err := nsStore.CreateIfNotExists(ctx, ns); err != nil {
			t.Fatal(err)
		}
		backendState := store.OperatorState{
			Type:           store.BackendOperator,
			Name:           "backend1",
			CheckInTimeout: time.Hour,
			Present:        true,
		}
		agentState := store.OperatorState{
			Namespace: "default",
			Type:      store.AgentOperator,
			Name:      "agent1",
			Controller: &store.OperatorKey{
				Namespace: "",
				Type:      store.BackendOperator,
				Name:      "backend1",
			},
			CheckInTimeout: time.Millisecond * 100,
			Present:        true,
		}
		opc := NewOPC(db)
		if err := opc.CheckIn(ctx, backendState); err != nil {
			t.Fatal(err)
		}
		if err := opc.CheckIn(ctx, agentState); err != nil {
			t.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)

		req := store.MonitorOperatorsRequest{
			Type:           store.AgentOperator,
			ControllerType: store.BackendOperator,
			ControllerName: "backend1",
			Every:          100 * time.Millisecond,
			ErrorHandler: func(err error) {
				// can get closed pool error here due to a race between the test
				// terminating and the pool closing, so don't handle errors
			},
		}
		notifications := opc.MonitorOperators(ctx, req)

		for i := 0; i < 2; i++ {
			// run the loop twice to show that we get repeated notifications
			opStates := <-notifications

			if len(opStates) == 0 {
				// this can happen when the polling interval is out of phase
				// with the check in timeout multiple
				continue
			}

			if got, want := len(opStates), 1; got != want {
				t.Errorf("got %d op states, want %d", got, want)
			}

			got, want := opStates[0], agentState
			if !cmp.Equal(got.Namespace, want.Namespace) {
				t.Error("unequal namespace", cmp.Diff(got.Namespace, want.Namespace))
			}
			if !cmp.Equal(got.Name, want.Name) {
				t.Error("unequal name", cmp.Diff(got.Name, want.Name))
			}
			if !cmp.Equal(got.Type, want.Type) {
				t.Error("unequal type", cmp.Diff(got.Type, want.Type))
			}
			if !cmp.Equal(got.CheckInTimeout, want.CheckInTimeout) {
				t.Error("unequal check in timeout", got.CheckInTimeout, want.CheckInTimeout)
			}
			if got.LastUpdate.Add(100 * time.Duration(i) * time.Millisecond).After(time.Now()) {
				t.Errorf("bad last seen: %v", got.LastUpdate)
			}
			if got.Controller == nil {
				t.Error("nil controller")
			}
			if got.Present {
				t.Error("operator is present")

				op, err := opc.QueryOperator(ctx, store.OperatorKey{Namespace: "default", Type: store.AgentOperator, Name: "agent1"})
				if err != nil {
					t.Error(err)
				}
				fmt.Println(op.Present)
			}
		}

		if err := opc.CheckOut(ctx, store.OperatorKey{Namespace: "default", Type: store.AgentOperator, Name: "agent1"}); err != nil {
			t.Error(err)
		}

		time.Sleep(100 * time.Millisecond)

		if len(<-notifications) > 0 {
			t.Error("shouldn't have got a notification")
		}

	})
}
