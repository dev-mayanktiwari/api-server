# API Reference

Complete API documentation for the User Service endpoints, authentication, and data models.

## 🌐 Base Information

**Base URL**: `http://localhost:8082`  
**API Version**: v1  
**Full API Base**: `http://localhost:8082/api/v1/users`

## 🔐 Authentication

The API uses **JWT (JSON Web Token)** authentication with role-based access control.

### Authentication Header
```
Authorization: Bearer <jwt-token>
```

### Roles
- **user**: Standard user access (profile management)
- **admin**: Administrative access (user management)

## 📋 Response Format

All API responses follow this consistent structure:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {},
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "uuid-v4"
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error occurred",
  "error": {
    "code": "ERROR_CODE",
    "details": "Detailed error message"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "uuid-v4"
}
```

## 🏥 Health Endpoints

### Health Check
Check service health status.

**Endpoint**: `GET /health`  
**Authentication**: None  
**Rate Limiting**: None

**Response**: 200 OK
```json
{
  "status": "healthy",
  "service": "user-service",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Readiness Check
Check if service is ready to handle requests.

**Endpoint**: `GET /ready`  
**Authentication**: None  
**Rate Limiting**: None

**Response**: 200 OK (if ready) / 503 Service Unavailable (if not ready)
```json
{
  "status": "ready",
  "service": "user-service",
  "checks": {
    "database": {
      "status": "healthy"
    }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 👤 User Management Endpoints

### User Registration
Register a new user account.

**Endpoint**: `POST /api/v1/users/register`  
**Authentication**: None  
**Rate Limiting**: Standard

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
- `email`: Required, valid email format, unique
- `password`: Required, minimum 8 characters, must contain letters and numbers
- `first_name`: Required, 1-50 characters, letters only
- `last_name`: Required, 1-50 characters, letters only

**Response**: 201 Created
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
    "status": "active",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**Error Responses**:
- `400 Bad Request`: Validation errors
- `409 Conflict`: Email already exists
- `500 Internal Server Error`: Server error

### User Login
Authenticate user and receive JWT token.

**Endpoint**: `POST /api/v1/users/login`  
**Authentication**: None  
**Rate Limiting**: Standard

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
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
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
- `500 Internal Server Error`: Server error

## 👥 User Profile Endpoints

### Get User Profile
Retrieve current user's profile information.

**Endpoint**: `GET /api/v1/users/profile`  
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
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### Update User Profile
Update current user's profile information.

**Endpoint**: `PUT /api/v1/users/profile`  
**Authentication**: Required (User/Admin)  
**Rate Limiting**: Standard

**Request Body** (all fields optional):
```json
{
  "first_name": "Updated",
  "last_name": "Name"
}
```

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
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### Change Password
Change current user's password.

**Endpoint**: `POST /api/v1/users/change-password`  
**Authentication**: Required (User/Admin)  
**Rate Limiting**: Strict

**Request Body**:
```json
{
  "current_password": "old_password123",
  "new_password": "new_password123"
}
```

**Validation**:
- `current_password`: Must match existing password
- `new_password`: Same rules as registration password

**Response**: 200 OK
```json
{
  "success": true,
  "message": "Password changed successfully",
  "data": null
}
```

**Error Responses**:
- `400 Bad Request`: Validation errors
- `401 Unauthorized`: Incorrect current password
- `500 Internal Server Error`: Server error

## 🔧 Admin Endpoints

All admin endpoints require **admin role** authentication.

### List Users
Retrieve paginated list of all users.

**Endpoint**: `GET /api/v1/users/users`  
**Authentication**: Required (Admin only)  
**Rate Limiting**: Standard

**Query Parameters**:
- `page` (optional): Page number, default 1
- `limit` (optional): Items per page, default 10, max 100
- `status` (optional): Filter by status (active, inactive)
- `role` (optional): Filter by role (user, admin)

**Example**: `GET /api/v1/users/users?page=2&limit=20&status=active`

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
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 150,
      "total_pages": 15
    }
  }
}
```

### Get User by ID
Retrieve specific user information.

**Endpoint**: `GET /api/v1/users/users/{id}`  
**Authentication**: Required (Admin only)  
**Rate Limiting**: Standard

**Path Parameters**:
- `id`: User UUID

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
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**Error Responses**:
- `404 Not Found`: User not found
- `403 Forbidden`: Insufficient permissions

### Update User (Admin)
Update any user's information (admin only).

**Endpoint**: `PUT /api/v1/users/users/{id}`  
**Authentication**: Required (Admin only)  
**Rate Limiting**: Standard

**Path Parameters**:
- `id`: User UUID

**Request Body** (all fields optional):
```json
{
  "first_name": "Updated",
  "last_name": "Name",
  "role": "admin",
  "status": "inactive"
}
```

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
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### Delete User
Delete a user account.

**Endpoint**: `DELETE /api/v1/users/users/{id}`  
**Authentication**: Required (Admin only)  
**Rate Limiting**: Strict

**Path Parameters**:
- `id`: User UUID

**Response**: 200 OK
```json
{
  "success": true,
  "message": "User deleted successfully",
  "data": null
}
```

**Note**: This is a soft delete - the user is marked as inactive but not removed from the database.

## 📊 Data Models

### User Model
```json
{
  "id": "string (UUID)",
  "email": "string (unique)",
  "password": "string (hashed, not returned in responses)",
  "first_name": "string",
  "last_name": "string",
  "role": "string (user|admin)",
  "status": "string (active|inactive)",
  "email_verified": "boolean",
  "email_verified_at": "timestamp (nullable)",
  "password_changed_at": "timestamp (nullable)",
  "last_login_at": "timestamp (nullable)",
  "login_count": "integer",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### JWT Token Claims
```json
{
  "user_id": "string (UUID)",
  "email": "string",
  "role": "string",
  "iat": "integer (issued at)",
  "exp": "integer (expiration)",
  "iss": "string (issuer)"
}
```

## ⚠️ Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `UNAUTHORIZED` | 401 | Authentication required or failed |
| `INVALID_TOKEN` | 401 | JWT token is invalid or expired |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict (duplicate email) |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_SERVER_ERROR` | 500 | Internal server error |

## 🚦 Rate Limiting

Rate limiting is applied per client IP address:

| Endpoint Type | Limit | Window |
|---------------|--------|--------|
| Health checks | No limit | - |
| Authentication | 5 requests | 1 minute |
| Profile operations | 30 requests | 1 minute |
| Admin operations | 20 requests | 1 minute |
| General API | 100 requests | 1 minute |

**Rate limit headers** are included in responses:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining in current window
- `X-RateLimit-Reset`: Time when rate limit resets

## 🧪 Testing Examples

### Using cURL

#### Register User
```bash
curl -X POST http://localhost:8082/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8082/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### Get Profile (with token)
```bash
curl -X GET http://localhost:8082/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Admin: List Users
```bash
curl -X GET "http://localhost:8082/api/v1/users/users?page=1&limit=10" \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

### Using HTTPie

#### Register User
```bash
http POST localhost:8082/api/v1/users/register \
  email=test@example.com \
  password=password123 \
  first_name=Test \
  last_name=User
```

#### Get Profile
```bash
http GET localhost:8082/api/v1/users/profile \
  Authorization:"Bearer YOUR_JWT_TOKEN"
```

### JavaScript/Fetch Example

```javascript
// Register user
const registerUser = async () => {
  const response = await fetch('http://localhost:8082/api/v1/users/register', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      email: 'test@example.com',
      password: 'password123',
      first_name: 'Test',
      last_name: 'User'
    })
  });
  
  return await response.json();
};

// Login and get token
const login = async (email, password) => {
  const response = await fetch('http://localhost:8082/api/v1/users/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password })
  });
  
  const data = await response.json();
  return data.data.token; // Store this token
};

// Get profile with token
const getProfile = async (token) => {
  const response = await fetch('http://localhost:8082/api/v1/users/profile', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  return await response.json();
};
```

## 🔒 Security Considerations

1. **JWT Tokens**: Store securely on client-side, include in Authorization header
2. **Password Policy**: Minimum 8 characters, must contain letters and numbers
3. **Rate Limiting**: Respect rate limits to avoid blocking
4. **HTTPS**: Always use HTTPS in production
5. **Token Expiration**: Handle token refresh when tokens expire
6. **Input Validation**: API validates all input, but also validate on client-side

## 📝 Best Practices

1. **Error Handling**: Always check the `success` field in responses
2. **Pagination**: Use pagination for list endpoints to avoid large payloads
3. **Caching**: Cache user profiles on client-side to reduce API calls
4. **Logging**: Log API calls for debugging and monitoring
5. **Retry Logic**: Implement retry logic for transient errors
6. **Versioning**: API is versioned (`v1`), plan for future versions