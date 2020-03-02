
FILES = $(shell find . -type f -name '*.go')

gofmt:
	@gofmt -s -w $(FILES)
	@gofmt -r '&a{} -> new(a)' -w $(FILES)

test:
	@mkdir -p tmp
	go install .
	protoc --grpc_browser_out=tmp -I ~/projects/googleapis -I. ./testdata/example/example.proto
	@echo '--- output ---'
	@cat tmp/testdata/example/example.js
	@echo '--- output ---'

update-deps:
	go get -u
	go mod download
	go mod tidy
