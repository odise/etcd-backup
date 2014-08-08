# etcd-backup

etcd-backup is a simple, efficient and lightweight command line utility to backup and restore [etcd](https://github.com/coreos/etcd) keys.

## Dependencies

etcd-backup has only one dependency: go-etcd [the golang offical library for ETCD](https://github.com/coreos/go-etcd)

## Installation

  Installation composed of 3 steps:

* [Install go](http://golang.org/doc/install/source)
* Download the project `git clone git@github.com:fanhattan/etcd-backup.git`
* Download the dependency `go get github.com/coreos/go-etcd/etcd`
* Build the binary `cd etcd-backup` and then  `go install`

## Dumping

### Usage

    $ etcd-dump dump

This is the easiest way to dump the whole `etcd` keyspace. Results will be stored in a json file `etcd-dump.json`
in the directory where you executed the command.

The default Backup strategy for dumping is to dump all keys and preserve the order : `keys:["/"], recursive:true, sorted:true`
The backup strategy can be overwritten in the etcd-backup configuration file. See _fixtures/backup-configuration.json_

### Command line options and default values

  `-config` Mandatory etcd-backup configuration file location, default value: "_backup-configuration.json_". See [Configuration section](#config) for more information.<br/>
  `-retries` Number of retries that will be executed if the request fails, default value is 5.<br/>
  `-etcd-config` Mandatory etcd configuration file location, default value: "_etcd-configuration.json_". See fixtures folder for an example. **????????????????** <br/>
  `-file` Location of the dump file data will be stored in, default value: "_etcd-dump.json_".<br/>


    $ etcd-dump -config=myBackupConfig.json -retries=2 -etcd-config=myClusterConfig.json -file=result.json dump

### <a name="config"/>Configuration

The `dump.keys` supports different configurations:

  {
    "key": "/",
    "recursive": true
  }

Recursively dump all the keys inside `/`.

  {
    "key": "/myKey"
  }

Dump only the key `/myKey`.


### Dump File structure

Dumped keys are stored in an array of keys, the key path is the absolute path. By design non-empty directories are not saved in the dump file, and empty directories do not contain the `value` key:

    [{ "key": "/myKey", "value": "value1" },{ "key": "/dir/mydir/myKey", "value": "test" }, {"key": "/dir/emptyDir"}]

## Restoring

### Usage

    $ etcd-dump restore

Restore the keys from the `etcd-dump.json` file.

### Command line options and default values

  `-config` Mandatory etcd-backup configuration file location, default value: "_backup-configuration.json_". See [Configuration section](#config) for more information.<br/>
  `-concurrent-requests` Number of concurrent requests that will be executed during the restore (restore mode only), default value is 10.<br/>
  `-retries` Number of retries that will be executed if the request fails, default value is 5.<br/>
  `-etcd-config` Mandatory etcd configuration file location, default value: "_etcd-configuration.json_". See fixtures folder for an example. **????????????????** <br/>
  `-file` Location of the dump file data will be loaded from, default value: "_etcd-dump.json_".<br/>

    $ etcd-dump -config=myBackupConfig.json -retries=2 -etcd-config=myClusterConfig.json -file=dataset.json -concurrent-requests=100 restore

