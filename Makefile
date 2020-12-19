.PHONY: update fmt build integrationtest goose tunnel
.DEFAULT_GOAL := build

update:
	go get -u -f

fmt:
	go fmt ./...

build: fmt
	go build -ldflags="-s -w" kulloserver.go

integrationtest:
	python -m tests

goose:
	go get -u bitbucket.org/liamstask/goose/cmd/goose

tunnel:
	ssh -L 15432:localhost:5432 root@kullo.kullo.net -N
