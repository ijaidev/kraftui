set shell := ["bash", "-euo", "pipefail", "-c"]

root := justfile_directory()
bin  := root / "bin" / "kraftui"

# Default: list recipes
default:
    @just --list

# Install frontend dependencies
deps:
    cd frontend && pnpm install

# Build Next.js static export into internal/embed/dist
frontend-build: deps
    cd frontend && STATIC_EXPORT=1 pnpm build
    rm -rf internal/embed/dist
    mkdir -p internal/embed/dist
    cp -a frontend/out/. internal/embed/dist/

# Build Go binary only (expects embed dist to exist)
build-go:
    mkdir -p "{{ root }}/bin"
    go build -o "{{ bin }}" ./cmd/kraftui

# Full build: frontend embed + Go binary
build: frontend-build build-go

# Run the server (builds first)
run:
    "{{ bin }}"

# Run

# Next.js dev server (no embed; use while iterating on UI)
frontend-dev: deps
    cd frontend && pnpm dev

# Remove build artifacts; restore embed placeholder
clean:
    rm -rf bin frontend/.next frontend/out internal/embed/dist
    mkdir -p internal/embed/dist
