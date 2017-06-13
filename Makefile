PWD := $(shell pwd)
DOCKER := docker run --rm -v $(PWD):/go/src/github.com/vektorlab/gaffer --workdir /go/src/github.com/vektorlab/gaffer quay.io/vektorcloud/go:dep
DOCKER_IMAGE := quay.io/vektorcloud/gaffer

.PHONY: all
all: docker

.PHONY: docker
docker:
	if [ ! -d ./bin ]; then \
		mkdir ./bin; \
	fi
	$(DOCKER) go-bindata -pkg server -o server/bindata.go www/...
	$(DOCKER) go build -o ./bin/gaffer
	docker build -t $(DOCKER_IMAGE):gaffer .
