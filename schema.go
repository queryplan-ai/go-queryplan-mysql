package queryplan

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type QueryPlanTablesPayload struct {
	Environment string       `json:"environment"`
	Tables      []MysqlTable `json:"tables"`
}

type MysqlTable struct {
	TableName         string        `json:"table_name"`
	Columns           []MysqlColumn `json:"columns"`
	PrimaryKeys       []string      `json:"primary_keys"`
	EstimatedRowCount int64         `json:"estimated_row_count"`
}

type MysqlColumn struct {
	ColumnName    string  `json:"column_name"`
	DataType      string  `json:"data_type"`
	ColumnType    string  `json:"column_type"`
	IsNullable    bool    `json:"is_nullable"`
	ColumnKey     string  `json:"column_key"`
	ColumnDefault *string `json:"column_default,omitempty"`
	Extra         string  `json:"extra"`
}

const dbName = ""

func sendSchemaToQueryPlan() error {
	tables, err := listTables()
	if err != nil {
		return errors.Wrap(err, "failed to list tables")
	}

	primaryKeys, err := listPrimaryKeys()
	if err != nil {
		return errors.Wrap(err, "failed to list primary keys")
	}

	for i, table := range tables {
		if _, ok := primaryKeys[table.TableName]; !ok {
			primaryKeys[table.TableName] = []string{}
		}

		tables[i].PrimaryKeys = primaryKeys[table.TableName]
	}

	payload := QueryPlanTablesPayload{
		Environment: queryPlanEnvironment,
		Tables:      tables,
	}

	marshaled, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal table schema")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/schema", queryPlanEndpoint), bytes.NewBuffer(marshaled))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", skemticToken))

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func listPrimaryKeys() (map[string][]string, error) {
	db, err := getSessionFunc()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get mysql session")
	}

	rows, err := db.Query("SELECT TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME FROM  INFORMATION_SCHEMA.KEY_COLUMN_USAGE  WHERE  CONSTRAINT_NAME = 'PRIMARY' AND TABLE_SCHEMA = ? ORDER BY TABLE_NAME, ORDINAL_POSITION", dbName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query schema")
	}

	defer rows.Close()

	primaryKeys := map[string][]string{}
	for rows.Next() {
		tableName := ""
		columnName := ""
		if err := rows.Scan(&tableName, &tableName, &columnName); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		if _, ok := primaryKeys[tableName]; !ok {
			primaryKeys[tableName] = []string{}
		}

		primaryKeys[tableName] = append(primaryKeys[tableName], columnName)
	}

	return primaryKeys, nil
}

func listTables() ([]MysqlTable, error) {
	// read the schema from mysql
	db, err := getSessionFunc()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get mysql session")
	}

	rows, err := db.Query(`SELECT
c.TABLE_NAME, c.COLUMN_NAME, c.DATA_TYPE, c.COLUMN_TYPE, c.IS_NULLABLE, c.COLUMN_KEY, c.COLUMN_DEFAULT, c.EXTRA,
t.TABLE_ROWS
FROM INFORMATION_SCHEMA.COLUMNS c
INNER JOIN INFORMATION_SCHEMA.TABLES t ON t.TABLE_NAME = c.TABLE_NAME AND t.TABLE_SCHEMA = c.TABLE_SCHEMA
WHERE c.TABLE_SCHEMA = ?`, dbName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query schema")
	}

	defer rows.Close()

	tables := []MysqlTable{}
	for rows.Next() {
		column := MysqlColumn{}

		tableName := ""
		estimatedRowCount := int64(0)
		isNullable := ""
		columnDefault := sql.NullString{}
		if err := rows.Scan(&tableName, &column.ColumnName, &column.DataType, &column.ColumnType, &isNullable, &column.ColumnKey, &columnDefault, &column.Extra, &estimatedRowCount); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		if isNullable == "YES" {
			column.IsNullable = true
		}

		if columnDefault.Valid {
			column.ColumnDefault = &columnDefault.String
		}

		found := false
		for i, table := range tables {
			if table.TableName == tableName {
				tables[i].Columns = append(table.Columns, column)
				found = true
				continue
			}
		}

		if !found {
			tables = append(tables, MysqlTable{
				TableName:         tableName,
				Columns:           []MysqlColumn{column},
				EstimatedRowCount: estimatedRowCount,
			})
		}
	}

	return tables, nil
}
