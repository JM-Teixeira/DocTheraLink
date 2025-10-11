# CARE Backend API Mapping

This document maps the frontend API requirements to the Go backend implementation, providing a comprehensive overview of all endpoints that need to be implemented to support the CARE frontend.

## Overview

The CARE backend provides a RESTful API that matches the frontend's expectations. All endpoints follow the same URL patterns and request/response formats as the original Python backend to ensure seamless integration.

## Authentication & Authorization

### JWT Token-based Authentication

- **Access Token**: Short-lived token for API access (24h default)
- **Refresh Token**: Long-lived token for renewing access (7 days default)
- **Role-based Access Control**: Different user types have different permissions

### User Roles

- `Pharmacist` - Pharmacy management
- `Volunteer` - Basic volunteer access
- `StaffReadOnly` - Read-only staff access
- `Staff` - General staff access
- `Nurse` - Nursing staff access
- `PhysioTherapist` - Physiotherapy access
- `Doctor` - Medical doctor access
- `WardAdmin` - Ward administration
- `LocalBodyAdmin` - Local body administration
- `DistrictLabAdmin` - District lab administration
- `DistrictReadOnlyAdmin` - Read-only district admin
- `DistrictAdmin` - District administration
- `StateLabAdmin` - State lab administration
- `StateReadOnlyAdmin` - Read-only state admin
- `StateAdmin` - State administration

## API Endpoints Mapping

### Authentication Endpoints

| Frontend Endpoint | Backend Implementation | Status | Description |
|------------------|----------------------|--------|-------------|
| `POST /api/v1/auth/login/` | ✅ Implemented | Complete | User login with JWT tokens |
| `POST /api/v1/auth/token/refresh/` | ✅ Implemented | Complete | Refresh access token |
| `POST /api/v1/auth/token/verify/` | ✅ Implemented | Complete | Verify token validity |
| `POST /api/v1/password_reset/` | ✅ Implemented | Partial | Request password reset |
| `POST /api/v1/password_reset/check/` | ✅ Implemented | Partial | Check reset token |
| `POST /api/v1/password_reset/confirm/` | ✅ Implemented | Partial | Confirm password reset |
| `PUT /api/v1/password_change/` | ✅ Implemented | Complete | Change user password |

### User Management Endpoints

| Frontend Endpoint | Backend Implementation | Status | Description |
|------------------|----------------------|--------|-------------|
| `GET /api/v1/users/` | ✅ Implemented | Complete | List users with pagination |
| `POST /api/v1/users/` | ✅ Implemented | Complete | Create new user |
| `GET /api/v1/users/getcurrentuser/` | ✅ Implemented | Complete | Get current user details |
| `GET /api/v1/users/{username}/` | ✅ Implemented | Complete | Get user by username |
| `PUT /api/v1/users/{username}/` | ✅ Implemented | Complete | Update user |
| `PATCH /api/v1/users/{username}/` | ✅ Implemented | Complete | Partial update user |
| `DELETE /api/v1/users/{username}/` | ✅ Implemented | Complete | Delete user |
| `GET /api/v1/users/{username}/check_availability/` | ✅ Implemented | Complete | Check username availability |
| `GET /api/v1/users/{username}/skill/` | ✅ Implemented | Complete | List user skills |
| `POST /api/v1/users/{username}/skill/` | ✅ Implemented | Complete | Add user skill |
| `DELETE /api/v1/users/{username}/skill/{id}/` | ✅ Implemented | Complete | Remove user skill |
| `GET /api/v1/users/{username}/get_facilities/` | 🔄 Placeholder | Pending | List user facilities |
| `PUT /api/v1/users/{username}/add_facility/` | 🔄 Placeholder | Pending | Add user to facility |
| `DELETE /api/v1/users/{username}/delete_facility/` | 🔄 Placeholder | Pending | Remove user from facility |
| `DELETE /api/v1/users/{username}/clear_home_facility/` | 🔄 Placeholder | Pending | Clear home facility |
| `PATCH /api/v1/users/{username}/profile_picture/` | 🔄 Placeholder | Pending | Update profile picture |
| `DELETE /api/v1/users/{username}/profile_picture/` | 🔄 Placeholder | Pending | Delete profile picture |
| `GET /api/v1/users/{username}/pnconfig/` | 🔄 Placeholder | Pending | Get push notification config |
| `PATCH /api/v1/users/{username}/pnconfig/` | 🔄 Placeholder | Pending | Update push notification config |

### Skills Endpoints

| Frontend Endpoint | Backend Implementation | Status | Description |
|------------------|----------------------|--------|-------------|
| `GET /api/v1/skill/` | ✅ Implemented | Complete | List all available skills |

### Facility Management Endpoints

