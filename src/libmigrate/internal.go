package libmigrate

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const filenameFmt = "%04d_%s.%s.sql"

type migrator struct {
	db                  dbWrapper
	filesystem          filesystemWrapper
	disableTransactions bool
}

type paramFunc func() string

func (t ParamType) getFunc() (paramFunc, error) {
	switch t {
	case ParamTypeQuestionMark:
		return func() string { return "?" }, nil
	case ParamTypeDollarSign:
		var i = 0
		return func() string {
			i++
			return fmt.Sprintf("$%d", i)
		}, nil
	}

	return nil, &unknownParamTypeError{paramType: t}
}

type migration struct {
	Version int
	Name    string
}

func (m migration) Filename(isUp bool) string {
	direction := "up"
	if !isUp {
		direction = "down"
	}
	return fmt.Sprintf(filenameFmt, m.Version, m.Name, direction)
}

func (m *migrator) listMigrations(ctx context.Context) (result []migration, err error) {
	err = m.db.RequireSchema(ctx)
	if err != nil {
		return
	}

	names, err := m.filesystem.ListMigrationDir()
	return m.filenamesToMigrations(ctx, names)
}

func (m *migrator) filenamesToMigrations(ctx context.Context, names []string) (result []migration, err error) {
	upList := make([]migration, 0, len(names))
	downList := make([]migration, 0, len(names))
	for _, s := range names {
		up := strings.HasSuffix(s, ".up.sql")
		down := strings.HasSuffix(s, ".down.sql")
		if !up && !down {
			continue
		}

		parts := strings.SplitN(s, "_", 2)

		var version int
		version, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, &badMigrationFilenameError{
				filename: s,
				cause:    err,
			}
		}

		name := parts[1]
		if up {
			name = strings.TrimSuffix(name, ".up.sql")
		} else {
			name = strings.TrimSuffix(name, ".down.sql")
		}

		if up {
			upList = append(upList, migration{
				Version: version,
				Name:    name,
			})
		} else {
			downList = append(downList, migration{
				Version: version,
				Name:    name,
			})
		}
	}

	sort.Slice(upList, func(i, j int) bool { return upList[i].Version < upList[j].Version })
	sort.Slice(downList, func(i, j int) bool { return downList[i].Version < downList[j].Version })

	err = validateMigrationList(true, upList, downList)
	if err == nil {
		err = validateMigrationList(false, downList, upList)
	}
	if err != nil {
		return nil, err
	}
	err = validateMigrationListsMatch(upList, downList)
	if err != nil {
		return nil, err
	}

	err = m.testForUnknownMigrations(ctx, upList)
	if err != nil {
		return nil, err
	}

	result = upList

	return result, nil
}

func (m *migrator) testForUnknownMigrations(ctx context.Context, migrations []migration) (err error) {
	dbMigrations, err := m.db.ListMigrations(ctx)

	for _, dbMigration := range dbMigrations {
		if dbMigration.Version > len(migrations) {
			return &filesystemMissingDbMigrationError{
				version: dbMigration.Version,
			}
		}

		migration := migrations[dbMigration.Version-1]
		if migration != dbMigration {
			return &filesystemMigrationMismatchError{
				version:        dbMigration.Version,
				dbName:         dbMigration.Name,
				filesystemName: migration.Name,
			}
		}
	}

	return nil
}

func validateMigrationListsMatch(upList, downList []migration) error {
	if len(upList) != len(downList) {
		isUp := len(upList) < len(downList)
		var number int
		if isUp {
			number = len(upList) + 1
		} else {
			number = len(downList) + 1
		}
		return &missingMigrationError{
			number: number,
			isUp:   isUp,
		}
	}

	for i, upMigration := range upList {
		downMigration := downList[i]
		if upMigration.Name != downMigration.Name {
			return &migrationNameMismatchError{
				version:  upMigration.Version,
				upName:   upMigration.Name,
				downName: downMigration.Name,
			}
		}
	}

	return nil
}

func validateMigrationList(isUp bool, migrations, otherDirection []migration) error {
	for i, migration := range migrations {
		idx := i + 1
		if idx != migration.Version {
			return &missingMigrationError{
				number: idx,
				isUp:   isUp,
			}
		}

		if i >= len(otherDirection) {
			return &missingMigrationError{
				number: idx,
				isUp:   !isUp,
			}
		}
	}

	return nil
}

func (m *migrator) useTx(sql string) bool {
	if m.disableTransactions {
		return false
	}
	if strings.HasPrefix(sql, NoTransactionPrefix) {
		return false
	}
	return true
}

func (m *migrator) internalMigrate(ctx context.Context, migration migration, isUp bool) (err error) {
	note := "+"
	if !isUp {
		note = "-"
	}
	fmt.Printf(" %s %s\n", note, migration.Filename(isUp))

	sqlString, err := m.filesystem.ReadMigration(migration.Filename(isUp))
	if err != nil {
		return
	}

	useTx := m.useTx(sqlString)
	return m.db.ApplyMigration(ctx, useTx, isUp, migration.Version, migration.Name, sqlString)
}
