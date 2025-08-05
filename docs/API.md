# API Documentation

## Overview

The API Server provides a RESTful API for user management with JWT authentication, rate limiting, and comprehensive middleware support.

**Base URL:** `http://localhost:8080/api/v1`

## Authentication

Most endpoints require authentication using JWT tokens. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Rate Limiting

- **General API endpoints:** 100 requests per minute
- **Authentication endpoints:** 5 requests per minute

## Response Format

All API responses follow this structure:

```json
{
  "success": true|false,
  "message": "Human readable message",
  "data": {}, // Response data (success only)
  "error": {  // Error details (error only)
    "code": "ERROR_CODE",
    "message": "Error description"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "uuid-v4"
}
```

## Endpoints

### Health Check

#### GET /health
Check if the API is running.

**Authentication:** Not required

**Response:**
```json
{
  "success": true,
  "message": "API Server is healthy",
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

#### GET /ready
Check if the API is ready to serve requests.

**Authentication:** Not required

#### GET /live
Liveness probe for container orchestration.

**Authentication:** Not required

#### GET /version
Get API version information.

**Authentication:** Not required

### Public Endpoints

#### GET /api/v1/ping
Simple ping endpoint for testing connectivity.

**Authentication:** Not required

**Response:**
```json
{
  "success": true,
  "message": "pong",
  "data": {
    "timestamp": "2024-01-01T12:00:00Z",
    "server": "api-server"
  }
}
```

### Authentication

#### POST /api/v1/auth/register
Register a new user account.

**Authentication:** Not required  
**Rate Limited:** 5 requests per minute

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe",
  "role": "user" // Optional, defaults to "user"
}
```

**Validation:**
- `email`: Required, valid email format
- `password`: Required, minimum 8 characters
- `first_name`: Required, minimum 1 character
- `last_name`: Required, minimum 1 character
- `role`: Optional, valid values: "user", "admin"

**Response (201 Created):**
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "id": "uuid-v4",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "user",
    "is_active": true,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Validation errors
- `409 Conflict`: User already exists
- `429 Too Many Requests`: Rate limit exceeded

#### POST /api/v1/auth/login
Authenticate user and receive JWT token.

**Authentication:** Not required  
**Rate Limited:** 5 requests per minute

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "uuid-v4",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "user",
      "is_active": true,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    },
    "token": "jwt-token-string"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Validation errors
- `401 Unauthorized`: Invalid credentials
- `429 Too Many Requests`: Rate limit exceeded

### User Profile

#### GET /api/v1/profile
Get current user's profile.

**Authentication:** Required

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "id": "uuid-v4",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "user",
    "is_active": true,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

#### PUT /api/v1/profile
Update current user's profile.

**Authentication:** Required

**Request Body:**
```json
{
  "email": "newemail@example.com", // Optional
  "first_name": "NewFirstName",    // Optional
  "last_name": "NewLastName"       // Optional
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": {
    "id": "uuid-v4",
    "email": "newemail@example.com",
    "first_name": "NewFirstName",
    "last_name": "NewLastName",
    "role": "user",
    "is_active": true,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

#### POST /api/v1/profile/change-password
Change current user's password.

**Authentication:** Required

**Request Body:**
```json
{
  "current_password": "oldpassword123",
  "new_password": "newpassword123"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Password changed successfully",
  "data": null
}
```

### Admin Endpoints

All admin endpoints require authentication and admin role.

#### GET /api/v1/admin/users
List all users with pagination.

**Authentication:** Required (Admin only)

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 100)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": {
    "users": [
      {
        "id": "uuid-v4",
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "role": "user",
        "is_active": true,
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 5,
      "per_page": 10,
      "total_items": 50
    }
  }
}
```

#### GET /api/v1/admin/users/:id
Get a specific user by ID.

**Authentication:** Required (Admin only)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "id": "uuid-v4",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "user",
    "is_active": true,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

#### PUT /api/v1/admin/users/:id
Update a specific user.

**Authentication:** Required (Admin only)

**Request Body:**
```json
{
  "email": "newemail@example.com", // Optional
  "first_name": "NewFirstName",    // Optional
  "last_name": "NewLastName",      // Optional
  "role": "admin",                 // Optional
  "is_active": false               // Optional
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "User updated successfully",
  "data": {
    "id": "uuid-v4",
    "email": "newemail@example.com",
    "first_name": "NewFirstName",
    "last_name": "NewLastName",
    "role": "admin",
    "is_active": false,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

#### DELETE /api/v1/admin/users/:id
Delete a specific user (soft delete).

**Authentication:** Required (Admin only)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "User deleted successfully",
  "data": null
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `USER_ALREADY_EXISTS` | 409 | User with email already exists |
| `LOGIN_FAILED` | 401 | Invalid email or password |
| `MISSING_AUTH_HEADER` | 401 | Authorization header missing |
| `INVALID_AUTH_HEADER` | 401 | Invalid authorization header format |
| `INVALID_TOKEN` | 401 | Invalid or expired JWT token |
| `INSUFFICIENT_PERMISSIONS` | 403 | User lacks required permissions |
| `USER_NOT_FOUND` | 404 | User not found |
| `EMAIL_ALREADY_TAKEN` | 409 | Email is already in use |
| `INCORRECT_PASSWORD` | 400 | Current password is incorrect |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `METHOD_NOT_ALLOWED` | 405 | HTTP method not allowed |
| `INTERNAL_SERVER_ERROR` | 500 | Internal server error |

## Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required or failed
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `405 Method Not Allowed`: HTTP method not supported
- `409 Conflict`: Resource conflict (e.g., duplicate email)
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

## Testing

### Using curl

**Register a new user:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Get profile (requires token from login):**
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Using HTTPie

**Register:**
```bash
http POST localhost:8080/api/v1/auth/register \
  email=test@example.com \
  password=password123 \
  first_name=Test \
  last_name=User
```

**Login:**
```bash
http POST localhost:8080/api/v1/auth/login \
  email=test@example.com \
  password=password123
```

**Get profile:**
```bash
http GET localhost:8080/api/v1/profile \
  Authorization:"Bearer YOUR_JWT_TOKEN"
```