package queryplan

import (
	"database/sql"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	getSessionFunc func() (*sql.DB, error)

	skemticToken         = ""
	queryPlanEndpoint    = ""
	queryPlanEnvironment = ""

	pendingQueries      = []QueryPlanQuery{}
	pendingQueriesMutex = sync.Mutex{}
)

func init() {
	sql.Register("mysql-queryplan", &queryPlanDriver{sqlDriver: mysql.MySQLDriver{}})
}

type QueryPlanOpts struct {
	Token        string
	Endpoint     string
	Environment  string
	DatabaseName string
}

func InitQueryPlan(opts QueryPlanOpts, getFunc func() (*sql.DB, error)) {
	getSessionFunc = getFunc
	skemticToken = opts.Token
	queryPlanEndpoint = opts.Endpoint
	queryPlanEnvironment = opts.Environment

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
