package queryplan

import (
	"github.com/xwb1989/sqlparser"
)

func redactSQL(statement string) (string, error) {
	redactedQuery, err := sqlparser.RedactSQLQuery(statement)
	if err != nil {
		return statement, nil
	}

	return redactedQuery, nil
}