| Frontend Endpoint | Backend Implementation | Status | Description |
|------------------|----------------------|--------|-------------|
| `GET /api/v1/facility/` | ✅ Implemented | Complete | List facilities with pagination |
| `POST /api/v1/facility/` | ✅ Implemented | Complete | Create new facility |
| `GET /api/v1/facility/{id}/` | ✅ Implemented | Complete | Get facility details |
| `PUT /api/v1/facility/{id}/` | ✅ Implemented | Complete | Update facility |
| `PATCH /api/v1/facility/{id}/` | ✅ Implemented | Complete | Partial update facility |
| `DELETE /api/v1/facility/{id}/` | ✅ Implemented | Complete | Delete facility |
| `GET /api/v1/getallfacilities/` | ✅ Implemented | Complete | Public facility list |
| `GET /api/v1/getallfacilities/{id}/` | ✅ Implemented | Complete | Public facility details |
| `GET /api/v1/facility/{facility_id}/get_users/` | 🔄 Placeholder | Pending | Get facility users |

### Geographic Data Endpoints

| Frontend Endpoint | Backend Implementation | Status | Description |
|------------------|----------------------|--------|-------------|
| `GET /api/v1/state/` | ✅ Implemented | Complete | List all states |
| `GET /api/v1/state/{id}/` | ✅ Implemented | Complete | Get state details |
| `GET /api/v1/state/{id}/districts/` | 🔄 Placeholder | Pending | Get districts by state |
| `GET /api/v1/district/` | 🔄 Placeholder | Pending | List districts |
| `GET /api/v1/district/{id}/` | 🔄 Placeholder | Pending | Get district details |
| `GET /api/v1/district/{id}/get_all_local_body/` | 🔄 Placeholder | Pending | Get local bodies by district |
| `GET /api/v1/district/{id}/local_bodies/` | 🔄 Placeholder | Pending | Get local bodies by district |
| `GET /api/v1/local_body/` | 🔄 Placeholder | Pending | List local bodies |
| `GET /api/v1/local_body/{id}/` | 🔄 Placeholder | Pending | Get local body details |
| `GET /api/v1/ward/` | 🔄 Placeholder | Pending | List wards |
| `GET /api/v1/ward/{id}/` | 🔄 Placeholder | Pending | Get ward details |

### Patient Management Endpoints

| Frontend Endpoint | Backend Implementation | Status | Description |
|------------------|----------------------|--------|-------------|
| `GET /api/v1/patient/` | 🔄 Placeholder | Pending | List patients |
| `POST /api/v1/patient/` | 🔄 Placeholder | Pending | Create patient |
| `GET /api/v1/patient/search/` | 🔄 Placeholder | Pending | Search patients |
| `GET /api/v1/patient/{id}/` | 🔄 Placeholder | Pending | Get patient details |
| `PUT /api/v1/patient/{id}/` | 🔄 Placeholder | Pending | Update patient |
| `PATCH /api/v1/patient/{id}/` | 🔄 Placeholder | Pending | Partial update patient |
| `POST /api/v1/patient/{id}/transfer/` | 🔄 Placeholder | Pending | Transfer patient |
| `GET /api/v1/patient/{patientId}/notes/` | 🔄 Placeholder | Pending | Get patient notes |
| `POST /api/v1/patient/{patientId}/notes/` | 🔄 Placeholder | Pending | Add patient note |
| `PUT /api/v1/patient/{patientId}/notes/{noteId}/` | 🔄 Placeholder | Pending | Update patient note |
| `GET /api/v1/patient/{patientId}/notes/{noteId}/edits/` | 🔄 Placeholder | Pending | Get note edit history |

### Consultation Endpoints

| Frontend Endpoint | Backend Implementation | Status | Description |
|------------------|----------------------|--------|-------------|
| `GET /api/v1/consultation/` | 🔄 Placeholder | Pending | List consultations |
| `POST /api/v1/consultation/` | 🔄 Placeholder | Pending | Create consultation |
| `GET /api/v1/consultation/{id}/` | 🔄 Placeholder | Pending | Get consultation details |
| `PUT /api/v1/consultation/{id}/` | 🔄 Placeholder | Pending | Update consultation |
| `PATCH /api/v1/consultation/{id}/` | 🔄 Placeholder | Pending | Partial update consultation |
| `DELETE /api/v1/consultation/{id}/` | 🔄 Placeholder | Pending | Delete consultation |

### Daily Rounds Endpoints

| Frontend Endpoint | Backend Implementation | Status | Description |
|------------------|----------------------|--------|-------------|
| `GET /api/v1/consultation/{consultationId}/daily_rounds/` | 🔄 Placeholder | Pending | Get daily rounds |
| `POST /api/v1/consultation/{consultationId}/daily_rounds/` | 🔄 Placeholder | Pending | Create daily round |
| `GET /api/v1/consultation/{consultationId}/daily_rounds/{id}/` | 🔄 Placeholder | Pending | Get daily round details |
| `PUT /api/v1/consultation/{consultationId}/daily_rounds/{id}/` | 🔄 Placeholder | Pending | Update daily round |
| `PATCH /api/v1/consultation/{consultationId}/daily_rounds/{id}/` | 🔄 Placeholder | Pending | Partial update daily round |

