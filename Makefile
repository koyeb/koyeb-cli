CARGO ?= cargo

help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_0-9-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

fmt: ## apply rustfmt
	$(CARGO) fmt --all

lint: ## run clippy
	$(CARGO) clippy --all-targets --all-features -- -D warnings

test: fmt lint ## run tests
	$(CARGO) test --all-targets --all-features

build: ## build release binary
	$(CARGO) build --release
