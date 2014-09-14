.SUFFIXES:

.PHONY: all
all: build/dbuild build/dcontrol


build/dbuild: cmd/dbuild/*.go
	godep go build -o $@ $^

build/dcontrol: cmd/dcontrol/*.go
	godep go build -o $@ $^

.PHONY: test
test:
	cd cmd && godep go test -short -v ./...

.PHONY: test-all
test-all:
	cd cmd && godep go test -v ./...
