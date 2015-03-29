test: install
	go test -v ./...

get-deps:
	go get github.com/golang/protofub/proto
	go get gopkg.in/vmihailenco/msgpack.v2

install: get-deps

