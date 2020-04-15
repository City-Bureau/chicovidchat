.PHONY: install test format lint build deploy clean

cmd := $(shell find cmd -name \*main.go | awk -F'/' '{print $$2}')

install:
	npm install -g serverless
	npm install

test:
	go test ./...

format:
	test -z $$(gofmt -l .)

lint:
	golangci-lint run

build:
	@for c in $(cmd) ; do \
		env GOOS=linux go build -ldflags="-s -w" -o bin/$$c cmd/$$c/main.go ; \
	done

deploy:
	serverless deploy

clean:
	rm -rf bin
