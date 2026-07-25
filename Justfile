set shell := ["bash", "-euo", "pipefail", "-c"]

root          := justfile_directory()
frontend_dir  := root / "frontend"
frontend_out  := frontend_dir / "out"
frontend_next := frontend_dir / ".next"
bin_dir       := root / "bin"
bin           := bin_dir / "kraftui"
dist          := root / "internal" / "ui" / "dist"
cmd_pkg       := "./cmd/kraftui"

# list (Default): list recipes
list:
    @just --list

# Install frontend dependencies
deps:
    cd "{{ frontend_dir }}" && pnpm install

# Build Next.js static export into internal/ui/dist
frontend-build: deps
    cd "{{ frontend_dir }}" && STATIC_EXPORT=1 pnpm build
    rm -rf "{{ dist }}"
    mkdir -p "{{ dist }}"
    cp -a "{{ frontend_out }}/." "{{ dist }}/"

# Build Go binary only (expects embed dist to exist)
build-go:
    mkdir -p "{{ bin_dir }}"
    go build -o "{{ bin }}" "{{ cmd_pkg }}"

# Full build: frontend embed + Go binary
build: frontend-build build-go

# Builds and Runs the Go server
dev: build-go
    {{ bin }}

# Run the server (Builds first)
run: build
    "{{ bin }}"

# Next.js dev server (no embed; use while iterating on UI)
frontend-dev: deps
    cd "{{ frontend_dir }}" && pnpm dev

# Remove build artifacts; restore embed placeholder
clean:
    rm -rf "{{ bin_dir }}" "{{ frontend_next }}" "{{ frontend_out }}" "{{ dist }}"
    mkdir -p "{{ dist }}"
