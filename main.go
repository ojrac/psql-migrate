package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path"

	_ "github.com/lib/pq"

	"libmigrate"
)

var usage string = `Psql-migrate is a tool for managing a PostgreSQL database's schema.

Usage:

	psql-migrate [version] [arguments]
	psql-migrate command [arguments]

The commands are:

	(none)		Migrate to the latest migration version on disk
	[number]	Migrate to the given migration version number
	create		Create a new up and down migration in the migration directory
	version		Print the current migration version
	pending		Print true if there are unapplied migrations, or else false

All commands require a connection information, which can be set by command-line
flags or environment variables:

	--conn-str (env: CONN_STR): lib/pq connection string, e.g.
		"user=xxx dbname=xxx password=xxx"
	--migrations-path (env: MIGRATIONS_PATH): Directory containing migration files
		(default: ./migrations)

`

var connStr string
var migrationsPath string

func parseFlags() {
	migrationsPath = path.Join(".", "migrations")
	parseEnv(map[string]*string{
		"CONN_STR":        &connStr,
		"MIGRATIONS_PATH": &migrationsPath,
	})

	flag.CommandLine.Usage = func() { fmt.Printf(usage) }
	flag.StringVar(&connStr, "conn-str", connStr, "")
	flag.StringVar(&migrationsPath, "migrations-path", migrationsPath, "")
	flag.Parse()
}

func main() {
	parseFlags()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Error connecting to postgres: %+v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	m := libmigrate.New(db, migrationsPath, libmigrate.ParamTypeDollarSign)
	run(m, flag.Args())
}
