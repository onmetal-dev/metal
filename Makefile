SHELL := /bin/bash
# override this in CI to set to prod
LDFLAGS ?= -ldflags "-X main.Environment=dev"

.PHONY: templ-generate templ-watch
templ-generate: templ
	templ generate
templ-watch: templ
	templ generate --watch


.PHONY: tailwind-build tailwind-watch
TAILWINDCLI := bunx tailwindcss
tailwind-watch: bun
	$(TAILWINDCLI) -i ./cmd/app/static/css/input.css -o ./cmd/app/static/css/style.css --watch
tailwind-build: bun
	bun install
	$(TAILWINDCLI) -c ./cmd/app/tailwind.config.js -i ./cmd/app/static/css/input.css -o ./cmd/app/static/css/style.min.css --minify
	$(TAILWINDCLI) -c ./cmd/app/tailwind.config.js -i ./cmd/app/static/css/input.css -o ./cmd/app/static/css/style.css


.PHONY: install_deps templ bun docker
install_deps: templ bun staticcheck
	bun install
	go mod download
templ:
	@if ! command -v templ &> /dev/null; then \
		go install github.com/a-h/templ/cmd/templ@latest; \
	fi
bun:
	@if ! command -v bun &> /dev/null; then \
		curl -fsSL https://bun.sh/install | bash; \
	fi
staticcheck:
	@if ! command -v staticcheck &> /dev/null; then \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
docker:
	@if ! command -v docker &> /dev/null; then \
		echo "docker not found, install it from https://docs.docker.com/get-docker/"; \
		exit 1; \
	fi
air:
	@if ! command -v air &> /dev/null; then \
		go install github.com/cosmtrek/air@latest; \
	fi

.PHONY: build
build: templ-generate tailwind-build
	mkdir -p bin/
	go build $(LDFLAGS) -o ./bin/app ./cmd/app

.PHONY: dev-app
dev: bun
	bunx concurrently --kill-others-on-fail "make dev-tracing" "make dev-app" "make dev-db"
dev-tracing: docker
	cd env/local/tracing && docker compose up
dev-app: air build
	mkdir -p bin/
	source env/local/app/.env && air -c env/local/app/air.toml
dev-db: docker
	cd env/local/db && docker compose up
dev-hetzner-sandbox:
	source env/local/sandbox-hetzner/.env && go run ./cmd/sandbox-hetzner/main.go $(ARGS)
dev-talhelper-sandbox:
	source env/local/sandbox-talhelper/.env && go run ./cmd/sandbox-talhelper/main.go $(ARGS)


.PHONY: vet staticheck test
test-vet:
	go vet ./...
test-staticcheck: staticcheck
	staticcheck ./...
test-db:
	cd env/local/db-for-tests && docker compose up
test-go:
	go test -race -v -timeout 30s ./...
# this test makes actual requests to the hetzner api if given the correct env vars. so separate them out and don't run them in CI
test-hetzner-provider:
	source env/local/app/.env.hetzner.test && \
	  go test -race -v -timeout 30s -run ^TestHetznerProvider$$ github.com/onmetal-dev/metal/lib/serverprovider
test:
	bunx concurrently -s first --kill-others "make test-db" "make test-go"
