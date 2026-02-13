TEST_OPTS=-v -test.timeout 300s

define gen-doc-in-dir
	rm -f ./$1/*
	go run cmd/gen-doc/gen-doc.go $1
	sed -i.bak 's/.*koyeb completion.*//' ./$1/*.md
	sed -i.bak 's/### SEE ALSO.*//' ./$1/*.md
	cat ./$1/koyeb.md >> ./$1/reference.md
	cat ./$1/koyeb_login.md >> ./$1/reference.md
	cat ./$1/koyeb_apps.md >> ./$1/reference.md
	cat ./$1/koyeb_apps_*.md >> ./$1/reference.md
	cat ./$1/koyeb_archives.md >> ./$1/reference.md
	cat ./$1/koyeb_archives_*.md >> ./$1/reference.md
	cat ./$1/koyeb_deploy.md >> ./$1/reference.md
	cat ./$1/koyeb_domains.md >> ./$1/reference.md
	cat ./$1/koyeb_domains_*.md >> ./$1/reference.md
	cat ./$1/koyeb_organizations.md >> ./$1/reference.md
	cat ./$1/koyeb_organizations_*.md >> ./$1/reference.md
	cat ./$1/koyeb_secrets.md >> ./$1/reference.md
	cat ./$1/koyeb_secrets_*.md >> ./$1/reference.md
	cat ./$1/koyeb_services.md >> ./$1/reference.md
	cat ./$1/koyeb_services_*.md >> ./$1/reference.md
	cat ./$1/koyeb_deployments.md >> ./$1/reference.md
	cat ./$1/koyeb_deployments_*.md >> ./$1/reference.md
	cat ./$1/koyeb_instances.md >> ./$1/reference.md
	cat ./$1/koyeb_instances_*.md >> ./$1/reference.md
	cat ./$1/koyeb_databases.md >> ./$1/reference.md
	cat ./$1/koyeb_databases_*.md >> ./$1/reference.md
	cat ./$1/koyeb_version.md >> ./$1/reference.md
	cat ./$1/koyeb_volumes.md >> ./$1/reference.md
	cat ./$1/koyeb_volumes_*.md >> ./$1/reference.md
	cat ./$1/koyeb_regions.md >> ./$1/reference.md
	cat ./$1/koyeb_regions_*.md >> ./$1/reference.md
	cat ./$1/koyeb_regional-deployments.md >> ./$1/reference.md
	cat ./$1/koyeb_regional-deployments_*.md >> ./$1/reference.md
	cat ./$1/koyeb_metrics.md >> ./$1/reference.md
	cat ./$1/koyeb_metrics_*.md >> ./$1/reference.md
	cat ./$1/koyeb_snapshots.md >> ./$1/reference.md
	cat ./$1/koyeb_snapshots_*.md >> ./$1/reference.md
	cat ./$1/koyeb_compose.md >> ./$1/reference.md
	cat ./$1/koyeb_compose_*.md >> ./$1/reference.md
	find ./$1 -type f -not -name 'reference.md' -delete
endef

help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_0-9-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

tidy: ## apply go mod tidy
	test ${CI} || go mod tidy

install: ## install
	go install cmd/koyeb/koyeb.go

fmt: ## apply go format
	gofmt -s -w ./

gen-doc: ## generate markdown documentation
	$(call gen-doc-in-dir,docs)

test: tidy lint
	@mkdir -p ./.temp
	@$(call gen-doc-in-dir,.temp)
	@diff -r -q ./docs ./.temp > /dev/null && { \
        test -z "`gofmt -d ./cmd ./pkg | tee /dev/stderr`"; \
        go test $(TEST_OPTS) ./...; \
    } || { \
        echo >&2 "make gen-doc has a diff"; \
	}
	@rm -rf ./.temp;

lint:
	golangci-lint run -v ./...
