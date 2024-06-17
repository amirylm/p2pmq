APP_NAME?=p2pmq
BUILD_TARGET?=${APP_NAME}
BUILD_IMG?=${APP_NAME}
APP_VERSION?=$(git describe --tags $(git rev-list --tags --max-count=1) 2> /dev/null || echo "nightly")
CFG_PATH?=./resources/config/default.p2pmq.yaml
TEST_PKG?=./core/...
TEST_TIMEOUT?=2m

protoc:
	./scripts/proto-gen.sh

lint:
	@docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.54 golangci-lint run -v --timeout=5m ./...

fmt:
	@go fmt ./...

test:
	@go test -v -race -timeout=${TEST_TIMEOUT} `go list ./... | grep -v -E "cmd|scripts|resources|examples|proto"`

test-localdon:
	@make TEST_PKG=./examples/don/... test-pkg

test-bls:
	@go test -v -tags blst_enabled -timeout=${TEST_TIMEOUT} ./examples/bls/...

test-pkg:
	@go test -v -race -timeout=${TEST_TIMEOUT} ${TEST_PKG}

test-cov:
	@go test -timeout=${TEST_TIMEOUT} -coverprofile cover.out `go list ./... | grep -v -E "cmd|scripts|resources|examples|proto"`

test-open-cov:
	@make test-cov
	@go tool cover -html cover.out -o cover.html
	open cover.html

keygen:
	@go run ./cmd/keygen/main.go

build:
	@go build -o "./bin/${BUILD_TARGET}" "./cmd/${BUILD_TARGET}"

docker-build:
	@docker build -t "${APP_NAME}" --build-arg APP_VERSION="${APP_VERSION}" --build-arg APP_NAME="${APP_NAME}" --build-arg BUILD_TARGET="${BUILD_TARGET}" .

docker-run-default:
	@docker run -d --restart unless-stopped --name "${APP_NAME}" -p "${TCP_PORT}":"${TCP_PORT}" -p "${GRPC_PORT}":"${GRPC_PORT}" -e "GRPC_PORT=${GRPC_PORT}" -it "${BUILD_IMG}" /p2pmq/app -config=./default.p2pmq.yaml

docker-run-boot:
	@docker run -d --restart unless-stopped --name "${APP_NAME}" -p "${TCP_PORT}":"${TCP_PORT}" -p "${GRPC_PORT}":"${GRPC_PORT}" -e "GRPC_PORT=${GRPC_PORT}" -it "${BUILD_IMG}" /p2pmq/app -config=./bootstrapper.p2pmq.yaml
