#!/bin/bash

# Kubernetes Deployment Script for API Server
# This script deploys all services to Kubernetes

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="${NAMESPACE:-api-server}"
ENVIRONMENT="${ENVIRONMENT:-production}"
KUBECTL_CONTEXT="${KUBECTL_CONTEXT:-}"
DRY_RUN="${DRY_RUN:-false}"
WAIT_FOR_DEPLOYMENT="${WAIT_FOR_DEPLOYMENT:-true}"
TIMEOUT="${TIMEOUT:-300s}"

# Kubernetes manifests in deployment order
MANIFESTS=(
    "k8s/namespace.yaml"
    "k8s/configmap.yaml"
    "k8s/secrets.yaml"
    "k8s/postgres/pvc.yaml"
    "k8s/postgres/deployment.yaml"
    "k8s/postgres/service.yaml"
    "k8s/redis-deployment.yaml"
    "k8s/user-service/deployment.yaml"
    "k8s/user-service/service.yaml"
    "k8s/gateway-deployment.yaml"
    "k8s/monitoring.yaml"
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
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    # Set kubectl context if provided
    if [ -n "${KUBECTL_CONTEXT}" ]; then
        log_info "Setting kubectl context to ${KUBECTL_CONTEXT}"
        kubectl config use-context "${KUBECTL_CONTEXT}"
    fi
    
    # Test kubectl connectivity
    if ! kubectl cluster-info > /dev/null 2>&1; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Validate manifest file
validate_manifest() {
    local manifest=$1
    
    if [ ! -f "${manifest}" ]; then
        log_error "Manifest file not found: ${manifest}"
        return 1
    fi
    
    # Basic YAML validation using kubectl
    if ! kubectl apply --dry-run=client -f "${manifest}" > /dev/null 2>&1; then
        log_error "Invalid YAML syntax in: ${manifest}"
        return 1
    fi
    
    return 0
}

# Apply Kubernetes manifest
apply_manifest() {
    local manifest=$1
    
    log_info "Applying manifest: ${manifest}"
    
    if ! validate_manifest "${manifest}"; then
        return 1
    fi
    
    local kubectl_args=(apply -f "${manifest}")
    
    if [ "${DRY_RUN}" == "true" ]; then
        kubectl_args+=(--dry-run=client)
        log_info "DRY RUN: Would apply ${manifest}"
    fi
    
    if kubectl "${kubectl_args[@]}"; then
        log_success "Successfully applied: ${manifest}"
        return 0
    else
        log_error "Failed to apply: ${manifest}"
        return 1
    fi
}

# Wait for deployment to be ready
wait_for_deployment() {
    local deployment=$1
    
    if [ "${WAIT_FOR_DEPLOYMENT}" == "true" ] && [ "${DRY_RUN}" == "false" ]; then
        log_info "Waiting for deployment ${deployment} to be ready (timeout: ${TIMEOUT})..."
        
        if kubectl wait --for=condition=available deployment/"${deployment}" \
            --namespace="${NAMESPACE}" --timeout="${TIMEOUT}"; then
            log_success "Deployment ${deployment} is ready"
        else
            log_error "Deployment ${deployment} failed to become ready within ${TIMEOUT}"
            return 1
        fi
    fi
}

# Wait for statefulset to be ready
wait_for_statefulset() {
    local statefulset=$1
    
    if [ "${WAIT_FOR_DEPLOYMENT}" == "true" ] && [ "${DRY_RUN}" == "false" ]; then
        log_info "Waiting for statefulset ${statefulset} to be ready (timeout: ${TIMEOUT})..."
        
        if kubectl wait --for=condition=ready pod -l app="${statefulset}" \
            --namespace="${NAMESPACE}" --timeout="${TIMEOUT}"; then
            log_success "StatefulSet ${statefulset} is ready"
        else
            log_error "StatefulSet ${statefulset} failed to become ready within ${TIMEOUT}"
            return 1
        fi
    fi
}

# Deploy all manifests
deploy_all() {
    log_info "Starting deployment to namespace: ${NAMESPACE}"
    log_info "Environment: ${ENVIRONMENT}"
    log_info "Dry run: ${DRY_RUN}"
    echo
    
    local failed_deployments=()
    
    # Apply manifests in order
    for manifest in "${MANIFESTS[@]}"; do
        if apply_manifest "${manifest}"; then
            # Wait for specific deployments
            case "${manifest}" in
                *postgres/deployment.yaml)
                    wait_for_deployment "postgres" || failed_deployments+=("postgres")
                    ;;
                *redis-deployment.yaml)
                    wait_for_deployment "redis" || failed_deployments+=("redis")
                    ;;
                *user-service/deployment.yaml)
                    wait_for_deployment "user-service" || failed_deployments+=("user-service")
                    ;;
                *gateway-deployment.yaml)
                    wait_for_deployment "api-gateway" || failed_deployments+=("api-gateway")
                    ;;
                *monitoring.yaml)
                    wait_for_deployment "prometheus" || failed_deployments+=("prometheus")
                    wait_for_deployment "grafana" || failed_deployments+=("grafana")
                    ;;
            esac
        else
            failed_deployments+=("${manifest}")
        fi
        echo
    done
    
    # Report results
    if [ ${#failed_deployments[@]} -eq 0 ]; then
        log_success "All manifests deployed successfully!"
        show_deployment_status
    else
        log_error "Failed to deploy the following: ${failed_deployments[*]}"
        exit 1
    fi
}

# Show deployment status
show_deployment_status() {
    if [ "${DRY_RUN}" == "false" ]; then
        log_info "Deployment status:"
        echo
        
        echo "Pods:"
        kubectl get pods -n "${NAMESPACE}" -o wide
        echo
        
        echo "Services:"
        kubectl get services -n "${NAMESPACE}" -o wide
        echo
        
        echo "Ingress:"
        kubectl get ingress -n "${NAMESPACE}" -o wide || true
        echo
        
        echo "PersistentVolumeClaims:"
        kubectl get pvc -n "${NAMESPACE}" -o wide
        echo
    fi
}

# Rollback deployment
rollback_deployment() {
    local deployment=$1
    
    log_warning "Rolling back deployment: ${deployment}"
    
    if kubectl rollout undo deployment/"${deployment}" --namespace="${NAMESPACE}"; then
        log_success "Successfully rolled back: ${deployment}"
        wait_for_deployment "${deployment}"
    else
        log_error "Failed to rollback: ${deployment}"
        return 1
    fi
}

# Delete deployment
delete_deployment() {
    log_warning "Deleting all resources in namespace: ${NAMESPACE}"
    
    read -p "Are you sure you want to delete all resources? (y/N): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Delete in reverse order
        for ((i=${#MANIFESTS[@]}-1; i>=0; i--)); do
            manifest="${MANIFESTS[i]}"
            if [ -f "${manifest}" ]; then
                log_info "Deleting manifest: ${manifest}"
                kubectl delete -f "${manifest}" --ignore-not-found=true
            fi
        done
        log_success "Cleanup completed"
    else
        log_info "Deletion cancelled"
    fi
}

# Show logs
show_logs() {
    local service=$1
    local lines=${2:-100}
    
    log_info "Showing logs for ${service} (last ${lines} lines)..."
    kubectl logs -n "${NAMESPACE}" -l app="${service}" --tail="${lines}" -f
}

# Show usage
show_usage() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  deploy          Deploy all services (default)"
    echo "  status          Show deployment status"
    echo "  rollback SERVICE Rollback a specific deployment"
    echo "  delete          Delete all resources"
    echo "  logs SERVICE    Show logs for a service"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -n, --namespace NS      Kubernetes namespace (default: api-server)"
    echo "  -e, --environment ENV   Environment (default: production)"
    echo "  -c, --context CONTEXT   Kubectl context to use"
    echo "  -d, --dry-run           Perform a dry run"
    echo "  -w, --no-wait           Don't wait for deployments"
    echo "  -t, --timeout TIMEOUT   Timeout for waiting (default: 300s)"
    echo ""
    echo "Environment Variables:"
    echo "  NAMESPACE               Kubernetes namespace"
    echo "  ENVIRONMENT             Environment name"
    echo "  KUBECTL_CONTEXT         Kubectl context"
    echo "  DRY_RUN                 Set to 'true' for dry run"
    echo "  WAIT_FOR_DEPLOYMENT     Set to 'false' to skip waiting"
    echo "  TIMEOUT                 Timeout for deployments"
}

# Parse command line arguments
COMMAND="deploy"
if [ $# -gt 0 ] && [[ ! "$1" =~ ^- ]]; then
    COMMAND=$1
    shift
fi

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -c|--context)
            KUBECTL_CONTEXT="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -w|--no-wait)
            WAIT_FOR_DEPLOYMENT=false
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Execute command
case $COMMAND in
    deploy)
        check_prerequisites
        deploy_all
        ;;
    status)
        show_deployment_status
        ;;
    rollback)
        if [ $# -lt 1 ]; then
            log_error "Rollback requires a service name"
            exit 1
        fi
        check_prerequisites
        rollback_deployment "$1"
        ;;
    delete)
        check_prerequisites
        delete_deployment
        ;;
    logs)
        if [ $# -lt 1 ]; then
            log_error "Logs command requires a service name"
            exit 1
        fi
        show_logs "$1" "${2:-100}"
        ;;
    *)
        log_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac