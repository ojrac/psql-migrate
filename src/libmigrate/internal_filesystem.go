package libmigrate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type filesystemWrapper interface {
	CreateFile(version int, name, direction string) error
	EnsureMigrationDir() error
	ListMigrationDir() ([]string, error)
	ReadMigration(filename string) (string, error)
}

type filesystemWrapperImpl struct {
	migrationDir string
}

func (w *filesystemWrapperImpl) ListMigrationDir() (names []string, err error) {
	dir, err := os.Open(w.migrationDir)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return
	}
	names, err = dir.Readdirnames(0)
	return
}

func (w *filesystemWrapperImpl) CreateFile(version int, name, direction string) error {
	fname := path.Join(w.migrationDir, fmt.Sprintf(filenameFmt, version, name, direction))

	f, err := os.Create(fname)
	if err != nil {
		return err
	}

	return f.Close()
}

func (w *filesystemWrapperImpl) EnsureMigrationDir() error {
	if stat, err := os.Stat(w.migrationDir); os.IsNotExist(err) {
		return os.Mkdir(w.migrationDir, os.ModeDir|0775)
	} else if err != nil {
		return err
	} else if !stat.IsDir() {
		return &badMigrationPathError{
			isNotDir: true,
		}
	}

	return nil
}

func (w *filesystemWrapperImpl) ReadMigration(filename string) (sql string, err error) {
	f, err := os.Open(path.Join(w.migrationDir, filename))
	if err != nil {
		return
	}

	migrationSql, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		return
	}

	sql = string(migrationSql)
	return
}
