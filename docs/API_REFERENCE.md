# API Reference

Complete API documentation for all microservices endpoints, authentication, data models, and integration patterns.

## 🌐 Base Information

**Architecture**: Microservices with API Gateway  
**Load Balancer**: http://localhost (Nginx)  
**API Gateway**: http://localhost:8080  
**Auth Service**: http://localhost:8081  
**User Service**: http://localhost:8082  
**API Version**: v1

## 📋 Request Flow

All client requests follow this pattern:
```
Client → Load Balancer (Nginx:80) → API Gateway (8080) → Service (8081/8082)
```

For direct development access, you can also call services directly, but production should always go through the load balancer.

## 🔐 Authentication

The API uses **JWT (JSON Web Token)** authentication with role-based access control managed by the Auth Service.

### Authentication Header
```
Authorization: Bearer <jwt-token>
```

### User Roles
- **user**: Standard user access (profile management, personal operations)
- **admin**: Administrative access (user management, system operations)

### JWT Token Structure
```json
{
  "user_id": "uuid-v4",
  "email": "user@example.com",
  "role": "user",
  "iat": 1672531200,
  "exp": 1672617600,
  "iss": "api-server"
}
```

## 📨 Response Format

All API responses follow this consistent structure:

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response data here
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "uuid-v4"
}
```

### Error Response
```json
{
  "success": false,
  "message": "Operation failed",
  "error": {
    "code": "ERROR_CODE",
    "details": "Detailed error message",
    "field_errors": {
      "email": ["Email is required", "Invalid email format"]
    }
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "uuid-v4"
}
```

## 🏥 System Health Endpoints

### Load Balancer Health
Check overall system health via load balancer.

**Endpoint**: `GET /health`  
**URL**: `http://localhost/health`  
**Authentication**: None

**Response**: 200 OK
```json
{
  "status": "healthy",
  "services": {
    "api_gateway": "healthy",
    "auth_service": "healthy", 
    "user_service": "healthy"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Individual Service Health

#### API Gateway Health
**Endpoint**: `GET /health`  
**URL**: `http://localhost:8080/health`

#### Auth Service Health
**Endpoint**: `GET /health`  
**URL**: `http://localhost:8081/health`

#### User Service Health
**Endpoint**: `GET /health`  
**URL**: `http://localhost:8082/health`

**Response** (for all services): 200 OK
```json
{
  "status": "healthy",
  "service": "user-service",
  "version": "v1.0.0",
  "uptime": "2h30m15s",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Service Readiness
Check if services are ready to handle requests.

**Endpoint**: `GET /ready`  
**Authentication**: None

**Response**: 200 OK (ready) / 503 Service Unavailable (not ready)
```json
{
  "status": "ready",
  "service": "user-service",
  "checks": {
    "database": {
      "status": "healthy",
      "response_time": "5ms"
    },
    "redis": {
      "status": "healthy",
      "response_time": "2ms"
    }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 🔐 Authentication Endpoints (Auth Service)

All authentication endpoints are accessed through the API Gateway but handled by the Auth Service.

### User Login
Authenticate user and receive JWT tokens.

**Endpoint**: `POST /api/v1/auth/login`  
**URL**: `http://localhost/api/v1/auth/login` (via Load Balancer)  
**Authentication**: None  
**Rate Limiting**: 5 requests per minute per IP

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Validation Rules**:
- `email`: Required, valid email format
- `password`: Required, minimum 1 character

**Response**: 200 OK
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": "uuid-v4",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "user",
      "status": "active"
    }
  }
}
```

**Error Responses**:
- `400 Bad Request`: Missing email or password
- `401 Unauthorized`: Invalid credentials or inactive user
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### Token Refresh
Refresh an expired JWT token using refresh token.

**Endpoint**: `POST /api/v1/auth/refresh`  
**URL**: `http://localhost/api/v1/auth/refresh`  
**Authentication**: None (requires refresh token)  
**Rate Limiting**: 10 requests per minute per IP

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response**: 200 OK
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

### Token Validation
Validate a JWT token (primarily for service-to-service communication).

