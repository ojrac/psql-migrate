package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path"

	_ "github.com/lib/pq"

	"github.com/ojrac/libmigrate"
)

var usage string = `Psql-migrate is a tool for managing a PostgreSQL database's schema.

Usage:

	psql-migrate <version> [arguments]
	psql-migrate command [arguments]

The commands are:

	[number]	Migrate to the given migration version number
	latest		Migrate to the latest migration version on disk
	create		Create a new up and down migration in the migration directory
	version		Print the current migration version
	pending		Print true if there are unapplied migrations, or else false

All commands require a connection information, which can be set by command-line
flags or environment variables:

	--conn-str (env: MIGRATIONS_CONN_STR): lib/pq connection string, e.g.
		"user=xxx dbname=xxx password=xxx"
	--migrations-path (env: MIGRATIONS_PATH): Directory containing migration files
		(default: ./migrations)
	--schema (env: MIGRATIONS_SCHEMA): Schema to hold the migration table. (default: public)
	--table (env: MIGRATIONS_TABLE): Table to track migration status. (default: migration_version)

`

var connStr string
var migrationsPath string
var migrationsTable string
var migrationsSchema string

func parseFlags() {
	// Defaults
	migrationsPath = path.Join(".", "migrations")
	migrationsTable = "migration_version"
	migrationsSchema = "public"

	parseEnv(map[string]*string{
		"MIGRATIONS_CONN_STR": &connStr,
		"MIGRATIONS_PATH":     &migrationsPath,
		"MIGRATIONS_SCHEMA":   &migrationsSchema,
		"MIGRATIONS_TABLE":    &migrationsTable,
	})

	flag.CommandLine.Usage = func() { fmt.Printf(usage) }
	flag.StringVar(&connStr, "conn-str", connStr, "")
	flag.StringVar(&migrationsPath, "migrations-path", migrationsPath, "")
	flag.StringVar(&migrationsSchema, "schema", migrationsSchema, "")
	flag.StringVar(&migrationsTable, "table", migrationsTable, "")
	flag.Parse()
}

func main() {
	parseFlags()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Error connecting to postgres: %+v\n", err)
		os.Exit(1)
		return
	}
	defer db.Close()

	m := libmigrate.New(db, migrationsPath, libmigrate.ParamTypeDollarSign)

	m.SetTableName(migrationsTable)
	m.SetTableSchema(migrationsSchema)

	run(m, flag.Args())
}
