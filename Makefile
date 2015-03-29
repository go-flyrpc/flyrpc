test: install
	go test -v ./...

get-deps:
	go get -t ./...

install: get-deps

