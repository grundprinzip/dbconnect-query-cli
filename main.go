package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/apache/spark-connect-go/spark/sql"
	"github.com/databricks/databricks-sdk-go/config"
	"github.com/grundprinzip/unofficial-dbconnect-go/dbconnect"
)

type Output struct {
	CSV                 string  `json:"csv"`
	ExplainPlan         string  `json:"explain_plan"`
	TimeToFirstResponse float64 `json:"time_to_first_response_ms"`
	TotalTime           float64 `json:"total_time_ms"`
}

func run() error {
	profile := flag.String("profile", "DEFAULT", "Databricks authentication profile")
	clusterID := flag.String("cluster-id", "serverless", "Cluster ID to use, or \"serverless\" for serverless compute")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <sql-query>\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return fmt.Errorf("missing required sql query argument")
	}
	query := args[0]

	ctx := context.Background()
	cb := dbconnect.NewDataBricksChannelBuilder()
	cb = cb.WithConfig(&config.Config{Profile: *profile})
	if *clusterID == "serverless" {
		cb = cb.UseServerless()
	} else {
		cb = cb.UseCluster(*clusterID)
	}

	spark, err := sql.NewSessionBuilder().WithChannelBuilder(cb).Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer spark.Stop()

	// Best-effort: collect the EXPLAIN EXTENDED plan before running the actual query.
	var explainBuf bytes.Buffer
	if explainDF, err := spark.Sql(ctx, fmt.Sprintf("EXPLAIN EXTENDED %s", query)); err == nil {
		if explainRows, err := explainDF.Collect(ctx); err == nil {
			for _, row := range explainRows {
				explainBuf.WriteString(fmt.Sprintf("%v", row.At(0)))
			}
		}
	}

	// Execute the actual query and measure timing.
	startTime := time.Now()

	df, err := spark.Sql(ctx, query)
	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}
	// Sql() performs a full server roundtrip (ExecuteCommand) that parses,
	// plans, and resolves the query. This marks the first server response.
	timeToFirstResponse := time.Since(startTime)

	rows, err := df.Collect(ctx)
	if err != nil {
		return fmt.Errorf("result collection failed: %w", err)
	}
	totalTime := time.Since(startTime)

	schema, err := df.Schema(ctx)
	if err != nil {
		return fmt.Errorf("schema retrieval failed: %w", err)
	}

	// Format collected rows as CSV.
	var csvBuf bytes.Buffer
	w := csv.NewWriter(&csvBuf)

	header := make([]string, len(schema.Fields))
	for i, f := range schema.Fields {
		header[i] = f.Name
	}
	w.Write(header)

	for _, row := range rows {
		record := make([]string, len(schema.Fields))
		for i := range schema.Fields {
			val := row.At(i)
			if val == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", val)
			}
		}
		w.Write(record)
	}
	w.Flush()

	output := Output{
		CSV:                 csvBuf.String(),
		ExplainPlan:         explainBuf.String(),
		TimeToFirstResponse: float64(timeToFirstResponse.Microseconds()) / 1000.0,
		TotalTime:           float64(totalTime.Microseconds()) / 1000.0,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
