package main

import (
	"reflect"
	"sync"
	"testing"

	"github.com/coreos/go-etcd/etcd"
)

func TestLoadDataSet(t *testing.T) {
	dataSet := LoadDataSet("fixtures/etcd-dump.json")
	value := "testValue"
	expectedDataSet := []BackupKey{BackupKey{Key: "/test", Value: &value}}

	if !reflect.DeepEqual(*dataSet, expectedDataSet) {
		t.Fatal("Unexpected dataSet, expected ", expectedDataSet, "got", *dataSet)
	}
}

func TestRestoreDataSet(t *testing.T) {
	etcdClientTest, _ := initTestClient()
	etcdClientTest.On("SetDir", "/test", 0).Return(&etcd.Response{}, nil)
	backupKeys := []BackupKey{BackupKey{Key: "/test"}}
	RestoreDataSet(backupKeys, config, etcdClientTest)

	etcdClientTest.Mock.AssertExpectations(t)
}

func TestNewRestoreStatistics(t *testing.T) {
	value := "testValue"
	backupKeys := []BackupKey{BackupKey{Key: "/test"}, BackupKey{Key: "/test1", Value: &value}}
	result := NewRestoreStatistics(backupKeys)

	if *result["DataSetSize"] != int32(2) {
		t.Fatal("Unexpected DataSetSize, expected 2 got", result["DataSetSize"])
	}
}

func TestRestoreKey(t *testing.T) {
	backupKey := &BackupKey{Key: "/test"}
	keyFunc := func(backupKey *BackupKey,
		statistics map[string]*int32,
		wg *sync.WaitGroup,
		throttle chan int,
		etcdClientTest *MockedEtcdClient) {
		throttle <- 1
		etcdClientTest.On("SetDir", "/test", 0).Return(&etcd.Response{}, nil)
		RestoreKey(backupKey, statistics, wg, throttle, etcdClientTest)
	}

	checkRestoreKeys(t, backupKey, keyFunc)
}

func TestRestoreDir(t *testing.T) {
	value := "testValue"
	backupKey := &BackupKey{Key: "/test1", Value: &value}
	keyFunc := func(backupKey *BackupKey,
		statistics map[string]*int32,
		wg *sync.WaitGroup,
		throttle chan int,
		etcdClientTest *MockedEtcdClient) {
		throttle <- 1
		etcdClientTest.On("Set", "/test1", value, 0).Return(&etcd.Response{}, nil)
		RestoreKey(backupKey, statistics, wg, throttle, etcdClientTest)
	}

	checkRestoreKeys(t, backupKey, keyFunc)
}

func checkRestoreKeys(t *testing.T,
	backupKey *BackupKey,
	function func(*BackupKey, map[string]*int32, *sync.WaitGroup, chan int, *MockedEtcdClient)) {
	etcdClientTest, _ := initTestClient()
	emptyDirectoriesNbr := int32(0)
	keysInsertedNbr := int32(0)
	statistics := map[string]*int32{"EmptyDirectories": &emptyDirectoriesNbr, "KeysInserted": &keysInsertedNbr}
	throttle := make(chan int, 1)
	var wg sync.WaitGroup

	wg.Add(1)
	go function(backupKey, statistics, &wg, throttle, etcdClientTest)
	wg.Wait()

	if backupKey.Value == nil {
		if *statistics["EmptyDirectories"] != int32(1) {
			t.Fatal("Unexpected EmptyDirectories number, expected 1 got", *statistics["EmptyDirectories"])
		}
	}

	if *statistics["KeysInserted"] != int32(1) {
		t.Fatal("Unexpected KeysInserted number, expected 1 got", *statistics["KeysInserted"])
	}
}

func TestSetKey(t *testing.T) {
	etcdClientTest, _ := initTestClient()
	value := "testValue"
	etcdClientTest.On("Set", "/test", value, 0).Return(&etcd.Response{}, nil)
	setKey(&BackupKey{Key: "/test", Value: &value}, etcdClientTest)

	etcdClientTest.Mock.AssertExpectations(t)
}

func TestSetDirectory(t *testing.T) {
	etcdClientTest, _ := initTestClient()
	etcdClientTest.On("SetDir", "/test", 0).Return(&etcd.Response{}, nil)
	setDirectory(&BackupKey{Key: "/test"}, etcdClientTest)

	etcdClientTest.Mock.AssertExpectations(t)
}
