package queryplan

import (
	"database/sql/driver"
	"sync"
	"time"

	"tailscale.com/util/ringbuffer"
)

const defaultMaxPendingTransactionsSize = 10000

var (
	maxPendingTransactionsSize = defaultMaxPendingTransactionsSize

	// pendingTransactions are completed transactions that need to be sent
	pendingTransactions = ringbuffer.New[QueryPlanTransaction](maxPendingTransactionsSize)

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
