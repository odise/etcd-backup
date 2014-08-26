package main

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

func TestIsDirectory(t *testing.T) {
	value := "test"
	backupKey := BackupKey{Value: &value}
	backupDir := BackupKey{}

	if backupKey.IsDirectory() != false || backupDir.IsDirectory() != true {
		t.Fatal("Unexpected value for IsDirectory(). Expected backupKey to be false and backupDir to be true.")
	}
}

func TestIsExpired(t *testing.T) {
	expiredTime := time.Now()
	unexpiredTime, _ := time.Parse("2006", strconv.Itoa(time.Now().Year()+100))
	expiredKey := BackupKey{Expiration: &expiredTime}
	unexpiredKey := BackupKey{Expiration: &unexpiredTime}
	notExpiringKey := BackupKey{}

	if (expiredKey.IsExpired() != true || expiredKey.TTL > 0) ||
		(unexpiredKey.IsExpired() != false || unexpiredKey.TTL < 10000) ||
		(notExpiringKey.IsExpired() != false || notExpiringKey.TTL != 0) {
		t.Fatal("Unexpected value for IsExpired(). Got:", expiredKey, unexpiredKey, notExpiringKey)
	}
}

func TestMatchBackupStrategy(t *testing.T) {
	key1 := BackupKey{Key: "/tests"}
	key2 := BackupKey{Key: "/tests/1"}
	key3 := BackupKey{Key: "/"}
	backupStrategy1 := &BackupStrategy{Keys: []string{"/"}, Recursive: false}
	backupStrategy2 := &BackupStrategy{Keys: []string{"/test"}, Recursive: false}
	backupStrategy3 := &BackupStrategy{Keys: []string{"/"}, Recursive: true}
	backupStrategy4 := &BackupStrategy{Keys: []string{"/none", "/tests"}, Recursive: true}

	evalStrategy(t, backupStrategy1, []bool{false, false, true}, key1, key2, key3)
	evalStrategy(t, backupStrategy2, []bool{false, false, false}, key1, key2, key3)
	evalStrategy(t, backupStrategy3, []bool{true, true, true}, key1, key2, key3)
	evalStrategy(t, backupStrategy4, []bool{true, true, false}, key1, key2, key3)
}

func evalStrategy(t *testing.T, backupStrategy *BackupStrategy, expectedResult []bool, keys ...BackupKey) {
	result := make([]bool, len(keys))
	for i, key := range keys {
		result[i] = key.MatchBackupStrategy(backupStrategy)
	}

	if reflect.DeepEqual(result, expectedResult) != true {
		t.Fatal("Unexpected result for backupStrategy:", fmt.Sprintf("%#v", backupStrategy), "expected:", expectedResult, ".Got: ", result)
	}
}

func TestDownloadDataSet(t *testing.T) {
	emptySet := mockEmptyDataSet(t, BackupStrategy{})
	if len(emptySet) > 0 {
		t.Fatal("Unexpected value for DownloadDataSet. expected empty set got:", emptySet)
	}

	set := mockDataSet(t, BackupStrategy{Keys: []string{"/"}, Sorted: true, Recursive: true})
	if len(set) == 0 {
		t.Fatal("Unexpected value for DownloadDataSet. Expected not empty set got:", set)
	}
}

func mockEmptyDataSet(t *testing.T, backupStrategy BackupStrategy) []*BackupKey {
	etcdClientTest, _ := initTestClient()

	emptySet := DownloadDataSet(&backupStrategy, etcdClientTest)
	etcdClientTest.Mock.AssertNotCalled(t, "Get", "/", true, true)

	return emptySet
}

func mockDataSet(t *testing.T, backupStrategy BackupStrategy) []*BackupKey {
	etcdClientTest, response := initTestClient()

	etcdClientTest.On("Get", "/", true, true).Return(response, nil)
	emptySet := DownloadDataSet(&backupStrategy, etcdClientTest)
	etcdClientTest.Mock.AssertExpectations(t)

	return emptySet
}

func TestExtractNodes(t *testing.T) {
	emptyNode := etcd.Node{}
	etcdNode1 := etcd.Node{Key: "test"}
	etcdNode2 := etcd.Node{Key: "Other_test"}
	complexNode := etcd.Node{Key: "/", Nodes: []*etcd.Node{&etcdNode1, &etcdNode2}}

	emptyKeys := extractNodes(&emptyNode, &BackupStrategy{Keys: []string{"/"}, Sorted: true, Recursive: true})
	oneKey := extractNodes(&complexNode, &BackupStrategy{})
	twoKeys := extractNodes(&complexNode, &BackupStrategy{Keys: []string{"/"}, Sorted: true, Recursive: true})

	if len(emptyKeys) != 0 || len(oneKey) != 1 || len(twoKeys) != 2 {
		t.Fatal("Unexpected value for extractNodes. Got:", emptyKeys, oneKey, twoKeys)
	}
}

func TestSingleNodeToBackupKey(t *testing.T) {
	time := time.Now()
	emptyNode := etcd.Node{}
	value := "test"
	keyNode := etcd.Node{Key: "key", Value: value, Expiration: &time}
	dirNode := etcd.Node{Key: "dir", Dir: true}

	BackupKeyValid(t, &emptyNode, &BackupKey{Key: ""})
	BackupKeyValid(t, &keyNode, &BackupKey{Key: "key", Value: &value, Expiration: &time})
	BackupKeyValid(t, &dirNode, &BackupKey{Key: "dir"})
}

func BackupKeyValid(t *testing.T, node *etcd.Node, expectedKey *BackupKey) {
	backupKey := SingleNodeToBackupKey(node)

	if reflect.DeepEqual(backupKey, expectedKey) != true {
		t.Fatal("Unexpected result:", fmt.Sprintf("%#v", backupKey), "expected:", fmt.Sprintf("%#v", expectedKey))
	}
}
