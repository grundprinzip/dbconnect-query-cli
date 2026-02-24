---
name: dbquery-cli
description: Run SQL queries against Databricks and analyze results using the dbquery CLI tool. Use when the user wants to execute SQL on Databricks, inspect query plans, measure query performance, or retrieve tabular data from Databricks clusters or serverless compute.
metadata:
  short-description: Execute SQL on Databricks and return results, explain plan, and timing as JSON
---

# dbquery CLI

Execute SQL queries against Databricks via Spark Connect and return structured JSON with results, explain plan, and timing.

## Usage

`dbquery` is available in PATH.

```bash
dbquery [flags] "<sql-query>"
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-profile` | `DEFAULT` | Databricks authentication profile from `~/.databrickscfg` |
| `-cluster-id` | `serverless` | Cluster ID, or `"serverless"` for serverless compute |

### Examples

```bash
# Serverless with default profile
dbquery "SELECT count(*) FROM samples.nyctaxi.trips"

# Specific profile and cluster
dbquery -profile PROD -cluster-id 0123-456789-abcdef "SELECT 1"
```

## Output Format

The tool writes a single JSON object to **stdout**. Errors go to **stderr** with a non-zero exit code.

```json
{
  "csv": "col1,col2\nval1,val2\n",
  "explain_plan": "== Physical Plan ==\n...",
  "time_to_first_response_ms": 1234.567,
  "total_time_ms": 5678.901
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `csv` | string | Query results as a single CSV string with a header row followed by data rows, newline-separated. NULL values are represented as empty strings. |
| `explain_plan` | string | Full output of `EXPLAIN EXTENDED <query>`, captured before the query runs. Contains parsed logical plan, analyzed logical plan, optimized logical plan, and physical plan. |
| `time_to_first_response_ms` | float | Milliseconds from query submission to first server response (query parsed, planned, and acknowledged). |
| `total_time_ms` | float | Milliseconds from query submission to all result rows collected. Includes execution, data transfer, and serialization. |

### Interpreting Timing

- **time_to_first_response_ms** reflects the server roundtrip for the `Sql()` call — the query is sent, parsed, planned, and a relation reference is returned. This does not include data transfer.
- **total_time_ms** includes the above plus `Schema()` and `Collect()` calls that trigger actual execution and stream back all result rows.
- The difference `total_time_ms - time_to_first_response_ms` approximates the data retrieval and transfer time.

### Parsing the Output

To extract the CSV data programmatically:

```bash
dbquery "SELECT 1 AS n" | jq -r '.csv'
```

To get just the explain plan:

```bash
dbquery "SELECT 1 AS n" | jq -r '.explain_plan'
```
