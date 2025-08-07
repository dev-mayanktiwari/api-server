#!/bin/bash

echo "Deploying microservices to Kubernetes..."

# Create namespace
echo "Creating namespace..."
kubectl apply -f k8s/namespace.yaml

# Deploy PostgreSQL
echo "Deploying PostgreSQL..."
kubectl apply -f k8s/postgres/

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/postgres -n microservices

# Deploy Auth Service
echo "Deploying Auth Service..."
kubectl apply -f k8s/auth-service/

# Wait for Auth Service to be ready
echo "Waiting for Auth Service to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/auth-service -n microservices

# Deploy User Service
echo "Deploying User Service..."
kubectl apply -f k8s/user-service/

# Wait for User Service to be ready
echo "Waiting for User Service to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/user-service -n microservices

# Deploy API Gateway
echo "Deploying API Gateway..."
kubectl apply -f k8s/api-gateway/

# Wait for API Gateway to be ready
echo "Waiting for API Gateway to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/api-gateway -n microservices

echo "All services deployed successfully!"
echo ""
echo "Check status with:"
echo "  kubectl get pods -n microservices"
echo ""
echo "Access the API at:"
echo "  kubectl get svc api-gateway -n microservices"