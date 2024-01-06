package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/queryplan-ai/go-queryplan-mysql"
)

func main() {
	// the database connection string should be available in QUERYPAN_DATABASE_URL
	dbUri := os.Getenv("QUERYPLAN_DB_URI")
	if dbUri == "" {
		panic("QUERYPLAN_DB_URI is not set")
	}

	mysql, err := sql.Open("mysql", dbUri)
	if err != nil {
		panic(err)
	}
	defer mysql.Close()

	var test int
	if err := mysql.QueryRow("SELECT 1").Scan(&test); err != nil {
		panic(err)
	}

	queryplanMysql, err := sql.Open("mysql-queryplan", dbUri)
	if err != nil {
		panic(err)
	}
	defer queryplanMysql.Close()

	if err := queryplanMysql.QueryRow("SELECT 1").Scan(&test); err != nil {
		panic(err)
	}

	// create a table name to use for each tests
	currentTimestamp := fmt.Sprintf("%d", time.Now().Unix())
	goMysqlTableName := "go_mysql_" + currentTimestamp
	goQueryplanMysqlTableName := "go_queryplan_mysql_" + currentTimestamp

	// create each table
	query := "CREATE TABLE " + goMysqlTableName + " (id INT NOT NULL AUTO_INCREMENT, name VARCHAR(255), PRIMARY KEY (id))"
	if _, err := mysql.Exec(query); err != nil {
		panic(err)
	}

	query = "CREATE TABLE " + goQueryplanMysqlTableName + " (id INT NOT NULL AUTO_INCREMENT, name VARCHAR(255), PRIMARY KEY (id))"
	if _, err := mysql.Exec(query); err != nil {
		panic(err)
	}

	defer func() {
		if _, err := mysql.Exec("DROP TABLE " + goMysqlTableName); err != nil {
			panic(err)
		}

		if _, err := mysql.Exec("DROP TABLE " + goQueryplanMysqlTableName); err != nil {
			panic(err)
		}
	}()

	results := Results{}

	// run each test twice, storing only the second (why? warming up the database cache, this is hacky, but works to get a more accurate result)
	for i := 0; i < 2; i++ {
		// insert test using go-queryplan-mysql
		queryPlanInsertResults, err := insertTest(queryplanMysql, goQueryplanMysqlTableName)
		if err != nil {
			panic(err)
		}
		if i == 1 {
			results.InsertQueryPlan = *queryPlanInsertResults
		}

		// insert test using go-mysql
		nativeInsertResults, err := insertTest(mysql, goMysqlTableName)
		if err != nil {
			panic(err)
		}
		if i == 1 {
			results.InsertNative = *nativeInsertResults
		}
	}

	fmt.Println(results.String())
}
