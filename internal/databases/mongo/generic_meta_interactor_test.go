package mongo_test

import (
	"github.com/wal-g/wal-g/internal/databases/mongo"
	"github.com/wal-g/wal-g/internal/databases/mongo/models"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wal-g/wal-g/internal"
	"github.com/wal-g/wal-g/testtools"
)

func init() {
	internal.ConfigureSettings("")
	internal.InitConfig()
	internal.Configure()
}

func TestFetch(t *testing.T) {
	backupName := "test"
	backupType := "type"
	hostname := "hostname"
	data := "Data"
	uncompressedSize := rand.Int63()
	compressedSize := rand.Int63()

	date := time.Date(2022, 3, 21, 0, 0, 0, 0, time.UTC)
	isPermanent := false

	testObject := models.Backup{
		BackupName:       backupName,
		BackupType:       backupType,
		Hostname:         hostname,
		StartLocalTime:   date,
		FinishLocalTime:  date,
		UserData:         data,
		MongoMeta:        models.MongoMeta{},
		Permanent:        isPermanent,
		UncompressedSize: uncompressedSize,
		CompressedSize:   compressedSize,
	}

	expectedResult := internal.GenericMetadata{
		BackupName:       backupName,
		UncompressedSize: uncompressedSize,
		CompressedSize:   compressedSize,
		Hostname:         hostname,
		StartTime:        date,
		FinishTime:       date,
		IsPermanent:      isPermanent,
		IncrementDetails: &internal.NopIncrementDetailsFetcher{},
		UserData:         data,
	}

	folder := testtools.MakeDefaultInMemoryStorageFolder()
	marshaller, _ := internal.NewDtoSerializer()
	file, _ := marshaller.Marshal(testObject)
	_ = folder.PutObject(internal.SentinelNameFromBackup(backupName), file)
	actualResult, err := mongo.NewGenericMetaFetcher().Fetch(backupName, folder)

	//check equality of time separately
	isEqualTimeStart := expectedResult.StartTime.Equal(actualResult.StartTime)
	assert.True(t, isEqualTimeStart)

	isEqualTimeFinish := expectedResult.FinishTime.Equal(actualResult.FinishTime)
	assert.True(t, isEqualTimeFinish)

	// since assert.Equal doesn't compare time properly, just assign the actual to the expected time
	expectedResult.StartTime = actualResult.StartTime
	expectedResult.FinishTime = actualResult.FinishTime

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, actualResult)
}

func TestSetIsPermanent(t *testing.T) {
	backupName := "test"
	testObject := models.Backup{
		UserData:       nil,
		StartLocalTime: time.Now(),
		Permanent:      false,
	}

	folder := testtools.MakeDefaultInMemoryStorageFolder()
	marshaller, _ := internal.NewDtoSerializer()
	file, _ := marshaller.Marshal(testObject)
	_ = folder.PutObject(internal.SentinelNameFromBackup(backupName), file)

	_ = mongo.NewGenericMetaSetter().SetIsPermanent(backupName, folder, true)
	backup, err := mongo.NewGenericMetaFetcher().Fetch(backupName, folder)

	assert.NoError(t, err)
	assert.True(t, backup.IsPermanent)
}

func TestSetUserData(t *testing.T) {
	backupName := "test"
	updatedData := "Updated Data"
	oldData := "Old Data"
	testObject := models.Backup{
		UserData:       oldData,
		StartLocalTime: time.Now(),
		Permanent:      false,
	}

	folder := testtools.MakeDefaultInMemoryStorageFolder()
	marshaller, _ := internal.NewDtoSerializer()
	file, _ := marshaller.Marshal(testObject)
	_ = folder.PutObject(internal.SentinelNameFromBackup(backupName), file)

	_ = mongo.NewGenericMetaSetter().SetUserData(backupName, folder, updatedData)

	backup, err := mongo.NewGenericMetaFetcher().Fetch(backupName, folder)

	assert.NoError(t, err)
	assert.Equal(t, updatedData, backup.UserData)
}
