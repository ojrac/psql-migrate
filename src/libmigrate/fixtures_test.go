package libmigrate

import (
	"context"
	"testing"
)

func Fixture(t *testing.T) (*migrator, *dbMock, *fsMock) {
	db := &dbMock{
		requireSchema: func(ctx context.Context) error { return nil },
		listMigrations: func(ctx context.Context) ([]migration, error) {
			return []migration{}, nil
		},
	}
	fs := &fsMock{
		listMigrationDir: func() ([]string, error) {
			return []string{
				"0001_v1.up.sql",
				"0001_v1.down.sql",
				"0002_v2.up.sql",
				"0002_v2.down.sql",
				"0003_v3.up.sql",
				"0003_v3.down.sql",
			}, nil
		},
		readMigration: func(name string) (string, error) { return "", nil },
	}

	return &migrator{
		db:         db,
		filesystem: fs,
	}, db, fs
}

type dbMock struct {
	applyMigration func(ctx context.Context, useTx, isUp bool, version int, name, query string) error
	requireSchema  func(ctx context.Context) error
	listMigrations func(ctx context.Context) ([]migration, error)
	getVersion     func(ctx context.Context) (int, error)
}

func (m dbMock) ApplyMigration(ctx context.Context, useTx, isUp bool, version int, name, query string) error {
	return m.applyMigration(ctx, useTx, isUp, version, name, query)
}
func (m dbMock) RequireSchema(ctx context.Context) error {
	return m.requireSchema(ctx)
}

func (m dbMock) ListMigrations(ctx context.Context) ([]migration, error) {
	return m.listMigrations(ctx)
}

func (m dbMock) GetVersion(ctx context.Context) (int, error) {
	return m.getVersion(ctx)
}

type fsMock struct {
	createFile         func(version int, name, direction string) error
	ensureMigrationDir func() error
	listMigrationDir   func() ([]string, error)
	readMigration      func(filename string) (string, error)
}

func (m fsMock) CreateFile(version int, name, direction string) error {
	return m.createFile(version, name, direction)
}
func (m fsMock) EnsureMigrationDir() error {
	return m.ensureMigrationDir()
}
func (m fsMock) ListMigrationDir() ([]string, error) {
	return m.listMigrationDir()
}
func (m fsMock) ReadMigration(filename string) (string, error) {
	return m.readMigration(filename)
}
