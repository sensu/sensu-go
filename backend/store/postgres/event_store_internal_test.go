package postgres

import (
	"testing"

	"github.com/lib/pq"
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetHistory(t *testing.T) {
	testCases := []struct {
		op    string
		in    []pq.Int64Array
		index int64
		out   []corev2.CheckHistory
	}{
		{
			op: "empty",
			in: []pq.Int64Array{
				pq.Int64Array{},
				pq.Int64Array{},
			},
			index: 0,
			out:   []corev2.CheckHistory{},
		},
		{
			op: "one element",
			in: []pq.Int64Array{
				pq.Int64Array{0},
				pq.Int64Array{127},
			},
			index: 1,
			out: []corev2.CheckHistory{
				corev2.CheckHistory{
					Executed: 0,
					Status:   127,
				},
			},
		},
		{
			op: "two elements",
			in: []pq.Int64Array{
				pq.Int64Array{2, 1},
				pq.Int64Array{2, 0},
			},
			index: 2,
			out: []corev2.CheckHistory{
				corev2.CheckHistory{
					Executed: 2,
					Status:   2,
				},
				corev2.CheckHistory{
					Executed: 1,
					Status:   0,
				},
			},
		},
		{
			op: "pre-sorted",
			in: []pq.Int64Array{
				pq.Int64Array{0, 1, 2},
				pq.Int64Array{1, 1, 2},
			},
			out: []corev2.CheckHistory{
				corev2.CheckHistory{
					Executed: 0,
					Status:   1,
				},
				corev2.CheckHistory{
					Executed: 1,
					Status:   1,
				},
				corev2.CheckHistory{
					Executed: 2,
					Status:   2,
				},
			},
			index: 3,
		},
		{
			op: "mid",
			in: []pq.Int64Array{
				pq.Int64Array{1, 2, 0},
				pq.Int64Array{1, 1, 2},
			},
			out: []corev2.CheckHistory{
				corev2.CheckHistory{
					Executed: 1,
					Status:   1,
				},
				corev2.CheckHistory{
					Executed: 2,
					Status:   1,
				},
				corev2.CheckHistory{
					Executed: 0,
					Status:   2,
				},
			},
			index: 3,
		},
		{
			op: "full",
			in: []pq.Int64Array{
				pq.Int64Array{21, 22, 23, 24, 25, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				pq.Int64Array{21, 22, 23, 24, 25, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			out: []corev2.CheckHistory{
				corev2.CheckHistory{
					Executed: 5,
					Status:   5,
				},
				corev2.CheckHistory{
					Executed: 6,
					Status:   6,
				},
				corev2.CheckHistory{
					Executed: 7,
					Status:   7,
				},
				corev2.CheckHistory{
					Executed: 8,
					Status:   8,
				},
				corev2.CheckHistory{
					Executed: 9,
					Status:   9,
				},
				corev2.CheckHistory{
					Executed: 10,
					Status:   10,
				},
				corev2.CheckHistory{
					Executed: 11,
					Status:   11,
				},
				corev2.CheckHistory{
					Executed: 12,
					Status:   12,
				},
				corev2.CheckHistory{
					Executed: 13,
					Status:   13,
				},
				corev2.CheckHistory{
					Executed: 14,
					Status:   14,
				},
				corev2.CheckHistory{
					Executed: 15,
					Status:   15,
				},
				corev2.CheckHistory{
					Executed: 16,
					Status:   16,
				},
				corev2.CheckHistory{
					Executed: 17,
					Status:   17,
				},
				corev2.CheckHistory{
					Executed: 18,
					Status:   18,
				},
				corev2.CheckHistory{
					Executed: 19,
					Status:   19,
				},
				corev2.CheckHistory{
					Executed: 20,
					Status:   20,
				},
				corev2.CheckHistory{
					Executed: 21,
					Status:   21,
				},
				corev2.CheckHistory{
					Executed: 22,
					Status:   22,
				},
				corev2.CheckHistory{
					Executed: 23,
					Status:   23,
				},
				corev2.CheckHistory{
					Executed: 24,
					Status:   24,
				},
				corev2.CheckHistory{
					Executed: 25,
					Status:   25,
				},
			},
			index: 6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.op, func(t *testing.T) {
			if len(tc.in[0]) != len(tc.in[1]) || len(tc.in[0]) > 21 || len(tc.out) != len(tc.in[0]) {
				t.Fatalf("bad test: %s", tc.op)
			}
			out, err := getHistory(tc.in[0], tc.in[1], tc.index)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, tc.out, out)
		})
	}
}
