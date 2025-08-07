#!/bin/bash

echo "Building microservices..."

echo "Building Auth Service..."
cd services/auth-service
go mod tidy
docker build -t auth-service:latest .
cd ../..

echo "Building User Service..."
cd services/user-service
go mod tidy
docker build -t user-service:latest .
cd ../..

echo "Building API Gateway..."
cd services/api-gateway
go mod tidy
docker build -t api-gateway:latest .
cd ../..

echo "All services built successfully!"