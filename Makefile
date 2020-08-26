
init:
	go mod download

build:
	go build

test:
	go test ./... -count=1

format:
	gofmt -w ./

lint:
	gofmt -d ./
	test -z $(shell gofmt -l ./)

verify:
	go mod verify

tidy:
	go mod tidy
