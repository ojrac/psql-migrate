package libmigrate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateMigrationList(t *testing.T) {
	up := []migration{
		migration{Version: 1, Name: "test"},
		migration{Version: 2, Name: "test"},
		migration{Version: 3, Name: "test"},
	}
	down := []migration{
		migration{Version: 1, Name: "test"},
		migration{Version: 2, Name: "test"},
		migration{Version: 3, Name: "test"},
	}

	err := validateMigrationList(true, up, down)
	require.NoError(t, err)
}

func TestMissingMiddleMigration(t *testing.T) {
	up := []migration{
		migration{Version: 1, Name: "test"},
		migration{Version: 3, Name: "test"},
	}
	down := []migration{
		migration{Version: 1, Name: "test"},
		migration{Version: 3, Name: "test"},
	}

	err := validateMigrationList(true, up, down)
	require.Equal(t, &missingMigrationError{
		number: 2,
		isUp:   true,
	}, err)
}

func TestMissingLastMigration(t *testing.T) {
	up := []migration{
		migration{Version: 1, Name: "test"},
		migration{Version: 2, Name: "test"},
	}
	down := []migration{
		migration{Version: 1, Name: "test"},
	}

	err := validateMigrationList(true, up, down)
	require.Equal(t, &missingMigrationError{
		number: 2,
		isUp:   false,
	}, err)
}
