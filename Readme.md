psql-migrate
============

A PostgreSQL database migration tool.

This tool has a few goals:
* Minimal
* Production-friendly
* Transactional DDL
* Warn aggressively about unexpected migration files (especially merge "conflicts")
* Don't build sqlite3 drivers for a postgresql tool

Usage
-----

The tool looks for migrations in a directory, `./migrations` by default, with
names like `0001_initial.up.sql` and `0001_initial.down.sql`. As a shortcut,
the tool can create empty up and down migration files for you. If you run
migrations as part of an automated process, you can use the output of
`psql-migrate pending` to determine when to block production deploys until the
migrations are applied manually.

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

	--conn-str (env: CONN_STR): lib/pq connection string, e.g.
		"user=xxx dbname=xxx password=xxx"
	--migrations-path (env: MIGRATIONS_PATH): Directory containing migration files
		(default: ./migrations)

Structure
---------

The tool is split into a shared library (libmigrate) and a database-specific
command-line tool (this library).

If you need to run a non-transactional migration, start the migration file with this line:

    -- migrate: no-transaction
