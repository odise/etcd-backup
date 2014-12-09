build:
	mkdir -p out/darwin out/linux
	GOOS=darwin go build -o out/darwin/etcd-backup
	GOOS=linux go build -o out/linux/etcd-backup

clean:
	rm -rf out

test:
	go test -cover
