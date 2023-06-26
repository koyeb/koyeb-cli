TEST_OPTS=-v -test.timeout 300s

help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_0-9-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

tidy: ## apply go mod tidy
	test ${CI} || go mod tidy

install: ## install
	go install cmd/koyeb/koyeb.go

fmt: ## apply go format
	gofmt -s -w ./

gen-doc: ## generate markdown documentation
	rm -f ./docs/*
	go run cmd/gen-doc/gen-doc.go
	sed -i.bak 's/.*koyeb completion.*//' ./docs/*.md
	sed -i.bak 's/### SEE ALSO.*//' ./docs/*.md
	cat ./docs/koyeb.md >> ./docs/reference.md
	cat ./docs/koyeb_login.md >> ./docs/reference.md
	cat ./docs/koyeb_apps.md >> ./docs/reference.md
	cat ./docs/koyeb_apps_*.md >> ./docs/reference.md
	cat ./docs/koyeb_domains.md >> ./docs/reference.md
	cat ./docs/koyeb_domains_*.md >> ./docs/reference.md
	cat ./docs/koyeb_secrets.md >> ./docs/reference.md
	cat ./docs/koyeb_secrets_*.md >> ./docs/reference.md
	cat ./docs/koyeb_services.md >> ./docs/reference.md
	cat ./docs/koyeb_services_*.md >> ./docs/reference.md
	cat ./docs/koyeb_deployments.md >> ./docs/reference.md
	cat ./docs/koyeb_deployments_*.md >> ./docs/reference.md
	cat ./docs/koyeb_instances.md >> ./docs/reference.md
	cat ./docs/koyeb_instances_*.md >> ./docs/reference.md
	cat ./docs/koyeb_version.md >> ./docs/reference.md
	find ./docs -type f -not -name 'reference.md' -delete

test-linux:
	go test $(TEST_OPTS) ./...

test-windows:
	GOOS=windows go test $(TEST_OPTS) ./...

test-fmt:
	test -z "`gofmt -d . | tee /dev/stderr`"

test: tidy cmd pkg test-fmt test-linux ## launch tests

test-all: tidy cmd pkg test-fmt test-linux test-windows ## launch tests

lint:
	golangci-lint run -v ./...