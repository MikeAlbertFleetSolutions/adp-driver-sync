package := $(shell basename `pwd`)

# strip debugging information & include buildnumber in executable
LDFLAGS=-ldflags "-s -w -X main.buildnum=${BUILD_NUM}"

.PHONY: default get codetest build fmt lint vet vuln run

default: fmt codetest

get:
ifneq ("$(CI)", "true")
	go get -u ./...
	go mod tidy
endif
	go mod download
	go mod verify

codetest: lint vet vuln

build: default
	mkdir -p target
	rm -f target/*
	GOOS=linux GOARCH=amd64 go build -tags=nomsgpack ${LDFLAGS} -v -o target/$(package) github.com/MikeAlbertFleetSolutions/adp-driver-sync/cmd/adp-driver-sync

fmt:
	go fmt ./...

lint:
ifeq ("$(CI)", "true")
	$(shell go env GOPATH)/bin/golangci-lint run --verbose --timeout 3m
else
	$(shell go env GOPATH)/bin/golangci-lint run --fix
endif

vet:
	go vet -all ./...

vuln:
	govulncheck -test ./...

run: build
	target/adp-driver-sync -config adp-driver-sync.yaml
