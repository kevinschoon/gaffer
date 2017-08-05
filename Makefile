DOCKER_IMAGE := mesanine/gaffer
SRCPATH := src/github.com/mesanine/gaffer
VERSION_PATH ?= "github.com/mesanine/gaffer/version"
GOPATH := $(shell echo $$GOPATH)
GITSHA ?= $(shell git rev-parse HEAD)
VERSION ?= $(shell git describe --tags 2>/dev/null)
PACKAGES ?= $(shell go list ./...|grep -v vendor | grep -v tests)
LDFLAGS ?= -w -s -X $(VERSION_PATH).Version=$(VERSION) -X $(VERSION_PATH).GitSHA=$(GITSHA)

.PHONY: all bindata dep docker protos test

all: protos bindata test build

ci: dep all

test:
	go $@ -v $(PACKAGES)
	go vet $(PACKAGES)

bindata:
	go-bindata -pkg server -o server/bindata.go www/...

dep:
	dep ensure

protos: 
	rm -v supervisor/*.pb.go 2>/dev/null || true
	rm -v host/*.pb.go 2>/dev/null || true
	rm -v service/*.pb.go 2>/dev/null || true
	protoc --proto_path=$(GOPATH)/src --go_out=plugins=grpc:$(GOPATH)/src $(GOPATH)/$(SRCPATH)/supervisor/*.proto
	protoc --proto_path=$(GOPATH)/src --go_out=$(GOPATH)/src $(GOPATH)/$(SRCPATH)/host/*.proto
	protoc --proto_path=$(GOPATH)/src --go_out=$(GOPATH)/src $(GOPATH)/$(SRCPATH)/service/*.proto

build:
	mkdir -v ./bin 2>/dev/null || true
	GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ./bin/gaffer

docker: 
	docker build -t $(DOCKER_IMAGE) .


