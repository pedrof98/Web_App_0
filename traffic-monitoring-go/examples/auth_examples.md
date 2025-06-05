# V2X SIEM Authentication Examples

This document provides examples of how to use the authentication endpoints in the V2X SIEM API.

## Default Admin User

A default admin user is created during database migration:
- **Email**: `admin@v2x-siem.com`
- **Password**: `admin123`
- **Role**: `admin`

## Authentication Flow

### 1. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@v2x-siem.com",
    "password": "admin123"
  }'
```

Response:
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "email": "admin@v2x-siem.com",
      "role": "admin",
      "first_name": "System",
      "last_name": "Administrator",
      "full_name": "System Administrator",
      "is_active": true,
      "created_at": "2025-05-30T12:00:00Z",
      "updated_at": "2025-05-30T12:00:00Z"
    }
  }
}
```

### 2. Register New User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "analyst@example.com",
    "password": "securepassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### 3. Using Authentication Token

Include the token in the Authorization header for protected endpoints:

```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 4. Get Current User Profile

```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 5. Update Profile

```bash
curl -X PUT http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane",
    "last_name": "Smith"
  }'
```

### 6. Change Password

```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "oldpassword",
    "new_password": "newpassword123"
  }'
```

## User Management (Admin Only)

### List Users

```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer ADMIN_TOKEN_HERE"
```

### Create User (Admin Only)

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer ADMIN_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "operator@example.com",
    "password": "password123",
    "first_name": "Bob",
    "last_name": "Wilson",
    "role": "operator"
  }'
```

### Deactivate User (Admin Only)

```bash
curl -X POST http://localhost:8080/api/v1/users/2/deactivate \
  -H "Authorization: Bearer ADMIN_TOKEN_HERE"
```

## Role-Based Access Control

The system implements role-based access control with the following roles:

- **admin**: Full access to all endpoints
- **analyst**: Can manage rules and alerts, view events
- **operator**: Can manage alerts, view rules and events
- **viewer**: Can only view data (read-only access)

### Endpoint Permissions

| Endpoint | Admin | Analyst | Operator | Viewer |
|----------|-------|---------|----------|--------|
| GET /rules | ✓ | ✓ | ✓ | ✓ |
| POST /rules | ✓ | ✓ | ✗ | ✗ |
| PUT /rules | ✓ | ✓ | ✗ | ✗ |
| DELETE /rules | ✓ | ✓ | ✗ | ✗ |
| GET /alerts | ✓ | ✓ | ✓ | ✓ |
| POST /alerts | ✓ | ✓ | ✓ | ✗ |
| PUT /alerts | ✓ | ✓ | ✓ | ✗ |
| DELETE /alerts | ✓ | ✓ | ✗ | ✗ |
| POST /alerts/:id/assign | ✓ | ✓ | ✓ | ✗ |
| GET /security-events | ✓ | ✓ | ✓ | ✓ |
| POST /security-events | ✓ | ✓ | ✗ | ✗ |
| DELETE /security-events | ✓ | ✗ | ✗ | ✗ |
| /users/* | ✓ | ✗ | ✗ | ✗ |

## Error Responses

### Invalid Credentials
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "invalid email or password"
  }
}
```

### Missing Token
```json
{
  "error": {
    "code": "MISSING_TOKEN",
    "message": "Authorization header is required"
  }
}
```

### Expired Token
```json
{
  "error": {
    "code": "EXPIRED_TOKEN",
    "message": "token has expired"
  }
}
```

### Insufficient Permissions
```json
{
  "error": {
    "code": "INSUFFICIENT_PERMISSIONS",
    "message": "Insufficient permissions"
  }
}
```

## Environment Variables

For production deployment, make sure to set these environment variables:

```bash
# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_DURATION=24h

# Database Configuration
DSN=host=localhost user=postgres password=password dbname=siem port=5432 sslmode=disable

# Server Configuration
PORT=8080
GO_ENV=production
LOG_LEVEL=info
```

## Security Considerations

1. **JWT Secret**: Use a strong, random secret key in production
2. **Password Hashing**: Passwords are hashed using bcrypt with a cost of 10
3. **Token Expiration**: Tokens expire after 24 hours by default
4. **HTTPS**: Always use HTTPS in production
5. **Rate Limiting**: Consider implementing rate limiting for authentication endpoints
6. **Input Validation**: All inputs are validated using Go's validator package