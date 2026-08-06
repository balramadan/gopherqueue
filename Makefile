# ==============================================================================
# GopherQueue v0.1 - Automation Makefile
# ==============================================================================

# Variables
GO := go
BINARY_DIR := bin
CMD_SERVER := ./cmd/gopherq-server
CMD_DASHBOARD := ./cmd/gopherq-dashboard

help: ## Menampilkan daftar perintah yang tersedia
	@echo "Penggunaan: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Mengunduh dan membersihkan dependensi Go modules
	@echo "==> Downloading and tidying Go dependencies..."
	$(GO) mod download
	$(GO) mod tidy

lint: ## Menjalankan golangci-lint (atau go vet jika golangci-lint belum terinstall)
	@echo "==> Running static code analysis..."
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || ($(GO) vet ./... && echo "==> Static analysis passed via 'go vet ./...' (Install golangci-lint for advanced linting)")

test-unit: ## Menjalankan Unit Tests dengan Race Detector
	@echo "==> Running Unit Tests with Race Detector..."
	$(GO) test -v -race ./...

test-integration: ## Menjalankan Integration Tests
	@echo "==> Running Integration Tests..."
	$(GO) test -v -race -tags=integration ./...

test: test-unit ## Menjalankan seluruh pengujian unit test

test-coverage: ## Menghasilkan laporan test coverage dalam format HTML
	@echo "==> Generating Test Coverage Report..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "==> Laporan berhasil dibuat di coverage.html"

build: ## Kompilasi seluruh biner standalone (gopherq-server & gopherq-dashboard)
	@echo "==> Building binaries..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/gopherq-server $(CMD_SERVER)
	$(GO) build -o $(BINARY_DIR)/gopherq-dashboard $(CMD_DASHBOARD)
	@echo "==> Binaries saved in ./$(BINARY_DIR)"

run-redis: ## Menjalankan container Redis lokal untuk testing manual
	@echo "==> Starting local Redis container..."
	docker run -d --name gopherq-redis -p 6379:6379 redis:8.0.4-alpine || docker start gopherq-redis

stop-redis: ## Menghentikan dan menghapus container Redis lokal
	@echo "==> Stopping local Redis container..."
	docker stop gopherq-redis || true
	docker rm gopherq-redis || true

clean: ## Membersihkan file binary dan artefak coverage
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BINARY_DIR) coverage.out coverage.html