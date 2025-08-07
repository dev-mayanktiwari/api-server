module github.com/dev-mayanktiwari/api-server/services/user-service

go 1.23.2

require (
	github.com/dev-mayanktiwari/api-server/shared v0.0.0
	github.com/gin-gonic/gin v1.10.1
	github.com/google/uuid v1.6.0
)

// Use local shared module for development
replace github.com/dev-mayanktiwari/api-server/shared => ../../shared