**Endpoint**: `POST /api/v1/auth/validate`  
**URL**: `http://localhost/api/v1/auth/validate`  
**Authentication**: None  
**Rate Limiting**: 100 requests per minute per IP

**Request Body**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response**: 200 OK
```json
{
  "success": true,
  "message": "Token is valid",
  "data": {
    "valid": true,
    "user": {
      "id": "uuid-v4",
      "email": "user@example.com",
      "role": "user"
    },
    "expires_at": "2024-01-02T12:00:00Z"
  }
}
```

### User Logout
Logout user and invalidate tokens.

**Endpoint**: `POST /api/v1/auth/logout`  
**URL**: `http://localhost/api/v1/auth/logout`  
**Authentication**: Required (Bearer token)  
**Rate Limiting**: Standard

**Request Body**: None

**Response**: 200 OK
```json
{
  "success": true,
  "message": "Logout successful",
  "data": null
}
```

### Get Current User
Get current authenticated user information.

**Endpoint**: `GET /api/v1/auth/me`  
**URL**: `http://localhost/api/v1/auth/me`  
**Authentication**: Required (Bearer token)  
**Rate Limiting**: Standard

**Response**: 200 OK
```json
{
  "success": true,
  "message": "User information retrieved",
  "data": {
    "id": "uuid-v4",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "user",
    "status": "active",
    "created_at": "2024-01-01T10:00:00Z",
    "last_login_at": "2024-01-01T12:00:00Z"
  }
}
```

## 👤 User Management Endpoints (User Service)

All user management endpoints are accessed through the API Gateway but handled by the User Service.

### User Registration
Register a new user account.

**Endpoint**: `POST /api/v1/users/register`  
**URL**: `http://localhost/api/v1/users/register`  
**Authentication**: None  
**Rate Limiting**: 10 requests per minute per IP

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Validation Rules**:
- `email`: Required, valid email format, unique, max 255 characters
- `password`: Required, minimum 8 characters, must contain letters and numbers
- `first_name`: Required, 1-50 characters, letters and spaces only
- `last_name`: Required, 1-50 characters, letters and spaces only

**Response**: 201 Created
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": "uuid-v4",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "user",
    "status": "active",
    "email_verified": false,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**Error Responses**:
- `400 Bad Request`: Validation errors
- `409 Conflict`: Email already exists
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### Get User Profile
Retrieve current user's profile information.

**Endpoint**: `GET /api/v1/users/profile`  
**URL**: `http://localhost/api/v1/users/profile`  
**Authentication**: Required (User/Admin)  
**Rate Limiting**: Standard

**Response**: 200 OK
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
    "status": "active",
    "email_verified": true,
    "email_verified_at": "2024-01-01T11:00:00Z",
    "last_login_at": "2024-01-01T12:00:00Z",
    "login_count": 15,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### Update User Profile
Update current user's profile information.

**Endpoint**: `PUT /api/v1/users/profile`  
**URL**: `http://localhost/api/v1/users/profile`  
**Authentication**: Required (User/Admin)  
**Rate Limiting**: Standard

**Request Body** (all fields optional):
```json
{
  "first_name": "Updated",
  "last_name": "Name"
}
```

**Validation Rules**:
- `first_name`: Optional, 1-50 characters, letters and spaces only
- `last_name`: Optional, 1-50 characters, letters and spaces only

**Response**: 200 OK
```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": {
    "id": "uuid-v4",
    "email": "user@example.com",
    "first_name": "Updated",
    "last_name": "Name",
    "role": "user",
    "status": "active",
    "updated_at": "2024-01-01T12:30:00Z"
  }
}
```

### Change Password
Change current user's password.

**Endpoint**: `POST /api/v1/users/change-password`  
**URL**: `http://localhost/api/v1/users/change-password`  
**Authentication**: Required (User/Admin)  
**Rate Limiting**: 5 requests per minute per user

**Request Body**:
```json
{
  "current_password": "old_password123",
  "new_password": "new_password456"
}
```

**Validation**:
- `current_password`: Required, must match existing password
- `new_password`: Required, minimum 8 characters, must contain letters and numbers, must be different from current password

