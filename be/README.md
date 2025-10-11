# CARE Backend - Go Implementation

This is a Go backend implementation that provides APIs compatible with the CARE frontend healthcare management system.

## Architecture

- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT tokens
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker support
- **Testing**: Go testing framework

## Project Structure

```
be/
├── cmd/
│   └── server/           # Application entry point
├── internal/
│   ├── api/             # HTTP handlers and routes
│   ├── auth/            # Authentication middleware
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and migrations
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   └── utils/           # Utility functions
├── pkg/                 # Public packages
├── migrations/          # Database migrations
├── docs/               # API documentation
├── docker/             # Docker configurations
├── scripts/            # Build and deployment scripts
└── tests/              # Integration tests
```

## Features

- RESTful API endpoints matching frontend requirements
- JWT-based authentication
- Role-based access control
- Database migrations
- API documentation with Swagger
- Docker containerization
- Comprehensive logging
- Error handling middleware
- Request validation
- CORS support

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 13+
- Docker (optional)

### Development Setup

1. Clone and navigate to backend:
```bash
cd be
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run database migrations:
```bash
go run cmd/migrate/main.go
```

5. Start the development server:
```bash
go run cmd/server/main.go
```

The API will be available at `http://localhost:9000`

### Docker Setup

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or build individual container
docker build -t care-backend .
docker run -p 9000:9000 care-backend
```

## API Endpoints

The backend provides the following main API groups:

### Authentication
- `POST /api/v1/auth/login/` - User login
- `POST /api/v1/auth/token/refresh/` - Refresh JWT token
- `POST /api/v1/auth/token/verify/` - Verify JWT token
- `POST /api/v1/password_reset/` - Request password reset
- `POST /api/v1/password_reset/confirm/` - Confirm password reset

### Users
- `GET /api/v1/users/` - List users
- `POST /api/v1/users/` - Create user
- `GET /api/v1/users/{username}/` - Get user details
- `PUT /api/v1/users/{username}/` - Update user
- `DELETE /api/v1/users/{username}/` - Delete user
- `GET /api/v1/users/getcurrentuser/` - Get current user

### Facilities
- `GET /api/v1/facility/` - List facilities
- `POST /api/v1/facility/` - Create facility
- `GET /api/v1/facility/{id}/` - Get facility details
- `PUT /api/v1/facility/{id}/` - Update facility
- `DELETE /api/v1/facility/{id}/` - Delete facility

### Patients
- `GET /api/v1/patient/` - List patients
- `POST /api/v1/patient/` - Create patient
- `GET /api/v1/patient/{id}/` - Get patient details
- `PUT /api/v1/patient/{id}/` - Update patient
- `POST /api/v1/patient/{id}/transfer/` - Transfer patient

### Consultations
- `GET /api/v1/consultation/` - List consultations
- `POST /api/v1/consultation/` - Create consultation
- `GET /api/v1/consultation/{id}/` - Get consultation details
- `PUT /api/v1/consultation/{id}/` - Update consultation

### Assets
- `GET /api/v1/asset/` - List assets
- `POST /api/v1/asset/` - Create asset
- `GET /api/v1/asset/{external_id}/` - Get asset details
- `PUT /api/v1/asset/{external_id}/` - Update asset

### Notifications
- `GET /api/v1/notification/` - List notifications
- `POST /api/v1/notification/notify/` - Send notification
- `PATCH /api/v1/notification/{id}/` - Mark as read/unread

## Configuration

Environment variables:

```env
# Server
PORT=9000
GIN_MODE=release

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=care
DB_PASSWORD=care
DB_NAME=care
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:4000,https://care.ohc.network

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## Testing

```bash
# Run unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./tests/...
```

## Deployment

### Docker Production

```bash
# Build production image
docker build -f docker/Dockerfile.prod -t care-backend:latest .

# Run with docker-compose
docker-compose -f docker-compose.prod.yml up -d
```

### Manual Deployment

```bash
# Build binary
go build -o bin/server cmd/server/main.go

# Run migrations
./bin/server migrate

# Start server
./bin/server
```

## API Documentation

API documentation is available at:
- Swagger UI: `http://localhost:9000/swagger/`
- OpenAPI JSON: `http://localhost:9000/swagger/doc.json`

## Contributing

1. Follow Go coding standards
2. Write tests for new features
3. Update API documentation
4. Use conventional commit messages
5. Ensure all tests pass before submitting PR

## License

MIT License - see LICENSE file for details