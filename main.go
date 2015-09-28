package main

import (
	"os"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

var (
	etcdClient *etcd.Client
)

func main() {
	args := os.Args
	startTime := time.Now()
	backupModeAction := args[len(args)-1]
	etcdClient = initializeClient(config.EtcdConfigPath)

	ExecuteAction(backupModeAction, etcdClient)
	config.LogPrintln("Done!", backupModeAction, "executed in", time.Now().Sub(startTime))
}

func initializeClient(configFilePath string) *etcd.Client {
	etcdClient, error := etcd.NewClientFromFile(configFilePath)
	if error != nil {
		config.LogFatal("Error when trying to load the configuration file: `"+configFilePath+"`. Error: ", error)
	}

	if config.SyncCluster {
		success := etcdClient.SyncCluster()
		if !success {
			config.LogFatal("cannot sync machines")
		}
	}

	return etcdClient
}

func ExecuteAction(action string, etcdClient EtcdClient) {
	if action == "restore" {
		dataSetLoad := LoadDataSet(config.DumpFilePath)

		config.LogPrintln("Restoring dataSet in progress...")
		RestoreDataSet(*dataSetLoad, config, etcdClient)
	} else if action == "dump" {
		dataSet := DownloadDataSet(config.BackupStrategy, etcdClient)
		DumpDataSet(dataSet, config.DumpFilePath)
	} else {
		config.LogFatal("Error no default mode found. Got `" + action + "`. Try `restore` to restore or `dump` to persist data")
	}
}
