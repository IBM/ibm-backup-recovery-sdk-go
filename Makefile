# Makefile to build the project
GO=go
LINT=golangci-lint
GOSEC=gosec
TEST_TAGS=
COVERAGE = -coverprofile=coverage.txt -covermode=atomic
SCANOPTS=
MIGRATIONV2_DIR=migrationv2
EXAMPLES_OUT_DIR=$(MIGRATIONV2_DIR)/bin/examples

all: tidy test lint

travis-ci: tidy test-cov lint scan-gosec

test:
	${GO} test ./... ${TEST_TAGS}

test-cov:
	${GO} test ./... ${TEST_TAGS} ${COVERAGE}

test-int:
	${GO} test ./... -tags=integration

test-int-cov:
	${GO} test ./... -tags=integration ${COVERAGE}

lint:
	${LINT} run --build-tags=integration,examples --timeout 3m

scan-gosec:
	${GOSEC} ${SCANOPTS} ./...

tidy:
	${GO} mod tidy

migrationv2-deps:
	cd $(MIGRATIONV2_DIR) && $(GO) mod tidy

migrationv2-example-deps: migrationv2-deps
	cd $(MIGRATIONV2_DIR) && $(GO) mod download

migrationv2-examples-build: migrationv2-example-deps
	mkdir -p $(EXAMPLES_OUT_DIR)
	cd $(MIGRATIONV2_DIR) && $(GO) build -o bin/examples/connector-deployment ./examples/connector-deployment
	cd $(MIGRATIONV2_DIR) && $(GO) build -o bin/examples/microservice ./examples/microservice
	cd $(MIGRATIONV2_DIR) && $(GO) build -o bin/examples/migration-tool ./examples/migration-tool
	cd $(MIGRATIONV2_DIR) && $(GO) build -o bin/examples/policy-task-api ./examples/policy-task-api

migrationv2-examples-run-help:
	@echo "Build examples:"
	@echo "  make migrationv2-examples-build"
	@echo ""
	@echo "Run connector-deployment:"
	@echo "  ./migrationv2/bin/examples/connector-deployment"
	@echo ""
	@echo "Run microservice:"
	@echo "  ./migrationv2/bin/examples/microservice"
	@echo ""
	@echo "Run migration-tool:"
	@echo "  ./migrationv2/bin/examples/migration-tool start"
	@echo "  ./migrationv2/bin/examples/migration-tool status"
	@echo "  ./migrationv2/bin/examples/migration-tool monitor"
	@echo "  ./migrationv2/bin/examples/migration-tool list"
	@echo ""
	@echo "Run policy-task-api:"
	@echo "  ./migrationv2/bin/examples/policy-task-api"
	@echo ""
	@echo "Run directly with go:"
	@echo "  cd migrationv2 && go run ./examples/connector-deployment"
	@echo "  cd migrationv2 && go run ./examples/microservice"
	@echo "  cd migrationv2 && go run ./examples/migration-tool start"
	@echo "  cd migrationv2 && go run ./examples/policy-task-api"
