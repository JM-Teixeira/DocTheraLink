# Development Guide

## Quick Start

### First Time Setup

```bash
# Clone the repository
git clone <repository-url>
cd DocTheraLink

# Start all services
docker-compose up --build -d

# Check if everything is running
docker ps
```

Access the application:
- **Frontend**: http://localhost:4000
- **Backend API**: http://localhost:9000
- **API Documentation**: http://localhost:9000/swagger/
- **Database Admin**: http://localhost:8080

### Test Credentials

All test users have the password: `1234@1`

Available users:
- `doctor` - Doctor role
- `nurse` - Nurse role
- `physiotherapist` - Physiotherapist role
- `stateadmin` - State Administrator
- `districtadmin` - District Administrator
- `facilitycoordinator` - Facility Coordinator
- `staff` - General Staff

## Development Workflow

### Making Changes

#### Backend Changes (Go)

When you modify backend code:

```bash
# Rebuild and restart backend
docker-compose build backend
docker-compose up -d backend

# Or rebuild everything
docker-compose up --build -d
```

**Note:** Backend changes require a rebuild because Go is compiled.

#### Frontend Changes (React)

When you modify frontend code:

```bash
# Rebuild and restart frontend
docker-compose build frontend
docker-compose up -d frontend
```

**Note:** Frontend changes require a rebuild because the production build is served by nginx.

### Hot Reload and Rebuilding

**When do you need to rebuild?**

| Change Type | Rebuild Required | Command |
|------------|------------------|---------|
| Frontend code (`.tsx`, `.ts`, `.css`) | ✅ Yes | `docker-compose build frontend && docker-compose up -d` |
| Backend code (`.go`) | ✅ Yes | `docker-compose build backend && docker-compose up -d` |
| Environment variables | ✅ Yes | `docker-compose up -d` (no build needed) |
| `docker-compose.yml` | ✅ Yes | `docker-compose up -d` |
| Database schema | ✅ Yes | `docker-compose down -v && docker-compose up --build -d` |
| Nginx config | ✅ Yes | `docker-compose build frontend && docker-compose up -d` |
| Package dependencies | ✅ Yes | `docker-compose build <service> && docker-compose up -d` |

### Local Development (Without Docker)

#### Backend

```bash
cd be

# Install dependencies
go mod download

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=care
export DB_PASSWORD=care
export DB_NAME=care
export JWT_SECRET=your-secret-key

# Run
go run cmd/server/main.go
```

#### Frontend

```bash
cd fe

# Install dependencies
npm install

# Set environment variables
export REACT_CARE_API_URL=http://localhost:9000

# Run development server
npm run dev
```

Access at http://localhost:4000

## Project Structure

```
DocTheraLink/
├── be/                      # Backend (Go)
│   ├── cmd/server/         # Main application entry
│   ├── internal/           # Internal packages
│   │   ├── api/           # API handlers and routes
│   │   ├── auth/          # Authentication logic
│   │   ├── config/        # Configuration
│   │   ├── database/      # Database connection and migrations
│   │   └── models/        # Data models
│   └── scripts/           # Utility scripts
├── fe/                     # Frontend (React + TypeScript)
│   ├── src/               # Source code
│   │   ├── components/   # React components
│   │   ├── hooks/        # Custom hooks
│   │   ├── Utils/        # Utilities
│   │   └── Routers/      # Routing
│   ├── nginx.conf        # Nginx configuration
│   └── Dockerfile        # Frontend Docker build
├── docs/                  # Documentation
└── docker-compose.yml    # Docker orchestration
```

## Common Tasks

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker logs care-backend -f
docker logs klyniva-frontend -f
docker logs care-postgres -f
```

### Restart Services

```bash
# All services
docker-compose restart

# Specific service
docker-compose restart backend
docker-compose restart frontend
```

### Stop Services

```bash
# Stop all
docker-compose down

# Stop and remove volumes (fresh start)
docker-compose down -v
```

### Access Database

```bash
# Using Adminer (Web UI)
# Open http://localhost:8080
# Server: postgres
# Username: care
# Password: care
# Database: care

# Using psql
docker exec -it care-postgres psql -U care -d care
```

### Run Database Migrations

Migrations run automatically on backend startup. To force a fresh migration:

```bash
docker-compose down -v
docker-compose up --build -d
```

### Add New Dependencies

#### Backend

```bash
cd be
go get <package-name>
go mod tidy

# Rebuild
docker-compose build backend
docker-compose up -d backend
```

#### Frontend

```bash
cd fe
npm install <package-name>

# Rebuild
docker-compose build frontend
docker-compose up -d frontend
```

## Testing

### Backend Tests

```bash
cd be
go test ./...
```

### Frontend Tests

```bash
cd fe
npm test
```

### API Testing

```bash
# Test login
curl -X POST http://localhost:4000/api/v1/auth/login/ \
  -H "Content-Type: application/json" \
  -d '{"username":"doctor","password":"1234@1"}'

# Test with token
TOKEN="<your-jwt-token>"
curl -X GET http://localhost:4000/api/v1/users/ \
  -H "Authorization: Bearer $TOKEN"
```

## Environment Variables

### Backend (.env or docker-compose.yml)

```bash
# Server
PORT=9000
HOST=0.0.0.0
GIN_MODE=debug

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=care
DB_PASSWORD=care
DB_NAME=care

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:4000
```

### Frontend (docker-compose.yml)

```bash
REACT_CARE_API_URL=http://localhost:4000
REACT_APP_TITLE=Klyniva
REACT_ALLOWED_LOCALES=en,pt
```

## Debugging

### Enable Debug Logs

Backend:
```yaml
# docker-compose.yml
environment:
  LOG_LEVEL: debug
  GIN_MODE: debug
```

### Check Container Health

```bash
# Health status
docker ps

# Inspect container
docker inspect care-backend

# Check resource usage
docker stats
```

## Code Style

### Backend (Go)

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `go vet` before committing

### Frontend (React/TypeScript)

- Follow the existing ESLint configuration
- Use Prettier for formatting
- Run `npm run lint` before committing

## Git Workflow

```bash
# Create feature branch
git checkout -b feature/your-feature

# Make changes and commit
git add .
git commit -m "feat: your feature description"

# Push and create PR
git push origin feature/your-feature
```

## Production Deployment

See [Deployment Guide](./DEPLOYMENT.md) (if available)

## Additional Resources

- [API Documentation](http://localhost:9000/swagger/)
- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://react.dev/)
