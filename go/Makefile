.PHONY: clean prepare analyse build test-unit test-integration all

.DEFAULT_GOAL := build

DOCKER_IMAGE_TAG = latest

APIKEY := ${APIKEY}
HTTP_PORT := ${HTTP_PORT}
SERVICE_NAME := ${SERVICE_NAME}
TEST_PORT := ${TEST_PORT}

clean:
	echo Cleaning

prepare:
	echo Preparing goimports
	go get -u golang.org/x/tools/cmd/goimports
	echo Preparing golint
	go get -u golang.org/x/lint/golint
	echo Preparing modules
	go mod tidy
	go mod vendor
	echo Preparing statik
	go get -u github.com/rakyll/statik
	go build -o ~/go-workspace/bin/statik github.com/rakyll/statik/fs
	statik -f -src ../swaggerui -dest .

analyse: prepare
	echo Analysing
        # Use of sed here likely due to incorrect usage of modules.
	for dir in $$(go list ./... | sed -e 's#github.com/danieljamesrees/${SERVICE_NAME}/##' | sed -e 's#${SERVICE_NAME}/##'); do goimports -l -w "$${dir}"/*.go | tee /dev/stderr; done
	CGO_ENABLED=0 go vet ./...
	CGO_ENABLED=0 go fix ./...
	golint .
 
build: prepare
	echo Building
	CGO_ENABLED=0 GOOS=linux go install -a ./statik
	CGO_ENABLED=0 GOOS=linux go build -a -o ${GOPATH}/${SERVICE_NAME} -v ./internal/app

test-unit: build
	echo Unit testing packages serially
	CGO_ENABLED=0 go test -failfast -parallel=1 -tags=unit -v ./...

test-integration:
	echo Integration testing packages serially
	CGO_ENABLED=0 go test -failfast -parallel=1 -tags=integration -v ./...
