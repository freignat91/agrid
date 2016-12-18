.PHONY: all clean install fmt check version build run test

SHELL := /bin/sh
BASEDIR := $(shell echo $${PWD})

# build variables (provided to binaries by linker LDFLAGS below)
VERSION := 1.0.0
BUILD := $(shell git rev-parse HEAD | cut -c1-8)

LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# ignore vendor directory for go files
SRC := $(shell find . -type f -name '*.go' -not -path './vendor/*' -not -path './.git/*')

# for walking directory tree (like for proto rule)
DIRS = $(shell find . -type d -not -path '.' -not -path './vendor' -not -path './vendor/*' -not -path './.git' -not -path './.git/*')

# generated files that can be cleaned
GENERATED := $(shell find . -type f -name '*.pb.go' -not -path './vendor/*' -not -path './.git/*')

# ignore generated files when formatting/linting/vetting
CHECKSRC := $(shell find . -type f -name '*.go' -not -name '*.pb.go' -not -path './vendor/*' -not -path './.git/*')

OWNER := freignat91
NAME :=  agrid
TAG := latest

IMAGE := $(OWNER)/$(NAME):$(TAG)
REPO := github.com/$(OWNER)/$(NAME)

CLIENT := agrid
NODE := server

all: version check install

version: 
	@echo "version: $(VERSION) (build: $(BUILD))"

clean: 
	@rm -rf $(GENERATED)

install-client: 
	@go install $(LDFLAGS) $(REPO)/$(CLIENT)

install-node: 
	@go install $(LDFLAGS) $(REPO)/$(NODE)  

install: install-node install-client

proto: 
	@protoc server/gnode/gnode.proto --go_out=plugins=grpc:.

# format and simplify if possible (https://golang.org/cmd/gofmt/#hdr-The_simplify_command)
fmt:
	@gofmt -s -l -w $(CHECKSRC)

check:
	@test -z $(shell gofmt -l ${CHECKSRC} | tee /dev/stderr) || echo "[WARN] Fix formatting issues with 'make fmt'"
	@for d in $$(go list ./... | grep -v /vendor/); do golint $${d} | sed '/pb\.go/d'; done
	@go tool vet ${CHECKSRC}

build:	install-client
	@docker build -t $(IMAGE) .

run: 	build
        @CID=$(shell docker run --net=host -d --name $(NAME) $(IMAGE)) && echo $${CID}

install-deps:
	@glide install --strip-vcs --strip-vendor --update-vendored

update-deps:
	@glide update --strip-vcs --strip-vendor --update-vendored

test:
	@go test -v $(REPO)

start:
	@docker node inspect self > /dev/null 2>&1 || docker swarm inspect > /dev/null 2>&1 || (echo "> Initializing swarm" && docker swarm init --advertise-addr 127.0.0.1)
	@docker network ls | grep aNetwork || (echo "> Creating overlay network 'aNetwork'" && docker network create -d overlay aNetwork)
	@docker service create --network aNetwork --name agrid \
	--publish 30103:30103 \
	--mount type=bind,source=/home/freignat/data,target=/data \
	--replicas=5 \
	-e "NB_CONNECT=1" \
	$(IMAGE)
stop:
	@docker service rm agrid || true
init:
	@docker service rm agrid || true
	@rm -f ./leader.txt
	@echo "" > ./index.txt
	@rm -f ./logs/*
