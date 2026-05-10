.PHONY: dev build run clean deps test install uninstall config install-all-fedora

# Run in development mode with hot reload
dev:
	wails dev

# Build production binary
build:
	export CGO_LDFLAGS="-L$(PWD)/libs -Wl,-rpath,$(PWD)/libs" && \
	wails build

GO_BIN := $(shell go env GOPATH)/bin

# Install built binary to GOPATH/bin (accessible as 'myrics' from anywhere)
install: build
	cp build/bin/myrics-overlay $(GO_BIN)/myrics
	@echo "Installed to $(GO_BIN)/myrics"

# Remove installed binary
uninstall:
	rm -f $(GO_BIN)/myrics
	@echo "Removed $(GO_BIN)/myrics"

# Copy example config if config.yaml doesn't exist yet
config:
	@test -f configs/config.yaml && echo "config.yaml already exists" || \
		(cp configs/config.yaml.example configs/config.yaml && echo "Created configs/config.yaml")

# Run the built application (after building)
run-built:
	./build/bin/myrics-overlay

# Clean build artifacts
clean:
	rm -rf build/
	rm -rf frontend/dist/

# Install Go dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test ./internal/... -v

# Install all Fedora dependencies (PortAudio + Wails)
install-all-fedora:
	@echo "Installing all dependencies for Fedora..."
	sudo dnf install -y \
		portaudio-devel \
		gtk3-devel \
		webkit2gtk4.1-devel \
		libX11-devel \
		libXcursor-devel \
		libXrandr-devel \
		libXinerama-devel \
		libXi-devel \
		mesa-libGL-devel \
		libXxf86vm-devel

# Install system dependencies (Fedora)
install-deps-fedora:
	@echo "Installing PortAudio..."
	sudo dnf install -y portaudio-devel

# Install system dependencies (macOS)
install-deps-mac:
	@echo "Installing PortAudio..."
	brew install portaudio

# Build Windows executable — run this from Windows (PowerShell/CMD) with wails installed.
# Requires: Go, wails CLI, and a C compiler (MinGW-w64 or MSVC via CGo).
# Note: ACRCloud and PortAudio are Linux-only; SMTC handles detection on Windows.
build-windows:
	wails build -platform windows/amd64 -o myrics-overlay.exe

# Run the built Windows exe (from Windows terminal)
run-windows:
	./build/bin/myrics-overlay.exe

# Install ACRCloud library to system path (requires sudo)
install-acrcloud-lib:
	@echo "Installing ACRCloud library to /usr/lib64..."
	sudo cp libs/libacrcloud_extr_tool.so /usr/lib64/
	sudo ldconfig
	@echo "Done! You can now run 'make dev'."