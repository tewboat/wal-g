package greenplum

import (
	"os"

	"github.com/wal-g/tracelog"
	"github.com/wal-g/wal-g/internal"
	"github.com/wal-g/wal-g/pkg/storages/storage"
	"github.com/wal-g/wal-g/utility"
)

func HandleRestorePointList(folder storage.Folder, metaFetcher internal.GenericMetaFetcher, pretty, json bool) {
	getRestorePointsFunc := func() ([]internal.BackupTimeWithMetadata, error) {
		res, err := GetRestorePoints(folder)
		if _, ok := err.(NoRestorePointsFoundError); ok {
			err = nil
		}

		// TODO: remove this ugly hack to make current restore-point-list work
		backupTimes := make([]internal.BackupTimeWithMetadata, 0)
		for _, rp := range res {
			metadata, _ := metaFetcher.Fetch(rp.Name, folder)
			backupTimes = append(backupTimes, internal.BackupTimeWithMetadata{
				BackupTime: internal.BackupTime{
					BackupName:  rp.Name,
					Time:        rp.Time,
					WalFileName: utility.StripWalFileName(rp.Name),
				},
				GenericMetadata: metadata,
			})
		}
		return backupTimes, err
	}
	writeRestorePointsListFunc := func(restorePoints []internal.BackupTimeWithMetadata) {
		internal.SortBackupTimeWithMetadataSlices(restorePoints)
		switch {
		case json:
			err := internal.WriteAsJSON(restorePoints, os.Stdout, pretty)
			tracelog.ErrorLogger.FatalOnError(err)
		case pretty:
			internal.WritePrettyBackupList(restorePoints, os.Stdout)
		default:
			internal.WriteBackupList(restorePoints, os.Stdout)
		}
	}
	logging := internal.Logging{
		InfoLogger:  tracelog.InfoLogger,
		ErrorLogger: tracelog.ErrorLogger,
	}

	internal.HandleBackupList(getRestorePointsFunc, writeRestorePointsListFunc, logging)
}
