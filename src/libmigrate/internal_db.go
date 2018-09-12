package libmigrate

import (
	"context"
	"database/sql"
	"fmt"
)

type dbWrapper interface {
	ApplyMigration(ctx context.Context, useTx, isUp bool, version int, name, query string) error
	RequireSchema(ctx context.Context) error
	ListMigrations(ctx context.Context) ([]migration, error)
	GetVersion(ctx context.Context) (int, error)
}

type dbWrapperImpl struct {
	db        DB
	paramType ParamType
	tableName string
}

type dbOrTx interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (w *dbWrapperImpl) ApplyMigration(ctx context.Context, useTx, isUp bool, version int, name, query string) (err error) {
	var db dbOrTx = w.db
	if useTx {
		var tx *sql.Tx
		tx, err = w.db.BeginTx(ctx, nil)
		if err != nil {
			return
		}

		db = tx
		defer func() {
			if err == nil {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}()
	}

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return &migrateError{cause: err}
	}

	paramFunc, err := w.paramType.getFunc()
	if err != nil {
		return
	}
	if isUp {
		_, err = db.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO "%s"
						(version, name)
				 VALUES (%s, %s)
		`, w.tableName, paramFunc(), paramFunc()),
			version, name)
	} else {
		_, err = db.ExecContext(ctx, fmt.Sprintf(`
			DELETE FROM "%s"
				  WHERE version = %s
						AND name = %s
		`, w.tableName, paramFunc(), paramFunc()),
			version, name)
	}
	return err
}

func (w *dbWrapperImpl) RequireSchema(ctx context.Context) error {
	_, err := w.db.ExecContext(ctx, fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS "%s" (
		version integer PRIMARY KEY NOT NULL,
		name text NOT NULL
	);`, w.tableName))
	return err
}

func (w *dbWrapperImpl) ListMigrations(ctx context.Context) (result []migration, err error) {
	rows, err := w.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT version, name
		  FROM "%s"
	  ORDER BY version ASC
	`, w.tableName))
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m migration
		err = rows.Scan(&m.Version, &m.Name)
		if err != nil {
			return
		}

		result = append(result, m)
	}

	return
}

func (w *dbWrapperImpl) GetVersion(ctx context.Context) (version int, err error) {
	err = w.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT coalesce(max(version), 0)
		  FROM "%s"
		  `, w.tableName)).Scan(&version)
	return
}