**Response**: 200 OK
```json
{
  "success": true,
  "message": "Password changed successfully",
  "data": {
    "password_changed_at": "2024-01-01T12:30:00Z"
  }
}
```

**Error Responses**:
- `400 Bad Request`: Validation errors or new password same as current
- `401 Unauthorized`: Incorrect current password
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### Validate User Credentials
Internal endpoint for auth service to validate user credentials.

**Endpoint**: `POST /api/v1/users/validate-credentials`  
**URL**: `http://localhost/api/v1/users/validate-credentials`  
**Authentication**: Service-to-service (internal)  
**Rate Limiting**: High limit for internal use

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response**: 200 OK
```json
{
  "success": true,
  "message": "Credentials validated",
  "data": {
    "valid": true,
    "user": {
      "id": "uuid-v4",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "user",
      "status": "active"
    }
  }
}
```

## 🔧 Admin Endpoints (User Service)

All admin endpoints require **admin role** authentication.

### List Users
Retrieve paginated list of all users with filtering options.

**Endpoint**: `GET /api/v1/users/users`  
**URL**: `http://localhost/api/v1/users/users`  
**Authentication**: Required (Admin only)  
**Rate Limiting**: Standard

**Query Parameters**:
- `page` (optional): Page number, default 1, minimum 1
- `limit` (optional): Items per page, default 10, minimum 1, maximum 100
- `status` (optional): Filter by status (active, inactive, suspended)
- `role` (optional): Filter by role (user, admin)
- `search` (optional): Search by email, first name, or last name
- `sort` (optional): Sort field (created_at, updated_at, email), default created_at
- `order` (optional): Sort order (asc, desc), default desc

**Example**: `GET /api/v1/users/users?page=2&limit=20&status=active&search=john&sort=email&order=asc`

**Response**: 200 OK
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": {
    "users": [
      {
        "id": "uuid-v4",
        "email": "user1@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "role": "user",
        "status": "active",
        "email_verified": true,
        "last_login_at": "2024-01-01T11:30:00Z",
        "login_count": 25,
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T11:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 150,
      "total_pages": 15,
      "has_next": true,
      "has_prev": false
    },
    "filters": {
      "status": "active",
      "role": null,
      "search": "john"
    }
  }
}
```

### Get User by ID
Retrieve specific user information by ID.

**Endpoint**: `GET /api/v1/users/users/{id}`  
**URL**: `http://localhost/api/v1/users/users/{id}`  
**Authentication**: Required (Admin only)  
**Rate Limiting**: Standard

**Path Parameters**:
- `id`: User UUID (required)

**Response**: 200 OK
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
    "status": "active",
    "email_verified": true,
    "email_verified_at": "2024-01-01T10:30:00Z",
    "password_changed_at": "2024-01-01T10:00:00Z",
    "last_login_at": "2024-01-01T11:30:00Z",
    "login_count": 25,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T11:30:00Z"
  }
}
```

**Error Responses**:
- `404 Not Found`: User not found
- `403 Forbidden`: Insufficient permissions

### Update User (Admin)
Update any user's information as admin.

**Endpoint**: `PUT /api/v1/users/users/{id}`  
**URL**: `http://localhost/api/v1/users/users/{id}`  
**Authentication**: Required (Admin only)  
**Rate Limiting**: Standard

**Path Parameters**:
- `id`: User UUID (required)

**Request Body** (all fields optional):
```json
{
  "first_name": "Updated",
  "last_name": "Name",
  "role": "admin",
  "status": "inactive",
  "email_verified": true
}
```

**Validation Rules**:
- `first_name`: Optional, 1-50 characters, letters and spaces only
- `last_name`: Optional, 1-50 characters, letters and spaces only
- `role`: Optional, must be "user" or "admin"
- `status`: Optional, must be "active", "inactive", or "suspended"
- `email_verified`: Optional, boolean

**Response**: 200 OK
```json
{
  "success": true,
  "message": "User updated successfully",
  "data": {
    "id": "uuid-v4",
    "email": "user@example.com",
    "first_name": "Updated",
    "last_name": "Name",
    "role": "admin",
    "status": "inactive",
    "email_verified": true,
    "updated_at": "2024-01-01T12:30:00Z"
  }
}
```

