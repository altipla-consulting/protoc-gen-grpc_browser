
FILES = $(shell find . -type f -name '*.go')

gofmt:
	@gofmt -s -w $(FILES)
	@gofmt -r '&a{} -> new(a)' -w $(FILES)

build:
	@mkdir -p tmp
	cd tmp && git clone git@github.com:googleapis/googleapis.git
	npm ci

test:
	go install .
	protoc --grpc_browser_out=tmp -I tmp/googleapis -I. ./testdata/example/example.proto
	@echo '--- output ---'
	@cat tmp/testdata/example/example.js
	@echo '--- output ---'

update-deps:
	go get -u
	go mod download
	go mod tidy
