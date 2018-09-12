psql-migrate
============

A PostgreSQL database migration tool.

This tool has a few goals:
* Minimal
* Production-friendly
* Transactional DDL
* Warn aggressively about unexpected migration files (especially merge "conflicts")
* Don't build sqlite3 drivers for a postgresql tool

The tool is split into a shared library (libmigrate) and a database-specific
command-line tool (this library).

If you need to run a non-transactional migration, start the migration file with this line:

    -- migrate: no-transaction
