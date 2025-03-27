COVEROUT ?= coverage.html

.PHONY: test build

test:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o ${COVEROUT}

build:
	go build -o silent-score app/main.go

run: build
	./silent-score
