package main

import "time"

type Results struct {
	InsertNative    InsertResult
	InsertQueryPlan InsertResult
}

type InsertResult struct {
	TotalDuration   time.Duration
	AverageDuration time.Duration
}

func (r InsertResult) String() string {
	return "Total Duration: " + r.TotalDuration.String() + "\nAverage Duration: " + r.AverageDuration.String()
}

func (r Results) String() string {
	return "Insert Native:\n" + r.InsertNative.String() + "\nInsert Query Plan:\n" + r.InsertQueryPlan.String()
}
