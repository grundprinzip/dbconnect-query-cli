# dbquery — Databricks SQL CLI

A command-line tool that executes SQL queries against Databricks (serverless or cluster) and returns JSON with the query results, explain plan, and execution timing.

Built on [unofficial-dbconnect-go](https://github.com/grundprinzip/unofficial-dbconnect-go).

## Prerequisites

- Go 1.23+
- A configured [Databricks CLI profile](https://docs.databricks.com/en/dev-tools/cli/profiles.html) in `~/.databrickscfg`

## Build & Install

```bash
go build -o dbquery .
```

Move the binary somewhere on your PATH:

```bash
sudo mv dbquery /usr/local/bin/
```

Or add the project directory to your PATH instead.

## Usage

```bash
dbquery [flags] "<sql-query>"
```

| Flag | Default | Description |
|------|---------|-------------|
| `-profile` | `DEFAULT` | Databricks authentication profile |
| `-cluster-id` | `serverless` | Cluster ID, or `"serverless"` for serverless compute |

### Example

```bash
dbquery -profile DEFAULT "SELECT count(*) FROM samples.nyctaxi.trips"
```

Output (JSON on stdout):

```json
{
  "csv": "count(1)\n245205\n",
  "explain_plan": "== Physical Plan ==\n...",
  "time_to_first_response_ms": 1823.456,
  "total_time_ms": 2104.789
}
```

## Installing the Claude Code Skill

The included `SKILL.md` teaches Claude Code how to use `dbquery` to run SQL against Databricks on your behalf.

### Option 1: Symlink into your Claude Code skills directory

```bash
# Create the skills directory if it doesn't exist
mkdir -p ~/.claude/skills

# Symlink this project's SKILL.md
ln -s "$(pwd)/SKILL.md" ~/.claude/skills/dbquery-cli.md
```

### Option 2: Copy the skill file

```bash
mkdir -p ~/.claude/skills
cp SKILL.md ~/.claude/skills/dbquery-cli.md
```

### Option 3: Project-level skill

To make the skill available only within a specific project, place or symlink `SKILL.md` into that project's `.claude/skills/` directory:

```bash
mkdir -p /path/to/your/project/.claude/skills
ln -s "$(pwd)/SKILL.md" /path/to/your/project/.claude/skills/dbquery-cli.md
```

After installing, Claude Code will automatically detect and use the skill whenever you ask it to run SQL on Databricks, inspect query plans, or measure query performance.
