
GOBIN          =$(PWD)/_bin
GO             = go
OS             =$(shell go env var GOOS | xargs)

.PHONY: build-common
build-common: ## - execute build common tasks clean and mod tidy
	@ go version
	@ go clean
	@ go mod tidy && go mod download
	@ go mod verify

build: build-common ## - build a debug binary to the current platform (windows, linux or darwin(mac))
	@ echo cleaning...
	@ rm -f $(GOBIN)/debug/$(OS)/smelter
	@ echo building...
	@ go build -tags dev -o "_bin/debug/$(OS)/smelter" cmd/*.go
	@ ls -lah $(GOBIN)/debug/$(OS)/smelter


.PHONY: test
test: ## - execute go test command
	@ go test -v -cover `go list ./... | grep -v github.com/rahul0tripathi/smelter/vm`