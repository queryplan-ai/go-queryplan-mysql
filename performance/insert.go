package main

import (
	"database/sql"
	"time"

	"github.com/goombaio/namegenerator"
)

func insertTest(db *sql.DB, tableName string) (*InsertResult, error) {
	insertCount := 10000

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	names := []string{}

	// generate insertCount random names
	for i := 0; i < insertCount; i++ {
		names = append(names, nameGenerator.Generate())
	}

	// insert insertCount rows
	durations := []time.Duration{}
	start := time.Now()
	for _, name := range names {
		innerStart := time.Now()
		query := "INSERT INTO " + tableName + " (name) VALUES (?)"
		if _, err := db.Exec(query, name); err != nil {
			return nil, err
		}
		innerDur := time.Since(innerStart)
		durations = append(durations, innerDur)
	}
	dur := time.Since(start)

	averageDuration := time.Duration(0)
	for _, d := range durations {
		averageDuration += d
	}

	averageDuration = time.Duration(0)
	averageDuration = averageDuration / time.Duration(len(durations))

	return &InsertResult{
		TotalDuration:   dur,
		AverageDuration: averageDuration,
	}, nil
}
