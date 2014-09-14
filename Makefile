.SUFFIXES:

.PHONY: all
all: build/dbuild build/dcontrol


build/dbuild: cmd/dbuild/*.go
	godep go build -o $@ $^

build/dcontrol: cmd/dcontrol/*.go
	godep go build -o $@ $^


.PHONY: test
test:
	godep go test -v cmd/dbuild/*.go
	godep go test -v cmd/dcontrol/*.go
