#!/bin/bash

# Docker Build Script for API Server
# This script builds Docker images for all services

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOCKER_REGISTRY="${DOCKER_REGISTRY:-localhost}"
VERSION="${VERSION:-latest}"
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Services to build
SERVICES=(
    "user-service"
)

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Build Docker image
build_image() {
    local service=$1
    local dockerfile=${2:-Dockerfile}
    
    log_info "Building ${service} image..."
    
    local image_name="${DOCKER_REGISTRY}/api-server/${service}:${VERSION}"
    local latest_tag="${DOCKER_REGISTRY}/api-server/${service}:latest"
    
    # Build arguments
    local build_args=(
        --build-arg "BUILD_DATE=${BUILD_DATE}"
        --build-arg "GIT_COMMIT=${GIT_COMMIT}"
        --build-arg "VERSION=${VERSION}"
        --build-arg "SERVICE=${service}"
    )
    
    # Build the image
    if docker build \
        "${build_args[@]}" \
        -t "${image_name}" \
        -t "${latest_tag}" \
        -f "${dockerfile}" \
        .; then
        log_success "Successfully built ${service} image"
        return 0
    else
        log_error "Failed to build ${service} image"
        return 1
    fi
}

# Push Docker image
push_image() {
    local service=$1
    
    if [ "${PUSH_IMAGES:-false}" == "true" ]; then
        log_info "Pushing ${service} image to registry..."
        
        local image_name="${DOCKER_REGISTRY}/api-server/${service}:${VERSION}"
        local latest_tag="${DOCKER_REGISTRY}/api-server/${service}:latest"
        
        if docker push "${image_name}" && docker push "${latest_tag}"; then
            log_success "Successfully pushed ${service} image"
        else
            log_error "Failed to push ${service} image"
            return 1
        fi
    else
        log_info "Skipping push for ${service} (PUSH_IMAGES not set to true)"
    fi
}

# Build all services
build_all_services() {
    log_info "Building all services..."
    
    local failed_builds=()
    
    for service in "${SERVICES[@]}"; do
        if build_image "${service}"; then
            push_image "${service}" || true
        else
            failed_builds+=("${service}")
        fi
        echo
    done
    
    if [ ${#failed_builds[@]} -eq 0 ]; then
        log_success "All services built successfully!"
    else
        log_error "Failed to build the following services: ${failed_builds[*]}"
        exit 1
    fi
}

# Clean up old images
cleanup_old_images() {
    if [ "${CLEANUP_OLD_IMAGES:-false}" == "true" ]; then
        log_info "Cleaning up old images..."
        
        # Remove dangling images
        docker image prune -f
        
        # Remove old versions (keep latest 3)
        for service in "${SERVICES[@]}"; do
            local old_images=$(docker images "${DOCKER_REGISTRY}/api-server/${service}" --format "table {{.Repository}}:{{.Tag}}\t{{.CreatedAt}}" | grep -v latest | sort -k2 -r | tail -n +4 | awk '{print $1}')
            if [ -n "${old_images}" ]; then
                log_info "Removing old images for ${service}..."
                echo "${old_images}" | xargs -r docker rmi
            fi
        done
        
        log_success "Cleanup completed"
    fi
}

# Main function
main() {
    log_info "Starting Docker build process..."
    echo "Registry: ${DOCKER_REGISTRY}"
    echo "Version: ${VERSION}"
    echo "Git Commit: ${GIT_COMMIT}"
    echo "Build Date: ${BUILD_DATE}"
    echo
    
    check_prerequisites
    build_all_services
    cleanup_old_images
    
    log_success "Docker build process completed successfully!"
}

# Show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -r, --registry REGISTRY Docker registry (default: localhost)"
    echo "  -v, --version VERSION   Image version tag (default: latest)"
    echo "  -p, --push              Push images to registry"
    echo "  -c, --cleanup           Clean up old images"
    echo ""
    echo "Environment Variables:"
    echo "  DOCKER_REGISTRY         Docker registry URL"
    echo "  VERSION                 Image version tag"
    echo "  PUSH_IMAGES            Set to 'true' to push images"
    echo "  CLEANUP_OLD_IMAGES     Set to 'true' to cleanup old images"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -r|--registry)
            DOCKER_REGISTRY="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -p|--push)
            export PUSH_IMAGES=true
            shift
            ;;
        -c|--cleanup)
            export CLEANUP_OLD_IMAGES=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Run main function
main