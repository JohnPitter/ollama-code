.PHONY: build run install clean test help deps

BINARY_NAME=ollama-code
BUILD_DIR=build
INSTALL_PATH_UNIX=/usr/local/bin
INSTALL_PATH_WIN=$(USERPROFILE)/bin

# Detectar SO
ifeq ($(OS),Windows_NT)
	BINARY=$(BINARY_NAME).exe
	RM=del /Q
	RMDIR=rmdir /S /Q
else
	BINARY=$(BINARY_NAME)
	RM=rm -f
	RMDIR=rm -rf
endif

# Build otimizado
build:
	@echo "🔨 Building optimized binary..."
	@go build -ldflags="-s -w" -trimpath -o $(BUILD_DIR)/$(BINARY) ./cmd/ollama-code
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY)"

# Build rápido para desenvolvimento
dev:
	@echo "🔧 Building dev version..."
	@go build -o $(BUILD_DIR)/$(BINARY) ./cmd/ollama-code
	@echo "✅ Dev build complete"

# Build para todos os sistemas
build-all:
	@echo "🚀 Building for all platforms..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/ollama-code
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/ollama-code
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/ollama-code
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/ollama-code
	@echo "✅ All builds complete"

# Rodar diretamente
run:
	@go run ./cmd/ollama-code chat

# Instalar globalmente (Linux/macOS)
install: build
ifeq ($(OS),Windows_NT)
	@echo "📦 Installing to $(INSTALL_PATH_WIN)..."
	@if not exist "$(INSTALL_PATH_WIN)" mkdir "$(INSTALL_PATH_WIN)"
	@copy "$(BUILD_DIR)\$(BINARY)" "$(INSTALL_PATH_WIN)\"
	@echo "✅ Installed! Add $(INSTALL_PATH_WIN) to PATH if needed"
else
	@echo "📦 Installing to $(INSTALL_PATH_UNIX)..."
	@sudo cp $(BUILD_DIR)/$(BINARY) $(INSTALL_PATH_UNIX)/
	@echo "✅ Installed! Use: ollama-code"
endif

# Instalar local (sem sudo)
install-local: build
ifeq ($(OS),Windows_NT)
	@echo "📦 Installing to $(INSTALL_PATH_WIN)..."
	@if not exist "$(INSTALL_PATH_WIN)" mkdir "$(INSTALL_PATH_WIN)"
	@copy "$(BUILD_DIR)\$(BINARY)" "$(INSTALL_PATH_WIN)\"
	@echo "✅ Installed to $(INSTALL_PATH_WIN)"
	@echo "💡 Add to PATH: setx PATH \"%PATH%;$(INSTALL_PATH_WIN)\""
else
	@echo "📦 Installing to ~/bin..."
	@mkdir -p ~/bin
	@cp $(BUILD_DIR)/$(BINARY) ~/bin/
	@echo "✅ Installed to ~/bin/ollama-code"
	@echo "💡 Add to PATH: export PATH=\$$PATH:~/bin"
endif

# Limpar
clean:
	@echo "🧹 Cleaning..."
ifeq ($(OS),Windows_NT)
	@if exist "$(BUILD_DIR)" $(RMDIR) "$(BUILD_DIR)"
else
	@$(RMDIR) $(BUILD_DIR)
endif
	@go clean
	@echo "✅ Cleaned"

# Testes
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Testes com coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# Download dependencies
deps:
	@echo "📥 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies ready"

# Verificar código
lint:
	@echo "🔍 Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run
	@echo "✅ Lint complete"

# Formatar código
fmt:
	@echo "✨ Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Benchmark
benchmark:
	@echo "📊 Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Versão release (máxima otimização)
release: deps
	@echo "🎯 Building release version..."
	@go build -ldflags="-s -w" -trimpath -tags netgo -installsuffix netgo -o $(BUILD_DIR)/$(BINARY) ./cmd/ollama-code
	@echo "✅ Release build complete"
	@echo "📊 Binary stats:"
ifeq ($(OS),Windows_NT)
	@dir "$(BUILD_DIR)\$(BINARY)"
else
	@ls -lh $(BUILD_DIR)/$(BINARY)
	@file $(BUILD_DIR)/$(BINARY)
endif

# CI/CD targets
ci: deps lint test build
	@echo "✅ CI pipeline completed successfully"

ci-full: deps lint test-coverage build-all
	@echo "✅ Full CI pipeline completed successfully"

# Test specific tools
test-tools:
	@echo "🧪 Running tools tests..."
	@go test -v ./internal/tools/...

# GoReleaser dry-run
release-dry-run:
	@echo "🎯 Running GoReleaser dry-run..."
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)
	@goreleaser release --snapshot --skip-publish --clean
	@echo "✅ Dry-run complete"

# Install CI tools
ci-tools:
	@echo "🔧 Installing CI tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/goreleaser/goreleaser@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ CI tools installed"

# Go vet
vet:
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Vet complete"

# Check all (lint + vet + test)
check: lint vet test
	@echo "✅ All checks passed"

# Ajuda
help:
	@echo "Available targets:"
	@echo "  make build         - Build optimized binary"
	@echo "  make dev           - Quick dev build"
	@echo "  make build-all     - Build for all platforms"
	@echo "  make run           - Run directly"
	@echo "  make install       - Install globally (requires sudo on Unix)"
	@echo "  make install-local - Install to ~/bin (no sudo)"
	@echo "  make clean         - Remove binary"
	@echo "  make test          - Run tests"
	@echo "  make test-tools    - Run tools tests only"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make deps          - Download dependencies"
	@echo "  make lint          - Run linter"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Run go vet"
	@echo "  make check         - Run all checks (lint + vet + test)"
	@echo "  make benchmark     - Run benchmarks"
	@echo "  make release       - Build release version"
	@echo "  make ci            - Run CI pipeline (lint + test + build)"
	@echo "  make ci-full       - Run full CI pipeline (lint + coverage + build-all)"
	@echo "  make ci-tools      - Install CI tools"
	@echo "  make release-dry-run - Test GoReleaser without publishing"
	@echo ""
	@echo "Examples:"
	@echo "  make build && ./build/ollama-code chat"
	@echo "  make install-local"
	@echo "  ollama-code chat --mode autonomous"
