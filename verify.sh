#!/bin/bash

# Verification script for distributed-systems-go-demos
# This script tests all examples to ensure they work correctly

set -e  # Exit on error

echo "=================================================="
echo "Distributed Systems Go Demos - Verification Script"
echo "=================================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print success
success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Function to print error
error() {
    echo -e "${RED}✗${NC} $1"
}

# Function to print info
info() {
    echo -e "${YELLOW}→${NC} $1"
}

echo "Step 1: Verify Go version"
GO_VERSION=$(go version)
info "$GO_VERSION"
if [[ "$GO_VERSION" =~ go1\.(2[1-9]|[3-9][0-9]) ]]; then
    success "Go version 1.21+ detected"
else
    error "Go version must be 1.21 or higher"
    exit 1
fi
echo ""

echo "Step 2: Build all packages"
info "Running: go build ./..."
if go build ./...; then
    success "All packages build successfully"
else
    error "Build failed"
    exit 1
fi
echo ""

echo "Step 3: Format check"
info "Running: go fmt ./..."
FORMATTED=$(go fmt ./...)
if [ -z "$FORMATTED" ]; then
    success "All code is properly formatted"
else
    error "Some files need formatting:"
    echo "$FORMATTED"
fi
echo ""

echo "Step 4: Vet check"
info "Running: go vet ./..."
if go vet ./...; then
    success "No vet issues found"
else
    error "Vet found issues"
    exit 1
fi
echo ""

echo "Step 5: Check project structure"
REQUIRED_DIRS=(
    "cmd/01_raw_socket/server"
    "cmd/01_raw_socket/client"
    "cmd/02_simple_rpc/server"
    "cmd/02_simple_rpc/client"
    "cmd/03_message_broker/broker"
    "cmd/03_message_broker/producer"
    "cmd/03_message_broker/consumer"
    "cmd/04_real_world_examples/grpc_example/server"
    "cmd/04_real_world_examples/grpc_example/client"
    "cmd/04_real_world_examples/nats_example/publisher"
    "cmd/04_real_world_examples/nats_example/subscriber"
    "internal/rpc"
    "internal/broker"
    "pkg/socket"
    "api/proto"
    "docs"
)

for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        success "Directory exists: $dir"
    else
        error "Missing directory: $dir"
        exit 1
    fi
done
echo ""

echo "Step 6: Check documentation"
REQUIRED_DOCS=(
    "README.md"
    "docs/01_raw_socket.md"
    "docs/02_simple_rpc.md"
    "docs/03_message_broker.md"
    "docs/04_real_world_examples.md"
)

for doc in "${REQUIRED_DOCS[@]}"; do
    if [ -f "$doc" ]; then
        success "Documentation exists: $doc"
    else
        error "Missing documentation: $doc"
        exit 1
    fi
done
echo ""

echo "Step 7: Test examples (basic smoke test)"
info "Testing 01_raw_socket server can compile..."
if go build -o /tmp/raw_socket_server cmd/01_raw_socket/server/main.go; then
    success "Raw socket server compiles successfully"
    rm /tmp/raw_socket_server
else
    error "Raw socket server failed to compile"
fi

info "Testing 02_simple_rpc server can compile..."
if go build -o /tmp/simple_rpc_server cmd/02_simple_rpc/server/main.go; then
    success "Simple RPC server compiles successfully"
    rm /tmp/simple_rpc_server
else
    error "Simple RPC server failed to compile"
fi

info "Testing 03_message_broker can compile..."
if go build -o /tmp/message_broker cmd/03_message_broker/broker/main.go; then
    success "Message broker compiles successfully"
    rm /tmp/message_broker
else
    error "Message broker failed to compile"
fi

info "Testing 04_grpc_example server can compile..."
if go build -o /tmp/grpc_server cmd/04_real_world_examples/grpc_example/server/main.go; then
    success "gRPC server compiles successfully"
    rm /tmp/grpc_server
else
    error "gRPC server failed to compile"
fi

echo ""
echo "=================================================="
echo -e "${GREEN}All verification checks passed!${NC}"
echo "=================================================="
echo ""
echo "Project is ready for use. See README.md for usage instructions."
