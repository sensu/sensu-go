// +build integration,!race

package leader

import (
	"context"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDo(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)

	require.NoError(t, Initialize(client))

	workPerformed := false
	f := func(context.Context) error {
		workPerformed = true
		return nil
	}

	require.NoError(t, Do(f))
	require.True(t, workPerformed)

	// Test that work is cancelled after resigning
	var funcWg, workWg sync.WaitGroup
	funcWg.Add(1)
	workWg.Add(1)

	f = func(ctx context.Context) error {
		workWg.Done()
		<-ctx.Done()
		funcWg.Done()
		return nil
	}

	go func() {
		Do(f)
	}()

	workWg.Wait()
	require.NoError(t, Resign())

	funcWg.Wait()
}

func TestLeaderContention(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	sensuLeaderKey = "/somethingelse/"
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)

	ssn1, err := concurrency.NewSession(client)
	require.NoError(t, err)

	ssn2, err := concurrency.NewSession(client)
	require.NoError(t, err)

	s1 := newSupervisor(ssn1)
	s1.Start()
	s1.WaitLeader()
	s2 := newSupervisor(ssn2)
	s2.Start()

	assert.True(t, s1.IsLeader())
	assert.False(t, s2.IsLeader())

	workPerformed := false
	f := func(context.Context) error {
		workPerformed = true
		return nil
	}
	g := func(context.Context) error {
		return nil
	}
	w1 := newWork(f)
	w2 := newWork(g)

	s1.Exec(w1)
	go s2.Exec(w2)

	require.NoError(t, w1.Err())
	require.True(t, workPerformed)

	require.NoError(t, s1.Stop())

	// Wait for the other session to become leader.
	s2.WaitLeader()
	assert.True(t, s2.IsLeader())

	require.NoError(t, w2.Err())

	s1.Start()
	require.NoError(t, s2.Stop())

	s1.WaitLeader()
	assert.True(t, s1.IsLeader())

	h := func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	w3 := newWork(h)

	s1.Exec(w3)

	// simulate a loss of leadership during ongoing work
	s1.isFollower <- struct{}{}

	assert.Error(t, w3.Err())
	require.NoError(t, s1.Stop())
}
