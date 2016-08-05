GO ?= go
GOPATH := $(CURDIR)/../../../..
PACKAGES := $(shell GOPATH=$(GOPATH) go list ./... | grep -v /vendor/)

all: install

build: install_deps
	GOPATH=$(GOPATH) $(GO) build -ldflags "-X main.version=`cat VERSION`"

run: install_deps
	GOPATH=$(GOPATH) $(GO) run -ldflags "-X main.version=`cat VERSION`" `ls *.go | grep -v _test.go` -host analytics.wywy.com

test: install_deps
	GOPATH=$(GOPATH) $(GO) test -cover $(PACKAGES)
	GOPATH=$(GOPATH) $(GO) vet $(PACKAGES)

coverage.out:
	rm -f coverage.*.out
	for i in .; do GOPATH=$(GOPATH) $(GO) test -coverprofile=coverage.$$i.out -covermode=count ./$$i; done
	echo "mode: count" > coverage.out
	grep -v -h "mode: count" coverage.*.out >> coverage.out
	rm -f coverage.*.out

coverage: coverage.out
	GOPATH=$(GOPATH) $(GO) tool cover -html=coverage.out -o coverage.html
	rm -f coverage.out

fmt:
	GOPATH=$(GOPATH) find . -name "*.go" | xargs gofmt -w

install_deps: install_glide
	GOPATH=$(GOPATH) $(GOPATH)/bin/glide install

install_glide:
	GOPATH=$(GOPATH) $(GO) get github.com/Masterminds/glide