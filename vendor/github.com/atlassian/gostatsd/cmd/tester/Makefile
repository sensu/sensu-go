VERSION_VAR := main.Version
GIT_VAR := main.GitCommit
BUILD_DATE_VAR := main.BuildDate
REPO_VERSION := $(shell git describe --abbrev=0 --tags)
BUILD_DATE := $(shell date +%Y-%m-%d-%H:%M)
GIT_HASH := $(shell git rev-parse --short HEAD)
GOBUILD_VERSION_ARGS := -ldflags "-s -X $(VERSION_VAR)=$(REPO_VERSION) -X $(GIT_VAR)=$(GIT_HASH) -X $(BUILD_DATE_VAR)=$(BUILD_DATE)"
BINARY_NAME := statsd-tester
IMAGE_NAME := atlassianlabs/$(BINARY_NAME)
ARCH ?= darwin
MAIN_PKG := github.com/atlassian/gostatsd/cmd/tester

setup:
	go get -u github.com/githubnemo/CompileDaemon
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install --update
	glide install

build: fmt
	go build -o $(BINARY_NAME) $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)

fmt:
	gofmt -w=true -s $(shell find . -type f -name '*.go' -not -path "./vendor/*")
	goimports -w=true -d $(shell find . -type f -name '*.go' -not -path "./vendor/*")

test:
	go test $(shell go list ./... | grep -v /vendor/)

cover:
	go test -covermode=count -coverprofile=coverage.out $(shell go list ./... | grep -v /vendor/)
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

junit-test: build
	go get github.com/jstemmer/go-junit-report
	go test -v ./... | go-junit-report > test-report.xml

check: fmt
	go install
	gometalinter --deadline=10s ./... --vendor --linter="errcheck::-ignore 'net:Close'"

watch:
	CompileDaemon -color=true -build "make test check"

commit-hook:
	cp dev/commit-hook.sh .git/hooks/pre-commit

cross: fmt
	CGO_ENABLED=0 GOOS=linux go build -o $(BINARY_NAME)-linux $(GOBUILD_VERSION_ARGS) -a -installsuffix cgo $(MAIN_PKG)

docker: cross
	docker build -t $(IMAGE_NAME):$(GIT_HASH) .

release: test docker
	docker push $(IMAGE_NAME):$(GIT_HASH)
	docker tag -f $(IMAGE_NAME):$(GIT_HASH) $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):latest
	docker tag -f $(IMAGE_NAME):$(GIT_HASH) $(IMAGE_NAME):$(REPO_VERSION)
	docker push $(IMAGE_NAME):$(REPO_VERSION)

run: build
	./$(BINARY_NAME) --metrics-addr=:8125

run-docker: cross
	docker-compose rm -f
	docker-compose -f docker-compose.yml build
	docker-compose -f docker-compose.yml up --force-recreate

version:
	@echo $(REPO_VERSION)

clean:
	rm -f $(IMAGE_NAME)*
	-docker rm $(docker ps -a -f 'status=exited' -q)
	-docker rmi $(docker images -f 'dangling=true' -q)

.PHONY: build