### Asset Management Endpoints

| Frontend Endpoint | Backend Implementation | Status | Description |
|------------------|----------------------|--------|-------------|
| `GET /api/v1/asset/` | 🔄 Placeholder | Pending | List assets |
| `POST /api/v1/asset/` | 🔄 Placeholder | Pending | Create asset |
| `GET /api/v1/asset/{external_id}/` | 🔄 Placeholder | Pending | Get asset details |
| `PUT /api/v1/asset/{external_id}/` | 🔄 Placeholder | Pending | Update asset |
| `PATCH /api/v1/asset/{external_id}/` | 🔄 Placeholder | Pending | Partial update asset |
| `DELETE /api/v1/asset/{external_id}/` | 🔄 Placeholder | Pending | Delete asset |
| `GET /api/v1/public/asset_qr/{qr_code_id}/` | 🔄 Placeholder | Pending | Get asset by QR code |

### Notification Endpoints

| Frontend Endpoint | Backend Implementation | Status | Description |
|------------------|----------------------|--------|-------------|
| `GET /api/v1/notification/` | 🔄 Placeholder | Pending | List notifications |
| `GET /api/v1/notification/{id}/` | 🔄 Placeholder | Pending | Get notification details |
| `PATCH /api/v1/notification/{id}/` | 🔄 Placeholder | Pending | Mark notification as read |
| `POST /api/v1/notification/notify/` | 🔄 Placeholder | Pending | Send notification |

## Data Models

### Core Models Implemented

1. **User Model** - Complete with roles, skills, and relationships
2. **Facility Model** - Complete with geographic relationships
3. **Patient Model** - Complete with medical history and relationships
4. **Consultation Model** - Complete with daily rounds and events
5. **Asset Model** - Complete with location and service tracking
6. **Geographic Models** - State, District, LocalBody, Ward

### Key Features

- **UUID Primary Keys** - All main entities use UUIDs for external references
- **Soft Deletes** - Important entities support soft deletion
- **Audit Fields** - Created/updated timestamps and user tracking
- **Relationships** - Proper foreign key relationships with preloading
- **Validation** - Request validation using Gin binding
- **Pagination** - Consistent pagination across list endpoints

## Database Schema

### PostgreSQL with GORM

- **Auto-migrations** - Database schema is automatically managed
- **Indexes** - Proper indexing for performance
- **Constraints** - Foreign key constraints and validations
- **Extensions** - UUID and other PostgreSQL extensions

### Seeded Data

- States and districts (Kerala focus)
- Basic skills for healthcare workers
- Inventory items for facility management
- Facility types and categories

## Security Features

- **JWT Authentication** - Secure token-based authentication
- **Role-based Access Control** - Different permissions per user type
- **Password Hashing** - Bcrypt for secure password storage
- **CORS Configuration** - Proper cross-origin resource sharing
- **Request Validation** - Input validation and sanitization

## Development Features

- **Swagger Documentation** - Auto-generated API documentation
- **Docker Support** - Complete containerization setup
- **Hot Reload** - Development server with live reload
- **Database Migrations** - Automated database setup
- **Comprehensive Logging** - Structured logging with levels
- **Health Checks** - Service health monitoring

## Deployment Ready

- **Production Dockerfile** - Optimized production container
- **Docker Compose** - Complete development environment
- **Environment Configuration** - Flexible configuration management
- **Graceful Shutdown** - Proper server shutdown handling
- **Static File Serving** - File upload and serving capabilities

## Next Steps for Full Implementation

1. **Complete Patient Management** - Implement all patient-related endpoints
2. **Consultation System** - Full consultation and daily rounds functionality
3. **Asset Management** - Complete asset tracking and QR code system
4. **Notification System** - Push notifications and messaging
5. **File Upload System** - Document and image management
6. **Inventory Management** - Stock tracking and management
7. **Reporting System** - Data export and reporting features
8. **Integration Testing** - Comprehensive test coverage
9. **Performance Optimization** - Database query optimization
10. **Production Deployment** - CI/CD pipeline and monitoring

## API Response Format

All endpoints follow consistent response formats:

### Success Response

```json
{
  "count": 100,
  "next": "http://api/endpoint?offset=20",
  "previous": "http://api/endpoint?offset=0",
  "results": [...]
}
```

### Error Response

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": {
    "field": "Field specific error"
  }
}
```

This backend implementation provides a solid foundation that matches the frontend's expectations while following Go best practices and providing room for future enhancements.
