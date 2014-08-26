package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

func LoadDataSet(dumpFilePath string) *[]BackupKey {
	file, error := os.Open(dumpFilePath)
	defer file.Close()
	if error != nil {
		config.LogFatal("Error when trying to open the file `"+dumpFilePath+"`. Error: ", error)
	}

	jsonDataSet := &[]BackupKey{}
	jsonParser := json.NewDecoder(file)
	if err := jsonParser.Decode(jsonDataSet); err != nil {
		config.LogFatal("Error when trying to load data set into json. Error: ", err)
	}

	return jsonDataSet
}

func RestoreDataSet(backupKeys []BackupKey, config *Config, etcdClient EtcdClient) {
	statistics := NewRestoreStatistics(backupKeys)
	throttle := make(chan int, config.ConcurrentRequests)

	var wg sync.WaitGroup

	for index := range backupKeys {
		backupKey := backupKeys[index]
		RestoreKeyWithThrottle(&backupKey, statistics, &wg, throttle, etcdClient)
	}

	wg.Wait()
	printStatistics(statistics)
}

func NewRestoreStatistics(backupKeys []BackupKey) map[string]*int32 {
	DataSetSize := int32(len(backupKeys))
	KeysInserted := int32(0)
	DirectoriesInserted := int32(0)

	return map[string]*int32{
		"DataSetSize":      &DataSetSize,
		"KeysInserted":     &KeysInserted,
		"EmptyDirectories": &DirectoriesInserted,
	}
}

func printStatistics(statistics map[string]*int32) {
	config.LogPrintln("Backup restored succesfully! Results:")
	for keyName, value := range statistics {
		config.LogPrintln(keyName + ": " + fmt.Sprintf("%#v", *value))
	}
}

func RestoreKeyWithThrottle(backupKey *BackupKey, statistics map[string]*int32, wg *sync.WaitGroup, throttle chan int, etcdClient EtcdClient) {
	if backupKey.MatchBackupStrategy(config.BackupStrategy) {
		wg.Add(1)
		throttle <- 1
		go RestoreKey(backupKey, statistics, wg, throttle, etcdClient)
	}
}

func RestoreKey(backupKey *BackupKey, statistics map[string]*int32, wg *sync.WaitGroup, throttle chan int, etcdClient EtcdClient) {
	defer wg.Done()

	if !backupKey.IsExpired() {
		if backupKey.IsDirectory() {
			restoreKeyWithRetries(setDirectory, 0, backupKey, etcdClient)
			atomic.AddInt32(statistics["EmptyDirectories"], 1)
		} else {
			restoreKeyWithRetries(setKey, 0, backupKey, etcdClient)
		}

		atomic.AddInt32(statistics["KeysInserted"], 1)
	}

	<-throttle
}

func restoreKeyWithRetries(
	request func(*BackupKey, EtcdClient) (*etcd.Response, error),
	retries int, backupKey *BackupKey, etcdClient EtcdClient,
) {
	_, err := request(backupKey, etcdClient)
	if err != nil {
		if retries > config.Retries {
			config.LogFatal(err)
		} else {
			retries += 1
			time.Sleep(time.Duration(retries * 1000))
			restoreKeyWithRetries(request, retries, backupKey, etcdClient)
		}
	}
}

func setKey(backupKey *BackupKey, etcdClient EtcdClient) (*etcd.Response, error) {
	response, err := etcdClient.Set(backupKey.Key, *backupKey.Value, uint64(backupKey.TTL))
	if err != nil {
		err = errors.New("Error when trying to set the following key: " + backupKey.Key + ". Error: " + err.Error())
	}
	return response, err
}

func setDirectory(backupKey *BackupKey, etcdClient EtcdClient) (*etcd.Response, error) {
	response, err := etcdClient.SetDir(backupKey.Key, uint64(backupKey.TTL))
	if err != nil {
		if err.(*etcd.EtcdError) != nil && err.(*etcd.EtcdError).ErrorCode != 102 {
			err = errors.New("Error when trying to set the following directory : " + backupKey.Key + ". Error: " + err.Error())
		} else {
			err = nil
		}
	}

	return response, err
}
