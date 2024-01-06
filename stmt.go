package queryplan

import (
	"context"
	"database/sql/driver"
	"sync"
	"time"
)

type queryPlanStmt struct {
	stmt  driver.Stmt
	query string
	mu    sync.Mutex
	conn  *queryPlanConn
}

func (qs *queryPlanStmt) Close() error {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	return qs.stmt.Close()
}

func (qs *queryPlanStmt) NumInput() int {
	return qs.stmt.NumInput()
}

func (qs *queryPlanStmt) Query(args []driver.Value) (driver.Rows, error) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	start := time.Now()
	result, err := qs.stmt.Query(args)
	dur := time.Since(start)

	if err := queueQueryToSend(qs, dur); err != nil {
		LogWarnf("Failed to queue query to send to queryplan: %v", err)
	}

	return result, err
}

func (qs *queryPlanStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	if queryCtx, ok := qs.stmt.(driver.StmtQueryContext); ok {
		start := time.Now()
		result, err := queryCtx.QueryContext(ctx, args)
		dur := time.Since(start)

		if err := queueQueryToSend(qs, dur); err != nil {
			LogWarnf("Failed to queue query to send to queryplan: %v", err)
		}

		return result, err
	}
	// Convert args back to []driver.Value for the standard Query method
	var vArgs []driver.Value
	for _, arg := range args {
		vArgs = append(vArgs, arg.Value)
	}

	start := time.Now()
	result, err := qs.Query(vArgs)
	dur := time.Since(start)

	if err := queueQueryToSend(qs, dur); err != nil {
		LogWarnf("Failed to queue query to send to queryplan: %v", err)
	}

	return result, err
}

func (qs *queryPlanStmt) Exec(args []driver.Value) (driver.Result, error) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	start := time.Now()
	result, err := qs.stmt.Exec(args)
	dur := time.Since(start)

	if err := queueQueryToSend(qs, dur); err != nil {
		LogWarnf("Failed to queue query to send to queryplan: %v", err)
	}

	return result, err
}

func (qs *queryPlanStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	if execCtx, ok := qs.stmt.(driver.StmtExecContext); ok {
		start := time.Now()
		result, err := execCtx.ExecContext(ctx, args)
		dur := time.Since(start)

		if err := queueQueryToSend(qs, dur); err != nil {
			LogWarnf("Failed to queue query to send to queryplan: %v", err)
		}

		return result, err
	}
	// Convert args back to []driver.Value for the standard Exec method
	var vArgs []driver.Value
	for _, arg := range args {
		vArgs = append(vArgs, arg.Value)
	}

	start := time.Now()
	result, err := qs.Exec(vArgs)
	dur := time.Since(start)

	if err := queueQueryToSend(qs, dur); err != nil {
		LogWarnf("Failed to queue query to send to queryplan: %v", err)
	}

	return result, err
}
