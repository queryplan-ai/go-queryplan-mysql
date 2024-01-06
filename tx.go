package queryplan

import (
	"database/sql/driver"
	"sync"
	"time"
)

var (
	// pendingTransactions are completed transactions that need to be sent
	pendingTransactions      = []QueryPlanTransaction{}
	pendingTransactionsMutex = sync.Mutex{}

	// activeTransactions is currently open/running tx.  these will not be send
	activeTransactions      = make(map[int64]*QueryPlanTransaction)
	activeTransactionsMutex = sync.Mutex{}
)

type queryPlanTx struct {
	tx driver.Tx
	id int64
}

func (qt *queryPlanTx) Commit() error {
	activeTransactionsMutex.Lock()
	transaction, exists := activeTransactions[qt.id]
	if exists {
		delete(activeTransactions, qt.id)
		transaction.Duration = time.Since(time.Unix(0, transaction.BeginAt)).Nanoseconds()

		pendingTransactionsMutex.Lock()
		pendingTransactions = append(pendingTransactions, *transaction)
		pendingTransactionsMutex.Unlock()
	}
	activeTransactionsMutex.Unlock()

	return qt.tx.Commit()
}

func (qt *queryPlanTx) Rollback() error {
	activeTransactionsMutex.Lock()
	transaction, exists := activeTransactions[qt.id]
	if exists {
		delete(activeTransactions, qt.id)
		transaction.Duration = time.Since(time.Unix(0, transaction.BeginAt)).Nanoseconds()
		transaction.IsRollback = true

		pendingTransactionsMutex.Lock()
		pendingTransactions = append(pendingTransactions, *transaction)
		pendingTransactionsMutex.Unlock()
	}
	activeTransactionsMutex.Unlock()

	return qt.tx.Rollback()
}
