.PHONY: build install generate vet compile test doc fmt lint docker

default: test compile

compile:
	go build -v -o ./proto2gql ./*.go

test:
	go test ./*.go
