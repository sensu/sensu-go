VERSION_VAR := main.Version
GIT_VAR := main.GitCommit
BUILD_DATE_VAR := main.BuildDate
REPO_VERSION := $$(git describe --abbrev=0 --tags)
BUILD_DATE := $$(date +%Y-%m-%d-%H:%M)
GIT_HASH := $$(git rev-parse --short HEAD)
GOBUILD_VERSION_ARGS := -ldflags "-s -X $(VERSION_VAR)=$(REPO_VERSION) -X $(GIT_VAR)=$(GIT_HASH) -X $(BUILD_DATE_VAR)=$(BUILD_DATE)"
GOBUILD_VERSION_ARGS_WITH_SYMS := -ldflags "-X $(VERSION_VAR)=$(REPO_VERSION) -X $(GIT_VAR)=$(GIT_HASH) -X $(BUILD_DATE_VAR)=$(BUILD_DATE)"
BINARY_NAME := gostatsd
IMAGE_NAME := atlassianlabs/$(BINARY_NAME)
ARCH ?= $$(uname -s | tr A-Z a-z)
METALINTER_CONCURRENCY ?= 4
GOVERSION := 1.9
GP := /gopath
MAIN_PKG := github.com/atlassian/gostatsd/cmd/gostatsd
CLUSTER_PKG := github.com/atlassian/gostatsd/cmd/cluster

setup: setup-ci
	go get -u github.com/githubnemo/CompileDaemon
	go get -u github.com/jstemmer/go-junit-report
	go get -u golang.org/x/tools/cmd/goimports

setup-ci:
	go get -u github.com/Masterminds/glide
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
	glide install --strip-vendor

build-cluster: fmt
	go build -i -v -o build/bin/$(ARCH)/cluster $(GOBUILD_VERSION_ARGS) $(CLUSTER_PKG)

build: fmt
	go build -i -v -o build/bin/$(ARCH)/$(BINARY_NAME) $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)

build-race: fmt
	go build -i -v -race -o build/bin/$(ARCH)/$(BINARY_NAME) $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)

build-all:
	go install -v $$(glide nv)

fmt:
	gofmt -w=true -s $$(find . -type f -name '*.go' -not -path "./vendor/*")
	goimports -w=true -d $$(find . -type f -name '*.go' -not -path "./vendor/*")

test:
	go test $$(glide nv)

test-race:
	go test -race $$(glide nv)

bench:
	go test -bench=. -run=XXX $$(glide nv)

bench-race:
	go test -race -bench=. -run=XXX $$(glide nv)

cover:
	./cover.sh
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

coveralls:
	./cover.sh
	goveralls -coverprofile=coverage.out -service=travis-ci

junit-test: build
	go test -v $$(glide nv) | go-junit-report > test-report.xml

check:
	go install ./cmd/gostatsd
	go install ./cmd/tester
	gometalinter --concurrency=$(METALINTER_CONCURRENCY) --deadline=600s ./... --vendor \
		--linter='errcheck:errcheck:-ignore=net:Close' --cyclo-over=20 \
		--disable=interfacer --disable=golint --dupl-threshold=200

check-all:
	go install ./cmd/gostatsd
	go install ./cmd/tester
	gometalinter --concurrency=$(METALINTER_CONCURRENCY) --deadline=600s ./... --vendor --cyclo-over=20 \
		--dupl-threshold=65

fuzz-setup:
	go get -v -u github.com/dvyukov/go-fuzz/go-fuzz
	go get -v -u github.com/dvyukov/go-fuzz/go-fuzz-build

fuzz:
	go-fuzz-build github.com/atlassian/gostatsd/pkg/statsd
	go-fuzz -bin=./statsd-fuzz.zip -workdir=test_fixtures/lexer_fuzz

watch:
	CompileDaemon -color=true -build "make test"

git-hook:
	cp dev/push-hook.sh .git/hooks/pre-push

# Compile a static binary. Cannot be used with -race
docker:
	docker pull golang:$(GOVERSION)
	docker run \
		--rm \
		-v "$(GOPATH)":"$(GP)" \
		-w "$(GP)/src/github.com/atlassian/gostatsd" \
		-e GOPATH="$(GP)" \
		-e CGO_ENABLED=0 \
		golang:$(GOVERSION) \
		go build -o build/bin/linux/$(BINARY_NAME) $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)
	docker build --pull -t $(IMAGE_NAME):$(GIT_HASH) build

# Compile a binary with -race. Needs to be run on a glibc-based system.
docker-race:
	docker pull golang:$(GOVERSION)
	docker run \
		--rm \
		-v "$(GOPATH)":"$(GP)" \
		-w "$(GP)/src/github.com/atlassian/gostatsd" \
		-e GOPATH="$(GP)" \
		golang:$(GOVERSION) \
		go build -race -o build/bin/linux/$(BINARY_NAME) $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)
	docker build --pull -t $(IMAGE_NAME):$(GIT_HASH)-race -f build/Dockerfile-glibc build

# Compile a static binary with symbols. Cannot be used with -race
docker-symbols:
	docker pull golang:$(GOVERSION)
	docker run \
		--rm \
		-v "$(GOPATH)":"$(GP)" \
		-w "$(GP)/src/github.com/atlassian/gostatsd" \
		-e GOPATH="$(GP)" \
		-e CGO_ENABLED=0 \
		golang:$(GOVERSION) \
		go build -o build/bin/linux/$(BINARY_NAME) $(GOBUILD_VERSION_ARGS_WITH_SYMS) $(MAIN_PKG)
	docker build --pull -t $(IMAGE_NAME):$(GIT_HASH)-syms build

release-hash: docker
	docker push $(IMAGE_NAME):$(GIT_HASH)

release-normal: release-hash
	docker tag $(IMAGE_NAME):$(GIT_HASH) $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):latest
	docker tag $(IMAGE_NAME):$(GIT_HASH) $(IMAGE_NAME):$(REPO_VERSION)
	docker push $(IMAGE_NAME):$(REPO_VERSION)

release-hash-race: docker-race
	docker push $(IMAGE_NAME):$(GIT_HASH)-race

release-race: docker-race
	docker tag $(IMAGE_NAME):$(GIT_HASH)-race $(IMAGE_NAME):$(REPO_VERSION)-race
	docker push $(IMAGE_NAME):$(REPO_VERSION)-race

release-symbols: docker-symbols
	docker tag $(IMAGE_NAME):$(GIT_HASH)-syms $(IMAGE_NAME):$(REPO_VERSION)-syms
	docker push $(IMAGE_NAME):$(REPO_VERSION)-syms

release: release-normal release-race release-symbols

run: build
	./build/bin/$(ARCH)/$(BINARY_NAME) --backends=stdout --verbose --flush-interval=2s

run-docker: docker
	cd build/ && docker-compose rm -f gostatsd
	docker-compose -f build/docker-compose.yml build
	docker-compose -f build/docker-compose.yml up -d

stop-docker:
	cd build/ && docker-compose stop

version:
	@echo $(REPO_VERSION)

clean:
	rm -f build/bin/*
	-docker rm $(docker ps -a -f 'status=exited' -q)
	-docker rmi $(docker images -f 'dangling=true' -q)

.PHONY: build
