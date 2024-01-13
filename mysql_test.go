package queryplan

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_queueQueryToSend(t *testing.T) {
	type stmtAndDuration struct {
		stmt *queryPlanStmt
		dur  time.Duration
	}

	tests := []struct {
		name              string
		ringBufferSize    int
		stmtsAndDurations []stmtAndDuration
		want              []QueryPlanQuery
	}{
		{
			name:           "exceeding the ring buffer size",
			ringBufferSize: 2,
			stmtsAndDurations: []stmtAndDuration{
				{
					stmt: &queryPlanStmt{
						query: "SELECT * FROM table1 WHERE id = ?",
					},
					dur: 1 * time.Second,
				},
				{
					stmt: &queryPlanStmt{
						query: "SELECT * FROM table2 WHERE id = ?",
					},
					dur: 1 * time.Second,
				},
				{
					stmt: &queryPlanStmt{
						query: "SELECT * FROM table3 WHERE id = ?",
					},
					dur: 1 * time.Second,
				},
			},
			want: []QueryPlanQuery{
				{
					ExecutedAt:      0,
					Query:           "SELECT * FROM table2 WHERE id = ?",
					CallStack:       nil,
					ParsedCallStack: nil,
					Duration:        1,
				},
				{
					ExecutedAt:      0,
					Query:           "SELECT * FROM table3 WHERE id = ?",
					CallStack:       nil,
					ParsedCallStack: nil,
					Duration:        1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitQueryPlan(QueryPlanOpts{
				MaxPendingQueriesSize: tt.ringBufferSize,
			})

			for _, stmtAndDur := range tt.stmtsAndDurations {
				err := queueQueryToSend(stmtAndDur.stmt, stmtAndDur.dur)
				assert.NoError(t, err)
			}

			// only compare the queries, not the timestamps
			gotQueries := []string{}
			for _, query := range pendingQueries.GetAll() {
				gotQueries = append(gotQueries, query.Query)
			}

			wantQueries := []string{}
			for _, query := range tt.want {
				wantQueries = append(wantQueries, query.Query)
			}

			assert.Equal(t, wantQueries, gotQueries)
		})
	}
}
