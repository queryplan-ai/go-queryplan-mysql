package queryplan

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

var (
	DB *sql.DB
)

func GetMysqlConnection() (*sql.DB, error) {
	var connectionString string

	if mysqlConnectionData == nil {
		return nil, errors.New("mysql connection data is nil")
	}

	// first, if the connection uri is set, just use that
	if mysqlConnectionData.URI != "" {
		connectionString = mysqlConnectionData.URI
	} else {
		// otherwise, build the connection string from the other connection data
		if len(mysqlConnectionData.Host) == 0 && mysqlConnectionData.Port == 0 {
			if len(mysqlConnectionData.Pass) > 0 {
				connectionString = fmt.Sprintf("%s:%s@unix(/var/run/mysqld/mysqld.sock)/%s?parseTime=true", mysqlConnectionData.User, mysqlConnectionData.Pass, mysqlConnectionData.DatabaseName)
			} else {
				connectionString = fmt.Sprintf("%s@unix(/var/run/mysqld/mysqld.sock)/%s?parseTime=true", mysqlConnectionData.User, mysqlConnectionData.DatabaseName)
			}
		} else {
			connectionString = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", mysqlConnectionData.User, mysqlConnectionData.Pass, mysqlConnectionData.Host, mysqlConnectionData.Port, mysqlConnectionData.DatabaseName)
		}
	}

	var db *sql.DB
	var err error
	db, err = sql.Open("mysql", connectionString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open sql")
	}

	DB = db

	return db, nil
}
