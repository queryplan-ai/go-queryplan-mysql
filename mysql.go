package queryplan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	// Import the MySQL driver
	"github.com/pkg/errors"
)

type QueryPlanTransaction struct {
	BeginAt         int64            `json:"begin_at"`
	Duration        int64            `json:"duration"`
	IsRollback      bool             `json:"is_rollback"`
	Tx              queryPlanTx      `json:"-"`
	Conn            *queryPlanConn   `json:"-"`
	Queries         []QueryPlanQuery `json:"-"`
	CleanedQueries  []QueryPlanQuery `json:"queries"`
	CallStack       []string         `json:"-"`
	ParsedCallStack []CallStackEntry `json:"callstack"`
}

type QueryPlanQuery struct {
	ExecutedAt      int64            `json:"executed_at"`
	Duration        int64            `json:"duration"`
	Query           string           `json:"query"`
	CallStack       []string         `json:"-"`
	ParsedCallStack []CallStackEntry `json:"callstack"`
}

type QueryPlanQueryPayload struct {
	Environment  string                 `json:"environment"`
	Queries      []QueryPlanQuery       `json:"queries"`
	Transactions []QueryPlanTransaction `json:"transactions"`
}

func sendQueriesToQueryPlan() error {
	if pendingQueries.Len() == 0 {
		return nil
	}

	cleanedQueries := []QueryPlanQuery{}
	for _, query := range pendingQueries.GetAll() {
		redactedQuery, err := redactSQL(query.Query)
		if err != nil {
			fmt.Printf("failed to redact query: %q\n", query.Query)
			continue
		}

		var parsedEntries []CallStackEntry
		for _, line := range query.CallStack {
			entry, err := parseCallStackLine(line)
			if err != nil {
				logger.Warnf("Failed to parse call stack line: %v", err)
			}
			parsedEntries = append(parsedEntries, entry)
		}

		cleanedQueries = append(cleanedQueries, QueryPlanQuery{
			ExecutedAt:      query.ExecutedAt,
			Duration:        query.Duration,
			Query:           redactedQuery,
			ParsedCallStack: parsedEntries,
		})
	}

	// send pendingQueries to queryplan
	payload := QueryPlanQueryPayload{
		Environment: queryPlanEnvironment,
		Queries:     cleanedQueries,
	}

	// add transactions but parse the callstacks when we add them
	payload.Transactions = []QueryPlanTransaction{}

	for _, tx := range pendingTransactions.GetAll() {
		var parsedEntries []CallStackEntry
		for _, line := range tx.CallStack {
			entry, err := parseCallStackLine(line)
			if err != nil {
				logger.Warnf("Failed to parse call stack line: %v", err)
			}
			parsedEntries = append(parsedEntries, entry)
		}

		redactedQueries := []QueryPlanQuery{}
		for _, query := range tx.Queries {
			redactedQuery, err := redactSQL(query.Query)
			if err != nil {
				fmt.Printf("failed to redact query: %q\n", query.Query)
				continue
			}

			redactedQueries = append(redactedQueries, QueryPlanQuery{
				ExecutedAt:      query.ExecutedAt,
				Duration:        query.Duration,
				Query:           redactedQuery,
				ParsedCallStack: parsedEntries,
			})
		}

		payload.Transactions = append(payload.Transactions, QueryPlanTransaction{
			BeginAt:         tx.BeginAt,
			Duration:        tx.Duration,
			IsRollback:      tx.IsRollback,
			CleanedQueries:  redactedQueries,
			CallStack:       tx.CallStack,
			ParsedCallStack: parsedEntries,
		})
	}

	marshaled, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal pending queries")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/queries", queryPlanEndpoint), bytes.NewBuffer(marshaled))
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

	pendingQueries.Clear()
	pendingTransactions.Clear()

	return nil
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func queueQueryToSend(stmt *queryPlanStmt, dur time.Duration) error {
	activeTransactionsMutex.Lock()
	defer activeTransactionsMutex.Unlock()

	// look through active transactions to see if this query is using a connection
	// that is part of an active transaction

	for _, tx := range activeTransactions {
		if tx.Conn == stmt.conn {
			tx.Queries = append(tx.Queries, QueryPlanQuery{
				ExecutedAt: time.Now().UnixNano(),
				Query:      stmt.query,
				Duration:   dur.Nanoseconds(),
				CallStack:  captureCallStack(),
			})
			return nil
		}
	}

	// if pending queries is too big, we
	pendingQueries.Add(QueryPlanQuery{
		ExecutedAt: time.Now().UnixNano(),
		Query:      stmt.query,
		Duration:   dur.Nanoseconds(),
		CallStack:  captureCallStack(),
	})

	return nil
}
