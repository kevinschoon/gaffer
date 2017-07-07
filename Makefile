PWD := $(shell pwd)
GOPATH := $(shell echo $$GOPATH)
DOCKER_IMAGE := quay.io/vektorlab/gaffer
DOCKER := docker run --rm -v $(PWD):/go/src/github.com/vektorlab/gaffer --workdir /go/src/github.com/vektorlab/gaffer quay.io/vektorcloud/go:dep

.PHONY: all
all: docker

.PHONY: protos
protos: 
	rm -v supervisor/*.pb.go 2>/dev/null || true
	rm -v host/*.pb.go 2>/dev/null || true
	rm -v service/*.pb.go 2>/dev/null || true
	protoc --proto_path=$(GOPATH)/src --go_out=plugins=grpc:$(GOPATH)/src $(PWD)/supervisor/*.proto
	protoc --proto_path=$(GOPATH)/src --go_out=$(GOPATH)/src $(PWD)/host/*.proto
	protoc --proto_path=$(GOPATH)/src --go_out=$(GOPATH)/src $(PWD)/service/*.proto

.PHONY: docker
docker:
	if [ ! -d ./bin ]; then \
		mkdir ./bin; \
	fi
	$(DOCKER) go-bindata -pkg server -o server/bindata.go www/...
	$(DOCKER) go build -o ./bin/gaffer
	docker build -t $(DOCKER_IMAGE) .
