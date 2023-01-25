package internal_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wal-g/wal-g/internal"
	"github.com/wal-g/wal-g/testtools"
)

type someError struct {
	error
}

var shortBackups = []internal.BackupTimeWithMetadata{
	{
		BackupTime: internal.BackupTime{
			BackupName:  "b0",
			Time:        time.Time{},
			WalFileName: "shortWallName0",
		},
		GenericMetadata: internal.GenericMetadata{
			StartTime: time.Time{},
		},
	},
	{
		BackupTime: internal.BackupTime{
			BackupName:  "b1",
			Time:        time.Time{},
			WalFileName: "shortWallName1",
		},
		GenericMetadata: internal.GenericMetadata{
			StartTime: time.Time{},
		},
	},
}

var longBackups = []internal.BackupTimeWithMetadata{
	{
		BackupTime: internal.BackupTime{
			BackupName:  "backup000",
			Time:        time.Time{},
			WalFileName: "veryVeryVeryVeryVeryLongWallName0",
		},
		GenericMetadata: internal.GenericMetadata{
			StartTime: time.Time{},
		},
	},
	{
		BackupTime: internal.BackupTime{
			BackupName:  "backup001",
			Time:        time.Time{},
			WalFileName: "veryVeryVeryVeryVeryLongWallName1",
		},
		GenericMetadata: internal.GenericMetadata{
			StartTime: time.Time{},
		},
	},
}

var emptyColonsBackups = []internal.BackupTimeWithMetadata{
	{
		BackupTime: internal.BackupTime{
			Time:        time.Time{},
			WalFileName: "shortWallName0",
		},
		GenericMetadata: internal.GenericMetadata{
			StartTime: time.Time{},
		},
	},
	{
		BackupTime: internal.BackupTime{
			BackupName: "b1",
			Time:       time.Time{},
		},
		GenericMetadata: internal.GenericMetadata{
			StartTime: time.Time{},
		},
	},
	{},
}

