TEST_OPTS=-v -test.timeout 300s

help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_0-9-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

tidy: ## apply go mod tidy
	test ${CI} || go mod tidy

fmt: ## apply go format
	gofmt -s -w ./

test: tidy cmd pkg ## launch tests
	test -z "`gofmt -d . | tee /dev/stderr`"
	go test $(TEST_OPTS) ./...
