package main

import (
	"testing"

	"github.com/coreos/go-etcd/etcd"
	"github.com/stretchr/testify/mock"
)

type MockedEtcdClient struct {
	mock.Mock
}

func (c *MockedEtcdClient) Get(key string, sort, recursive bool) (*etcd.Response, error) {
	args := c.Mock.Called(key, sort, recursive)
	return args.Get(0).(*etcd.Response), args.Error(1)
}

func (c *MockedEtcdClient) Set(key string, value string, ttl uint64) (*etcd.Response, error) {
	args := c.Mock.Called(key, value, ttl)
	return args.Get(0).(*etcd.Response), args.Error(1)
}

func (c *MockedEtcdClient) SetDir(key string, ttl uint64) (*etcd.Response, error) {
	args := c.Mock.Called(key, ttl)
	return args.Get(0).(*etcd.Response), args.Error(1)
}

func init() {
	config = &Config{
		ConcurrentRequests: 1,
		Retries:          1,
		EtcdConfigPath:   "none",
		DumpFilePath:     "fixtures/etcd-dump.json",
		BackupStrategy:   &BackupStrategy{[]string{"/"}, true, true},
	}

	failures = 0
	config.LogPrintln = func(v ...interface{}) {}
	config.LogFatal = func(v ...interface{}) { failures += 1 }
}

var failures int

func initTestClient() (*MockedEtcdClient, *etcd.Response) {
	etcdClientTest := new(MockedEtcdClient)
	node := etcd.Node{Key: "/test", Value: "testValue"}
	response := etcd.Response{Node: &node}

	return etcdClientTest, &response
}

func TestExecuteActionDump(t *testing.T) {
	etcdClientTest, response := initTestClient()

	etcdClientTest.On("Get", "/", true, true).Return(response, nil)
	ExecuteAction("dump", etcdClientTest)
	etcdClientTest.Mock.AssertExpectations(t)
}

func TestExecuteActionRestore(t *testing.T) {
	etcdClientTest, response := initTestClient()

	etcdClientTest.On("Set", "/test", "testValue", 0).Return(response, nil)
	ExecuteAction("restore", etcdClientTest)
	etcdClientTest.Mock.AssertExpectations(t)
}

func TestExecuteActionBreaking(t *testing.T) {
	failures = 0
	etcdClientTest, _ := initTestClient()

	ExecuteAction("break", etcdClientTest)
	if failures != 1 {
		t.Fatal("Action is not breaking!")
	}
}