### Delete User
Soft delete a user account (marks as inactive).

**Endpoint**: `DELETE /api/v1/users/users/{id}`  
**URL**: `http://localhost/api/v1/users/users/{id}`  
**Authentication**: Required (Admin only)  
**Rate Limiting**: 10 requests per minute per admin

**Path Parameters**:
- `id`: User UUID (required)

**Response**: 200 OK
```json
{
  "success": true,
  "message": "User deleted successfully",
  "data": {
    "id": "uuid-v4",
    "status": "inactive",
    "deleted_at": "2024-01-01T12:30:00Z"
  }
}
```

**Note**: This is a soft delete - the user is marked as inactive but data is retained for audit purposes.

### Get User Statistics
Get system-wide user statistics (admin only).

**Endpoint**: `GET /api/v1/users/stats`  
**URL**: `http://localhost/api/v1/users/stats`  
**Authentication**: Required (Admin only)  
**Rate Limiting**: Standard

**Query Parameters**:
- `period` (optional): Time period (day, week, month, year), default month

**Response**: 200 OK
```json
{
  "success": true,
  "message": "Statistics retrieved successfully",
  "data": {
    "total_users": 1500,
    "active_users": 1350,
    "inactive_users": 120,
    "suspended_users": 30,
    "admin_users": 15,
    "email_verified_users": 1200,
    "registrations_this_period": 45,
    "logins_this_period": 2850,
    "period": "month",
    "generated_at": "2024-01-01T12:00:00Z"
  }
}
```

## 📊 Data Models

### User Model
Complete user entity with all fields:

