package queryplan

import (
	"context"
	"database/sql/driver"
	"sync"
	"time"
)

type queryPlanConn struct {
	conn              driver.Conn
	mu                sync.Mutex
	activeTransaction *QueryPlanTransaction
}

func (qc *queryPlanConn) Prepare(query string) (driver.Stmt, error) {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	stmt, err := qc.conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &queryPlanStmt{
		stmt:  stmt,
		query: query,
		conn:  qc,
	}, nil
}

func (qc *queryPlanConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	if prepCtx, ok := qc.conn.(driver.ConnPrepareContext); ok {
		stmt, err := prepCtx.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &queryPlanStmt{stmt: stmt, query: query, conn: qc}, nil
	}
	return qc.Prepare(query)
}

func (qc *queryPlanConn) Close() error {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	return qc.conn.Close()
}

func (qc *queryPlanConn) Begin() (driver.Tx, error) {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	tx, err := qc.conn.Begin()
	if err != nil {
		return nil, err
	}

	txId := time.Now().UnixNano()
	stx := &QueryPlanTransaction{
		BeginAt:   txId,
		Tx:        queryPlanTx{tx: tx, id: txId},
		Conn:      qc,
		CallStack: captureCallStack(),
		Queries:   []QueryPlanQuery{},
	}

	activeTransactionsMutex.Lock()
	activeTransactions[txId] = stx
	activeTransactionsMutex.Unlock()

	qc.activeTransaction = stx // Set the active transaction

	return &queryPlanTx{tx: tx, id: txId}, nil

}
