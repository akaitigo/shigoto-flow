.PHONY: build test lint format check

# Backend
build-backend:
	cd backend && go build -o bin/server ./cmd/server

test-backend:
	cd backend && go test -race -count=1 ./...

lint-backend:
	cd backend && golangci-lint run ./...

format-backend:
	cd backend && gofmt -w .
	cd backend && goimports -w .

# Frontend
install-frontend:
	cd frontend && pnpm install --frozen-lockfile

build-frontend:
	cd frontend && pnpm build

test-frontend:
	cd frontend && pnpm test

lint-frontend:
	cd frontend && pnpm lint

format-frontend:
	cd frontend && pnpm format

# Combined
build: build-backend build-frontend

test: test-backend test-frontend

lint: lint-backend lint-frontend

format: format-backend format-frontend

check: lint test build
	@echo "All checks passed"
