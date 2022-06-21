package poll

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPolling(t *testing.T) {

	// series of timestamps 1s apart for testing
	ts := make([]time.Time, 100)
	ts[0] = time.Now()
	for i := 1; i < len(ts); i++ {
		ts[i] = ts[i-1].Add(time.Second)
	}

	type testInterval struct {
		RowsFromDB         []Row
		ExpectedSinceArg   time.Time
		ExpectedRowChanges int
	}
	testCases := []struct {
		Name      string
		TxnWindow time.Duration
		StartTime time.Time
		Intervals []testInterval
	}{
		{
			Name:      "No Change",
			StartTime: ts[0],
			Intervals: []testInterval{
				{ExpectedSinceArg: ts[0]},
				{ExpectedSinceArg: ts[0]},
				{ExpectedSinceArg: ts[0]},
				{ExpectedSinceArg: ts[0]},
			},
		},
		{
			Name:      "Increment respects transaction window",
			StartTime: ts[0],
			TxnWindow: time.Second * 5,
			Intervals: []testInterval{
				{
					ExpectedSinceArg:   ts[0],
					ExpectedRowChanges: 1,
					RowsFromDB:         []Row{forgeRow(ts[3])},
				}, {
					ExpectedSinceArg:   ts[0], // 5s txn window not passed
					ExpectedRowChanges: 0,     // repeat row not included in changes
					RowsFromDB:         []Row{forgeRow(ts[3])},
				}, {
					ExpectedSinceArg:   ts[0],
					ExpectedRowChanges: 1,
					RowsFromDB:         []Row{forgeRow(ts[10])},
				}, {
					ExpectedSinceArg:   ts[5], // 5 = 10s-5s txn window
					ExpectedRowChanges: 0,
					RowsFromDB:         []Row{forgeRow(ts[10])},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			table := &stubTable{
				TInit:   tc.StartTime,
				Results: make([][]Row, len(tc.Intervals)),
			}
			for i := range tc.Intervals {
				table.Results[i] = tc.Intervals[i].RowsFromDB
			}

			pollerUnderTest := &Poller{
				Interval:  time.Duration(0),
				TxnWindow: tc.TxnWindow,
				Table:     table,
			}
			pollerUnderTest.Initialize(ctx)
			for i, interval := range tc.Intervals {
				actualChanges, err := pollerUnderTest.Next(ctx)
				assert.NoError(t, err, "unexpected error on interval: ", i)
				assert.Equal(
					t,
					interval.ExpectedRowChanges,
					len(actualChanges),
					"unexpected row changes on interval ",
					i,
					actualChanges,
				)
				assert.Equal(t,
					interval.ExpectedSinceArg,
					table.sinceArg,
					"unexpected Since timestamp on interval ",
					i,
				)
			}
		})
	}
}

type stubTable struct {
	TInit    time.Time
	Results  [][]Row
	sinceArg time.Time
	calls    int
}

func (s *stubTable) Now(ctx context.Context) (time.Time, error) {
	return s.TInit, nil
}

func (s *stubTable) Since(ctx context.Context, ts time.Time) ([]Row, error) {
	defer func() { s.calls++ }()
	s.sinceArg = ts
	if len(s.Results) == 0 {
		return []Row{}, nil
	}
	calls := s.calls
	if calls >= len(s.Results) {
		calls = len(s.Results) - 1
	}
	return s.Results[calls], nil
}

func forgeRow(ts time.Time) Row {
	return Row{
		CreatedAt: ts.Add(-1 * time.Second),
		UpdatedAt: ts,
		Id:        "fake",
	}
}
