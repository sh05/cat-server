#!/bin/bash

# Docker Build Script for cat-server
# This script builds the cat-server Docker image and provides useful information

set -e  # Exit on any error

# Configuration
IMAGE_NAME="cat-server"
IMAGE_TAG="${1:-latest}"  # Default to latest if no tag provided
DOCKERFILE_PATH="./Dockerfile"
BUILD_CONTEXT="."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."

    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi

    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        exit 1
    fi

    # Check if Dockerfile exists
    if [ ! -f "$DOCKERFILE_PATH" ]; then
        print_error "Dockerfile not found at $DOCKERFILE_PATH"
        exit 1
    fi

    print_success "Prerequisites check passed"
}

# Function to build Docker image
build_image() {
    local full_image_name="${IMAGE_NAME}:${IMAGE_TAG}"

    print_status "Building Docker image: $full_image_name"
    print_status "Dockerfile: $DOCKERFILE_PATH"
    print_status "Build context: $BUILD_CONTEXT"

    local start_time=$(date +%s)

    # Build the image
    if docker build -f "$DOCKERFILE_PATH" -t "$full_image_name" "$BUILD_CONTEXT"; then
        local end_time=$(date +%s)
        local build_duration=$((end_time - start_time))

        print_success "Image built successfully in ${build_duration} seconds"

        # Show image information
        show_image_info "$full_image_name"

        return 0
    else
        print_error "Failed to build Docker image"
        return 1
    fi
}

# Function to show image information
show_image_info() {
    local image_name="$1"

    print_status "Image Information:"

    # Get image size
    local image_size=$(docker images "$image_name" --format "{{.Size}}")
    echo "  Size: $image_size"

    # Get image ID
    local image_id=$(docker images "$image_name" --format "{{.ID}}")
    echo "  Image ID: $image_id"

    # Get creation time
    local created=$(docker images "$image_name" --format "{{.CreatedSince}}")
    echo "  Created: $created"

    # Check image size in bytes and warn if too large
    local size_bytes=$(docker inspect "$image_name" --format='{{.Size}}' 2>/dev/null || echo "0")
    local max_size_mb=50
    local max_size_bytes=$((max_size_mb * 1024 * 1024))

    if [ "$size_bytes" -gt "$max_size_bytes" ]; then
        print_warning "Image size ($image_size) exceeds ${max_size_mb}MB target"
    else
        print_success "Image size is within ${max_size_mb}MB target"
    fi

    # Show layers count
    local layers=$(docker history --no-trunc "$image_name" | wc -l)
    echo "  Layers: $((layers - 1))"  # Subtract header line
}

# Function to run basic tests
run_basic_tests() {
    local full_image_name="${IMAGE_NAME}:${IMAGE_TAG}"

    print_status "Running basic image tests..."

    # Test 1: Check if image can start
    print_status "Test 1: Checking if container can start..."
    local container_id=$(docker run -d --rm "$full_image_name" || echo "failed")

    if [ "$container_id" = "failed" ]; then
        print_error "Container failed to start"
        return 1
    fi

    # Wait a moment for startup
    sleep 2

    # Check if container is still running
    if docker ps -q --filter "id=$container_id" | grep -q "$container_id"; then
        print_success "Container started successfully"

        # Test 2: Check if running as non-root user
        print_status "Test 2: Checking user..."
        local user=$(docker exec "$container_id" whoami 2>/dev/null || echo "unknown")
        if [ "$user" = "app" ]; then
            print_success "Container is running as non-root user: $user"
        else
            print_warning "Container user: $user (expected: app)"
        fi

        # Clean up
        docker stop "$container_id" > /dev/null 2>&1
    else
        print_error "Container stopped unexpectedly"
        return 1
    fi
}

# Function to show usage information
show_usage() {
    echo "Usage: $0 [IMAGE_TAG]"
    echo ""
    echo "Build cat-server Docker image with optional tag"
    echo ""
    echo "Arguments:"
    echo "  IMAGE_TAG    Optional tag for the image (default: latest)"
    echo ""
    echo "Examples:"
    echo "  $0           # Build with tag 'latest'"
    echo "  $0 v1.0      # Build with tag 'v1.0'"
    echo "  $0 dev       # Build with tag 'dev'"
    echo ""
    echo "Environment variables:"
    echo "  SKIP_TESTS   Set to 'true' to skip basic tests"
}

# Function to clean up old images
cleanup_old_images() {
    print_status "Cleaning up old images..."

    # Remove dangling images
    local dangling=$(docker images -f "dangling=true" -q)
    if [ -n "$dangling" ]; then
        docker rmi $dangling > /dev/null 2>&1 || true
        print_success "Removed dangling images"
    fi
}

# Main execution
main() {
    # Show help if requested
    if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        show_usage
        exit 0
    fi

    echo "=========================================="
    echo "    cat-server Docker Build Script"
    echo "=========================================="
    echo ""

    # Run all steps
    check_prerequisites
    echo ""

    cleanup_old_images
    echo ""

    if build_image; then
        echo ""

        # Run tests unless skipped
        if [ "$SKIP_TESTS" != "true" ]; then
            run_basic_tests
            echo ""
        fi

        print_success "Docker build completed successfully!"
        echo ""
        print_status "To run the container:"
        echo "  docker run -d --name cat-server -p 8080:8080 ${IMAGE_NAME}:${IMAGE_TAG}"
        echo ""
        print_status "To test the application:"
        echo "  curl http://localhost:8080/health"

    else
        print_error "Docker build failed!"
        exit 1
    fi
}

# Run main function with all arguments
main "$@"