```json
{
  "id": "string (UUID)",
  "email": "string (unique, max 255)",
  "password": "string (hashed, not returned in responses)",
  "first_name": "string (max 50)",
  "last_name": "string (max 50)",
  "role": "string (user|admin)",
  "status": "string (active|inactive|suspended)",
  "email_verified": "boolean",
  "email_verified_at": "timestamp (nullable)",
  "password_changed_at": "timestamp (nullable)",
  "last_login_at": "timestamp (nullable)",
  "login_count": "integer (default 0)",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### JWT Token Structure
```json
{
  "user_id": "string (UUID)",
  "email": "string",
  "role": "string (user|admin)",
  "iat": "integer (issued at timestamp)",
  "exp": "integer (expiration timestamp)",
  "iss": "string (issuer: api-server)",
  "type": "string (access|refresh)"
}
```

### Session Model (Internal)
```json
{
  "id": "string (UUID)",
  "user_id": "string (UUID)",
  "token_hash": "string",
  "device_info": "object (optional)",
  "ip_address": "string (IP)",
  "user_agent": "string (optional)",
  "expires_at": "timestamp",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

## ⚠️ Error Codes

| Code | HTTP Status | Description | Common Causes |
|------|-------------|-------------|---------------|
| `VALIDATION_ERROR` | 400 | Request validation failed | Invalid input data, missing required fields |
| `UNAUTHORIZED` | 401 | Authentication required or failed | Missing/invalid JWT token, expired token |
| `INVALID_TOKEN` | 401 | JWT token is invalid or expired | Malformed token, expired token |
| `INVALID_CREDENTIALS` | 401 | Login credentials are invalid | Wrong email/password combination |
| `FORBIDDEN` | 403 | Insufficient permissions | User role doesn't have required permissions |
| `NOT_FOUND` | 404 | Resource not found | User ID doesn't exist, endpoint not found |
| `CONFLICT` | 409 | Resource conflict | Email already exists during registration |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests | Client exceeded rate limit |
| `INTERNAL_SERVER_ERROR` | 500 | Internal server error | Database error, service unavailable |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable | Database connection lost, service down |

## 🚦 Rate Limiting

Rate limiting is applied per client IP address with different limits for different endpoint types:

| Endpoint Type | Requests | Window | Scope |
|---------------|----------|--------|-------|
| Health checks | Unlimited | - | Per IP |
| User registration | 10 | 1 minute | Per IP |
| User login | 5 | 1 minute | Per IP |
| Token refresh | 10 | 1 minute | Per IP |
| Password change | 5 | 1 minute | Per user |
| Profile operations | 30 | 1 minute | Per user |
| Admin operations | 50 | 1 minute | Per admin |
| General API | 100 | 1 minute | Per IP |

### Rate Limit Headers

Rate limit information is included in response headers:

- `X-RateLimit-Limit`: Maximum requests allowed in the time window
- `X-RateLimit-Remaining`: Number of requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when the rate limit resets
- `X-RateLimit-RetryAfter`: Seconds until rate limit resets (only when limit exceeded)

### Rate Limit Error Response
```json
{
  "success": false,
  "message": "Rate limit exceeded",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "details": "You have exceeded the rate limit of 5 requests per minute for this endpoint",
    "retry_after": 45
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "uuid-v4"
}
```

## 🧪 Testing Examples

### Using cURL

#### Complete User Registration and Login Flow
```bash
# 1. Register a new user
curl -X POST http://localhost/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'

# 2. Login to get JWT token
RESPONSE=$(curl -s -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

# Extract token from response (requires jq)
TOKEN=$(echo $RESPONSE | jq -r '.data.access_token')

# 3. Get user profile with token
curl -X GET http://localhost/api/v1/users/profile \
  -H "Authorization: Bearer $TOKEN"

# 4. Update user profile
curl -X PUT http://localhost/api/v1/users/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Updated",
    "last_name": "Name"
  }'

# 5. Change password
curl -X POST http://localhost/api/v1/users/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "password123",
    "new_password": "newpassword456"
  }'
```

#### Admin Operations (requires admin token)
```bash
# Get admin token (assuming admin user exists)
ADMIN_RESPONSE=$(curl -s -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "adminpassword"
  }')

ADMIN_TOKEN=$(echo $ADMIN_RESPONSE | jq -r '.data.access_token')

# List users with pagination and filtering
curl -X GET "http://localhost/api/v1/users/users?page=1&limit=10&status=active" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Get specific user by ID
curl -X GET http://localhost/api/v1/users/users/USER_ID_HERE \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Update user as admin
curl -X PUT http://localhost/api/v1/users/users/USER_ID_HERE \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "admin",
    "status": "active",
    "email_verified": true
  }'

# Get user statistics
curl -X GET http://localhost/api/v1/users/stats \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Using HTTPie

#### Registration and Authentication
```bash
# Register user
http POST localhost/api/v1/users/register \
  email=test@example.com \
  password=password123 \
  first_name=Test \
  last_name=User

# Login
http POST localhost/api/v1/auth/login \
  email=test@example.com \
  password=password123

# Use token (replace YOUR_TOKEN with actual token)
http GET localhost/api/v1/users/profile \
  Authorization:"Bearer YOUR_TOKEN"
```

### JavaScript/Fetch Examples

```javascript
// API client class for easier management
class ApiClient {
  constructor(baseUrl = 'http://localhost') {
    this.baseUrl = baseUrl;
    this.token = localStorage.getItem('auth_token');
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    const config = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    if (this.token && !config.headers.Authorization) {
      config.headers.Authorization = `Bearer ${this.token}`;
    }

    const response = await fetch(url, config);
    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || 'Request failed');
    }

    return data;
  }

  // Authentication methods
  async register(userData) {
    return this.request('/api/v1/users/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
  }

  async login(credentials) {
    const response = await this.request('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });

    if (response.success) {
      this.token = response.data.access_token;
      localStorage.setItem('auth_token', this.token);
      localStorage.setItem('refresh_token', response.data.refresh_token);
    }

    return response;
  }

  async refreshToken() {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }

    const response = await this.request('/api/v1/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (response.success) {
      this.token = response.data.access_token;
      localStorage.setItem('auth_token', this.token);
      localStorage.setItem('refresh_token', response.data.refresh_token);
    }

    return response;
  }

  async logout() {
    await this.request('/api/v1/auth/logout', {
      method: 'POST',
    });

    this.token = null;
    localStorage.removeItem('auth_token');
    localStorage.removeItem('refresh_token');
  }

  // User methods
  async getProfile() {
    return this.request('/api/v1/users/profile');
  }

  async updateProfile(updates) {
    return this.request('/api/v1/users/profile', {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
  }

  async changePassword(passwords) {
    return this.request('/api/v1/users/change-password', {
      method: 'POST',
      body: JSON.stringify(passwords),
    });
  }

  // Admin methods
  async listUsers(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const endpoint = `/api/v1/users/users${queryString ? `?${queryString}` : ''}`;
    return this.request(endpoint);
  }

  async getUser(userId) {
    return this.request(`/api/v1/users/users/${userId}`);
  }

  async updateUser(userId, updates) {
    return this.request(`/api/v1/users/users/${userId}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
  }

  async deleteUser(userId) {
    return this.request(`/api/v1/users/users/${userId}`, {
      method: 'DELETE',
    });
  }

  async getUserStats(period = 'month') {
    return this.request(`/api/v1/users/stats?period=${period}`);
  }
}

// Usage examples
const api = new ApiClient();

// Registration flow
try {
  const registerResult = await api.register({
    email: 'user@example.com',
    password: 'password123',
    first_name: 'John',
    last_name: 'Doe'
  });
  console.log('User registered:', registerResult.data);

  const loginResult = await api.login({
    email: 'user@example.com',
    password: 'password123'
  });
  console.log('Login successful:', loginResult.data.user);

  const profile = await api.getProfile();
  console.log('User profile:', profile.data);

} catch (error) {
  console.error('API Error:', error.message);
}

// Admin operations
try {
  const users = await api.listUsers({
    page: 1,
    limit: 10,
    status: 'active'
  });
  console.log('Users:', users.data);

  const stats = await api.getUserStats('month');
  console.log('Statistics:', stats.data);

} catch (error) {
  console.error('Admin API Error:', error.message);
}
```

## 🔒 Security Considerations

### JWT Token Security
1. **Store Securely**: Use secure HTTP-only cookies or secure local storage
2. **Token Rotation**: Refresh tokens regularly before expiration
3. **Logout Handling**: Always call logout endpoint to invalidate tokens
4. **Token Validation**: Validate tokens on every protected request

### Password Security
1. **Strong Passwords**: Enforce minimum 8 characters with letters and numbers
2. **Password Hashing**: bcrypt with salt (handled server-side)
3. **Password Changes**: Require current password for changes
4. **Rate Limiting**: Strict limits on password-related operations

### API Security
1. **HTTPS Only**: Always use HTTPS in production
2. **Rate Limiting**: Respect rate limits to avoid blocking
3. **Input Validation**: Validate all input on client-side and server handles validation
4. **Error Handling**: Don't expose sensitive information in error messages

### Communication Security
1. **Service-to-Service**: Internal service communication is secured
2. **Request IDs**: Use correlation IDs for request tracking
3. **User Context**: User information is properly propagated between services
4. **Audit Logging**: All user actions are logged for security auditing

## 📝 Best Practices

### API Integration
1. **Error Handling**: Always check the `success` field in responses
2. **Token Management**: Implement automatic token refresh
3. **Retry Logic**: Implement exponential backoff for transient errors
4. **Pagination**: Use pagination parameters for list endpoints
5. **Caching**: Cache user profiles and other frequently accessed data

### Performance Optimization
1. **Request Batching**: Batch related API calls when possible
2. **Conditional Requests**: Use appropriate caching headers
3. **Connection Reuse**: Use HTTP connection pooling
4. **Compression**: Enable gzip compression for responses

### Monitoring and Debugging
1. **Request IDs**: Use correlation IDs for tracking requests across services
2. **Logging**: Log all API interactions for debugging
3. **Health Checks**: Regularly check service health endpoints
4. **Error Reporting**: Implement proper error reporting and alerting

This comprehensive API reference provides all the information needed to integrate with the microservices architecture. Each endpoint is thoroughly documented with request/response examples, error handling, and security considerations.