
install:
		brew install dep && dep ensure

dist: build
		@dep ensure

update: build
		@dep ensure -update

build:
		@go build ./...

test:
		@go test -v -race ./...

release:
		@goreleaser --rm-dist