func TestHandleBackupListWriteBackups(t *testing.T) {
	backups := []internal.BackupTimeWithMetadata{
		{
			BackupTime: internal.BackupTime{
				BackupName:  "backup000",
				Time:        time.Time{},
				WalFileName: "wallName0",
			},
			GenericMetadata: internal.GenericMetadata{
				StartTime: time.Date(2016, 3, 21, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			BackupTime: internal.BackupTime{
				BackupName:  "backup001",
				Time:        time.Time{},
				WalFileName: "wallName1",
			},
			GenericMetadata: internal.GenericMetadata{
				StartTime: time.Date(2017, 3, 21, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	getBackupsFunc := func() ([]internal.BackupTimeWithMetadata, error) {
		return backups, nil
	}
	writeBackupListCallsCount := 0
	var writeBackupListCallArgs []internal.BackupTimeWithMetadata
	writeBackupListFunc := func(backups []internal.BackupTimeWithMetadata) {
		writeBackupListCallsCount++
		writeBackupListCallArgs = backups
	}
	infoLogger, errorLogger := testtools.MockLoggers()

	internal.HandleBackupList(
		getBackupsFunc,
		writeBackupListFunc,
		internal.Logging{InfoLogger: infoLogger, ErrorLogger: errorLogger},
	)

	assert.Equal(t, 1, writeBackupListCallsCount)
	assert.Equal(t, backups, writeBackupListCallArgs)
}

func TestHandleBackupListLogError(t *testing.T) {
	backups := []internal.BackupTimeWithMetadata{
		{
			BackupTime: internal.BackupTime{
				BackupName:  "backup000",
				Time:        time.Time{},
				WalFileName: "wallName0",
			},
			GenericMetadata: internal.GenericMetadata{
				StartTime: time.Date(2016, 3, 21, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			BackupTime: internal.BackupTime{
				BackupName:  "backup001",
				Time:        time.Time{},
				WalFileName: "wallName1",
			},
			GenericMetadata: internal.GenericMetadata{
				StartTime: time.Date(2017, 3, 21, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	someErrorInstance := someError{errors.New("some error")}
	getBackupsFunc := func() ([]internal.BackupTimeWithMetadata, error) {
		return backups, someErrorInstance
	}
	writeBackupListFunc := func(backups []internal.BackupTimeWithMetadata) {}
	infoLogger, errorLogger := testtools.MockLoggers()

	internal.HandleBackupList(
		getBackupsFunc,
		writeBackupListFunc,
		internal.Logging{InfoLogger: infoLogger, ErrorLogger: errorLogger},
	)

	assert.Equal(t, 1, errorLogger.Stats.FatalOnErrorCallsCount)
	assert.Equal(t, someErrorInstance, errorLogger.Stats.Err)
}

func TestHandleBackupListLogNoBackups(t *testing.T) {
	getBackupsFunc := func() ([]internal.BackupTimeWithMetadata, error) {
		return []internal.BackupTimeWithMetadata{}, nil
	}
	writeBackupListFunc := func(backups []internal.BackupTimeWithMetadata) {}
	infoLogger, errorLogger := testtools.MockLoggers()

	internal.HandleBackupList(
		getBackupsFunc,
		writeBackupListFunc,
		internal.Logging{InfoLogger: infoLogger, ErrorLogger: errorLogger},
	)

	assert.Equal(t, 1, infoLogger.Stats.PrintLnCallsCount)
	assert.Equal(t, "No backups found", infoLogger.Stats.PrintMsg)
	assert.Equal(t, 1, errorLogger.Stats.FatalOnErrorCallsCount)
	assert.Equal(t, nil, errorLogger.Stats.Err)
}

func TestWritePrettyBackupList_LongColumnsValues(t *testing.T) {
	expectedRes := "+---+-----------+---------+-----------------------------------+\n" +
		"| # | NAME      | CREATED | WAL SEGMENT BACKUP START          |\n" +
		"+---+-----------+---------+-----------------------------------+\n" +
		"| 0 | backup000 | -       | veryVeryVeryVeryVeryLongWallName0 |\n" +
		"| 1 | backup001 | -       | veryVeryVeryVeryVeryLongWallName1 |\n" +
		"+---+-----------+---------+-----------------------------------+\n"

	b := bytes.Buffer{}
	internal.WritePrettyBackupList(longBackups, &b)

	assert.Equal(t, expectedRes, b.String())
}

func TestWritePrettyBackupList_ShortColumnsValues(t *testing.T) {
	expectedRes := "+---+------+---------+--------------------------+\n" +
		"| # | NAME | CREATED | WAL SEGMENT BACKUP START |\n" +
		"+---+------+---------+--------------------------+\n" +
		"| 0 | b0   | -       | shortWallName0           |\n" +
		"| 1 | b1   | -       | shortWallName1           |\n" +
		"+---+------+---------+--------------------------+\n"

	b := bytes.Buffer{}
	internal.WritePrettyBackupList(shortBackups, &b)

	assert.Equal(t, expectedRes, b.String())
}

func TestWritePrettyBackupList_WriteNoBackupList(t *testing.T) {
	expectedRes := "+---+------+---------+--------------------------+\n" +
		"| # | NAME | CREATED | WAL SEGMENT BACKUP START |\n" +
		"+---+------+---------+--------------------------+\n" +
		"+---+------+---------+--------------------------+\n"

	backups := make([]internal.BackupTimeWithMetadata, 0)

	b := bytes.Buffer{}
	internal.WritePrettyBackupList(backups, &b)

	assert.Equal(t, expectedRes, b.String())
}

func TestWritePrettyBackupList_EmptyColumnsValues(t *testing.T) {
	expectedRes := "+---+------+---------+--------------------------+\n" +
		"| # | NAME | CREATED | WAL SEGMENT BACKUP START |\n" +
		"+---+------+---------+--------------------------+\n" +
		"| 0 |      | -       | shortWallName0           |\n" +
		"| 1 | b1   | -       |                          |\n" +
		"| 2 |      | -       |                          |\n" +
		"+---+------+---------+--------------------------+\n"

	b := bytes.Buffer{}
	internal.WritePrettyBackupList(emptyColonsBackups, &b)

	assert.Equal(t, expectedRes, b.String())
}

func TestWriteBackupList_NoBackups(t *testing.T) {
	expectedRes := "name created wal_segment_backup_start\n"
	backups := make([]internal.BackupTimeWithMetadata, 0)

	b := bytes.Buffer{}
	internal.WriteBackupList(backups, &b)

	assert.Equal(t, expectedRes, b.String())
}

func TestWriteBackupList_EmptyColumnsValues(t *testing.T) {
	expectedRes := "name created wal_segment_backup_start\n" +
		"     -       shortWallName0\n" +
		"b1   -       \n" +
		"     -       \n"

	b := bytes.Buffer{}
	internal.WriteBackupList(emptyColonsBackups, &b)

	assert.Equal(t, expectedRes, b.String())
}

func TestWriteBackupList_ShortColumnsValues(t *testing.T) {
	expectedRes := "name created wal_segment_backup_start\n" +
		"b0   -       shortWallName0\n" +
		"b1   -       shortWallName1\n"
	b := bytes.Buffer{}
	internal.WriteBackupList(shortBackups, &b)

	assert.Equal(t, expectedRes, b.String())
}

func TestWriteBackupList_LongColumnsValues(t *testing.T) {
	expectedRes := "name      created wal_segment_backup_start\n" +
		"backup000 -       veryVeryVeryVeryVeryLongWallName0\n" +
		"backup001 -       veryVeryVeryVeryVeryLongWallName1\n"
	b := bytes.Buffer{}
	internal.WriteBackupList(longBackups, &b)

	assert.Equal(t, expectedRes, b.String())
}
