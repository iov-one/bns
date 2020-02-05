BUILD_VERSION ?= manual
BUILD_FLAGS := -mod=readonly -ldflags "-X main.buildHash=`git log -1 --pretty=format:"%H"` -X util.BuildVersion=${BUILD_VERSION}"
IMAGE_NAME = "iov1/bnsapi:${BUILD_VERSION}"

all: dist

dist: clean test build image

clean:
	rm -f bnsapi 2> /dev/null

build:
	GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS) -o bnsapi .

gen-docs:
	@go mod vendor
	docker run --rm -it -v $(shell pwd):/go/src/github.com/iov-one/weave -w="/go/src/github.com/iov-one/weave" golangci/golangci-lint:v1.17.1 golangci-lint run ./...
	@rm -rf vendor

install:
	go install $(BUILD_FLAGS) .

image:
	docker build --pull -t $(IMAGE_NAME) .

test:
	go test -mod=readonly -race ./...

run:
	go run .
