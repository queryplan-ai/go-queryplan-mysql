package queryplan

import (
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
	"tailscale.com/util/ringbuffer"
)

const defaultMaxPendingQueriesSize = 10000

var (
	getSessionFunc func() (*sql.DB, error)

	skemticToken         = ""
	queryPlanEndpoint    = ""
	queryPlanEnvironment = ""

	pendingQueries = ringbuffer.New[QueryPlanQuery](maxPendingQueriesSize)

	maxPendingQueriesSize = defaultMaxPendingQueriesSize
)

func init() {
	sql.Register("mysql-queryplan", &queryPlanDriver{sqlDriver: mysql.MySQLDriver{}})
}

type QueryPlanOpts struct {
	Token        string
	Endpoint     string
	Environment  string
	DatabaseName string

	MaxPendingQueriesSize      int
	MaxPendingTransactionsSize int
}

func InitQueryPlan(opts QueryPlanOpts, getFunc func() (*sql.DB, error)) {
	getSessionFunc = getFunc
	skemticToken = opts.Token
	queryPlanEndpoint = opts.Endpoint
	queryPlanEnvironment = opts.Environment

	maxPendingQueriesSize := defaultMaxPendingQueriesSize
	if opts.MaxPendingQueriesSize > 0 {
		maxPendingQueriesSize = opts.MaxPendingQueriesSize
	}

	maxPendingTransactionsSize := defaultMaxPendingTransactionsSize
	if opts.MaxPendingTransactionsSize > 0 {
		maxPendingTransactionsSize = opts.MaxPendingTransactionsSize
	}

	pendingQueries = ringbuffer.New[QueryPlanQuery](maxPendingQueriesSize)
	pendingTransactions = ringbuffer.New[QueryPlanTransaction](maxPendingTransactionsSize)

	go func() {
		for {
			time.Sleep(5 * time.Second)

			if err := sendQueriesToQueryPlan(); err != nil {
				LogError(err)
			}
		}
	}()

	go func() {
		for {
			if err := sendSchemaToQueryPlan(opts.DatabaseName); err != nil {
				LogError(err)
			}

			time.Sleep(45 * time.Minute)
		}
	}()
}
