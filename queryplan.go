package queryplan

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"tailscale.com/util/ringbuffer"
)

const (
	defaultMaxPendingQueriesSize      = 10000
	defaultMaxPendingTransactionsSize = 10000
)

var (
	skemticToken         = ""
	queryPlanEndpoint    = ""
	queryPlanEnvironment = ""

	// pendingQueries are completed queries that need to be sent
	pendingQueries = ringbuffer.New[QueryPlanQuery](maxPendingQueriesSize)

	// pendingTransactions are completed transactions that need to be sent
	pendingTransactions = ringbuffer.New[QueryPlanTransaction](maxPendingTransactionsSize)

	// activeTransactions is currently open/running tx.  these will not be send
	activeTransactions      = make(map[int64]*QueryPlanTransaction)
	activeTransactionsMutex = sync.Mutex{}

	maxPendingQueriesSize      = defaultMaxPendingQueriesSize
	maxPendingTransactionsSize = defaultMaxPendingTransactionsSize

	mysqlConnectionData *MysqlConnectionData
)

type MysqlConnectionData struct {
	Host         string
	Port         int
	User         string
	Pass         string
	DatabaseName string

	URI string
}

func init() {
	sql.Register("mysql-queryplan", &queryPlanDriver{sqlDriver: mysql.MySQLDriver{}})
}

type QueryPlanOpts struct {
	Token        string
	Endpoint     string
	Environment  string
	DatabaseName string

	MysqlHost string
	MysqlPort int
	MysqlUser string
	MysqlPass string

	MysqlURI string

	MaxPendingQueriesSize      int
	MaxPendingTransactionsSize int
}

func InitQueryPlan(opts QueryPlanOpts) {
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

	// parse the mysql connection data from the opts
	mysqlConnectionData = &MysqlConnectionData{
		Host:         opts.MysqlHost,
		Port:         opts.MysqlPort,
		User:         opts.MysqlUser,
		Pass:         opts.MysqlPass,
		DatabaseName: opts.DatabaseName,

		URI: opts.MysqlURI,
	}

	go func() {
		for {
			time.Sleep(5 * time.Second)

			if err := sendQueriesToQueryPlan(); err != nil {
				LogError(err)
			}
		}
	}()

	fmt.Println("b")
	go func() {
		fmt.Println("d")
		for {
			if err := sendSchemaToQueryPlan(opts.DatabaseName); err != nil {
				LogError(err)
			}

			time.Sleep(45 * time.Minute)
		}
	}()
}
