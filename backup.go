package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

type BackupKey struct {
	Key        string     `json:"key"`
	Value      *string    `json:"value,omitempty"`
	Expiration *time.Time `json:"expiration,omitempty"`
	TTL        int64      `json:"ttl,omitempty"`
}

type EtcdClient interface {
	Get(key string, sort, recursive bool) (*etcd.Response, error)
	Set(key string, value string, ttl uint64) (*etcd.Response, error)
	SetDir(key string, ttl uint64) (*etcd.Response, error)
}

func (bacupKey *BackupKey) IsDirectory() (isDirectory bool) {
	if bacupKey.Value == nil {
		isDirectory = true
	}

	return isDirectory
}

func (bacupKey *BackupKey) IsExpired() (isExpired bool) {
	if bacupKey.Expiration != nil {
		bacupKey.TTL = int64(bacupKey.Expiration.Sub(time.Now().UTC()))
		isExpired = bacupKey.TTL <= 0
	}

	return isExpired
}

func (bacupKey *BackupKey) MatchBackupStrategy(backupStrategy *BackupStrategy) (match bool) {
	for _, key := range backupStrategy.Keys {
		if (backupStrategy.Recursive == true && strings.HasPrefix(bacupKey.Key, key)) || bacupKey.Key == key {
			match = true
			break
		}
	}

	return match
}

func DownloadDataSet(backupStrategy *BackupStrategy, etcdClient EtcdClient) []*BackupKey {
	keysToPersist := make([]*BackupKey, 0)

	for _, key := range backupStrategy.Keys {
		response, err := etcdClient.Get(key, backupStrategy.Sorted, backupStrategy.Recursive)
		if err != nil {
			config.LogFatal("Error when trying to get the following key: "+key+". Error: ", err)
		}

		keysToPersist = append(keysToPersist, extractNodes(response.Node, backupStrategy)...)
		config.LogPrintln("Total number of key persisted:", fmt.Sprintf("%#v", len(keysToPersist)))
	}

	return keysToPersist
}

func extractNodes(node *etcd.Node, backupStrategy *BackupStrategy) []*BackupKey {
	backupKeys := make([]*BackupKey, 0)

	if backupStrategy.Recursive == true {
		backupKeys = NodesToBackupKeys(node)
	} else {
		backupKeys = append(backupKeys, SingleNodeToBackupKey(node))
	}

	return backupKeys
}

func SingleNodeToBackupKey(node *etcd.Node) *BackupKey {
	key := BackupKey{
		Key:        node.Key,
		Expiration: node.Expiration,
	}

	if node.Dir != true && node.Key != "" {
		key.Value = &node.Value
	}

	return &key
}

func NodesToBackupKeys(node *etcd.Node) []*BackupKey {
	backupKeys := make([]*BackupKey, 0)

	if len(node.Nodes) > 0 {
		for _, nodeChild := range node.Nodes {
			backupKeys = append(backupKeys, NodesToBackupKeys(nodeChild)...)
		}
	} else {
		backupKey := SingleNodeToBackupKey(node)
		if backupKey.Key != "" {
			backupKeys = append(backupKeys, backupKey)
		}
	}

	return backupKeys
}

func DumpDataSet(dataSet []*BackupKey, dumpFilePath string) {
	jsonDataSet, err := json.MarshalIndent(dataSet, "", "  ")
	if err != nil {
		config.LogFatal("Error when trying to encode data set into json. Error: ", err)
	}

	file, error := os.OpenFile(dumpFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if error != nil {
		config.LogFatal("Error when trying to open the file `"+dumpFilePath+"`. Error: ", error)
	}

	_, err = file.Write(jsonDataSet)
	if error != nil {
		config.LogFatal("Error when writing dump file to disk the file `"+dumpFilePath+"`. Error: ", error)
	}
}
