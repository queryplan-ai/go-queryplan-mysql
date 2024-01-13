package queryplan

import (
	"database/sql/driver"
	"time"
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

		pendingTransactions.Add(*transaction)
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

		pendingTransactions.Add(*transaction)
	}
	activeTransactionsMutex.Unlock()

	return qt.tx.Rollback()
}
