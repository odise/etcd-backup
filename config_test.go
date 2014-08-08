package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestConfigToString(t *testing.T) {
	newConfig := Config{}
	str := newConfig.ToString()
	expectedStr := "ConcurrentRequests: 0, Retries: 0, EtcdConfigPath: , DumpFilePath: , BackupStrategy: (*main.BackupStrategy)(nil)"
	stringCompare(t, str, expectedStr)

	OtherExpectedStr := "ConcurrentRequests: 1, Retries: 1, EtcdConfigPath: none, DumpFilePath: fixtures/etcd-dump.json, "
	OtherExpectedStr += "BackupStrategy: &main.BackupStrategy{Keys:[]string{\"/\"}, Sorted:true, Recursive:true}"
	stringCompare(t, config.ToString(), OtherExpectedStr)
}

func stringCompare(t *testing.T, str, expectedStr string) {
	if str != expectedStr {
		t.Fatal("Unexpected result:", str, "expected:", expectedStr)
	}
}

func TestConfigNilValueOverride(t *testing.T) {
	newConfig := &Config{}
	configNilValueOverride(newConfig, config)
	configCompare(t, *newConfig, *config, true)

	otherConfig := &Config{
		ConcurrentRequests: 2,
		Retries:          2,
		EtcdConfigPath:   "none",
		DumpFilePath:     "none",
		BackupStrategy:   &BackupStrategy{[]string{"/"}, false, false},
	}
	configNilValueOverride(otherConfig, config)
	configCompare(t, *otherConfig, *config, false)
}

func configCompare(t *testing.T, config, expectedConfig Config, equal bool) {
	// Deep equal not working on func :/...
	config.LogFatal = nil
	expectedConfig.LogFatal = nil
	config.LogPrintln = nil
	expectedConfig.LogPrintln = nil

	if reflect.DeepEqual(config, expectedConfig) != equal {
		t.Fatal("Unexpected result:", fmt.Sprintf("%#v", config), "expected:", fmt.Sprintf("%#v", expectedConfig))
	}
}

func TestLoadConfigFile(t *testing.T) {
	loadUnexistingConfigFile(t)
	loadbadConfigFile(t)
	loadValidConfigFile(t)
}

func loadUnexistingConfigFile(t *testing.T) {
	failures = 0
	name := "unexistingFile"
	loadConfigFile(&name)

	if failures == 0 {
		t.Fatal("No failure raised when config file does not exist.")
	}
}

func loadbadConfigFile(t *testing.T) {
	failures = 0
	name := "fixtures/etcd-dump.json"
	loadConfigFile(&name)

	if failures == 0 {
		t.Fatal("No failure raised when config file is bad.")
	}
}

func loadValidConfigFile(t *testing.T) {
	failures = 0
	name := "fixtures/backup-configuration.json"
	newConfig := loadConfigFile(&name)
	backupStrategy := BackupStrategy{Keys: []string{"/keys/all"}, Sorted: false, Recursive: false}
	expectedConfig := Config{ConcurrentRequests: 50, Retries: 5, EtcdConfigPath: "fixtures/etcd-dump.json", DumpFilePath: "dump.json", BackupStrategy: &backupStrategy}

	if failures != 0 {
		t.Fatal("No failure raised when config file is bads.")
	}

	if reflect.DeepEqual(*newConfig, expectedConfig) == false {
		fmt.Println(fmt.Sprintf("%#v", newConfig), fmt.Sprintf("%#v", newConfig.BackupStrategy))
		t.Fatal("Unexpected configuration from file. Got:", newConfig, "expected:", expectedConfig)
	}
}
