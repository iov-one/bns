.PHONY: all dist clean build install image test run swagger-docs

BUILD_VERSION ?= manual
BUILD_FLAGS := -mod=readonly -ldflags "-X main.buildHash=`git log -1 --pretty=format:"%H"` -X util.BuildVersion=${BUILD_VERSION}"
IMAGE_NAME = "iov1/bnsapi:${BUILD_VERSION}"
SWAG_BASE := docker run --rm -v $(shell pwd):/work iov1/swag-docker:latest
DOCKER_BUILD_FLAGS := -a -installsuffix cgo
BUILDOUT ?= bnsapi

all: dist

dist: clean test build image

clean:
	rm -f bnsapi 2> /dev/null

build:
	GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS) $(DOCKER_BUILD_FLAGS) -o $(BUILDOUT) .

install:
	go install $(BUILD_FLAGS) .

image:
	docker build --pull -t $(IMAGE_NAME) .

test:
	go test -mod=readonly -race ./...

run:
	go run .

swagger-docs:
	@cp -r ../../vendor .
	$(SWAG_BASE) swag init --parseVendor
	@rm -rf vendor
