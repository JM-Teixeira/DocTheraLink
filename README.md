# DocTheraLink - Healthcare Management System

DocTheraLink is a comprehensive healthcare management system designed for physical therapy clinics and healthcare facilities. It provides patient management, therapist scheduling, treatment tracking, and administrative tools.

## 🏗️ System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │    Backend      │    │   Database      │
│   (React)       │◄──►│     (Go)        │◄──►│  (PostgreSQL)   │
│   Port: 4000    │    │   Port: 9000    │    │   Port: 5432    │
│   Nginx + React │    │   Gin + GORM    │    │   PostgreSQL    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 4GB RAM minimum
- 10GB free disk space

**For local development only:**

- Node.js 22.11.0+
- Go 1.21+

### Running the Application

1. **Clone and start**

```bash
git clone <repository-url>
cd DocTheraLink

# Start all services
docker-compose up --build -d
```

2. **Verify services are running**

```bash
docker ps
# You should see: care-backend, klyniva-frontend, care-postgres, care-redis, care-adminer
```

3. **Access the application**

- **Frontend (Main App)**: <http://localhost:4000>
- **Backend API**: <http://localhost:9000>
- **API Documentation**: <http://localhost:9000/swagger/>
- **Database Admin (Adminer)**: <http://localhost:8080>

4. **Login with test credentials**

- Username: `doctor`
- Password: `1234@1`

**That's it!** The application is now running.

## 🔄 Making Changes & Rebuilding

### When do I need to rebuild?

**Frontend changes** (React, TypeScript, CSS):

```bash
# After modifying files in fe/src/
docker-compose build frontend
docker-compose up -d frontend
```

**Backend changes** (Go code):

```bash
# After modifying files in be/
docker-compose build backend
docker-compose up -d backend
```

**Configuration changes** (docker-compose.yml, .env):

```bash
# Just restart, no rebuild needed
docker-compose up -d
```

**Database schema changes**:

```bash
# Full reset (⚠️ deletes all data)
docker-compose down -v
docker-compose up --build -d
```

### Quick rebuild everything

```bash
docker-compose up --build -d
```

### View changes in real-time

```bash
# Watch logs while developing
docker-compose logs -f backend
docker-compose logs -f frontend
```

## 💻 Local Development (Without Docker)

For faster development with hot-reload:

**Backend:**

```bash
cd be
go mod download
export REACT_CARE_API_URL=http://localhost:9000
go run cmd/server/main.go
```

**Frontend:**

```bash
cd fe
npm install
export REACT_CARE_API_URL=http://localhost:9000
npm run dev
```

**Note:** You still need Docker for PostgreSQL and Redis:

```bash
docker-compose up -d postgres redis
```

## 🧪 Testing the Application

### Default Test Users

The system comes with pre-configured test users:

| Username | Password | Role | Description |
|----------|----------|------|-------------|
| `stateadmin` | `1234@1` | State Admin | Full system access |
| `districtadmin` | `1234@1` | District Admin | District-level access |
| `facilitycoordinator` | `1234@1` | Local Body Admin | Facility coordination |
| `doctor` | `1234@1` | Doctor | Medical professional |
| `physiotherapist` | `1234@1` | PhysioTherapist | Physical therapy |
| `nurse` | `1234@1` | Nurse | Nursing staff |
| `staff` | `1234@1` | Staff | General staff |

**⚠️ Important**: These are test accounts. Change passwords in production!

### Testing Workflow

1. **Login**: Use any of the test accounts above
2. **Facility Management**: Create and manage facilities
3. **User Management**: Add therapists and staff
4. **Patient Management**: Register patients and assign treatments
5. **Scheduling**: Create timetables and assign therapists
6. **Treatment Tracking**: Record treatment sessions and notes

### API Testing

```bash
# Health check
curl http://localhost:9000/health

# Login
curl -X POST http://localhost:9000/api/v1/auth/login/ \
  -H "Content-Type: application/json" \
  -d '{"username": "doctor", "password": "1234@1"}'

# Get current user (replace TOKEN with actual token)
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:9000/api/v1/users/getcurrentuser/
```

## 🛠️ Common Commands

### Managing Services

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# Restart a specific service
docker-compose restart backend
docker-compose restart frontend

# View logs
docker-compose logs -f
docker-compose logs -f backend  # specific service

# Check status
docker ps
docker-compose ps
```

### Development

```bash
# Rebuild after code changes
docker-compose build <service>
docker-compose up -d <service>

# Full rebuild
docker-compose up --build -d

# Fresh start (deletes all data!)
docker-compose down -v
docker-compose up --build -d
```

### Frontend Commands (local development)

```bash
cd fe
npm run dev          # Development server
npm run build        # Production build
npm run lint         # Check code quality
npm run format       # Format code
```

### Backend Commands (local development)

```bash
cd be
go run cmd/server/main.go  # Run server
go test ./...              # Run tests
go build -o main cmd/server/main.go  # Build binary
```

## 📊 Database Management

### Accessing the Database

**Via Adminer (Web Interface)**:

- URL: <http://localhost:8080>
- System: PostgreSQL
- Server: postgres
- Username: care
- Password: care
- Database: care

**Via Command Line**:

```bash
# Connect to database container
docker exec -it care-postgres psql -U care -d care

# Or from host (if PostgreSQL client installed)
psql -h localhost -U care -d care
```

### Database Operations

```bash
# Backup database
docker exec care-postgres pg_dump -U care care > backup.sql

# Restore database
docker exec -i care-postgres psql -U care -d care < backup.sql

# Reset database
docker-compose down postgres
docker volume rm doctheralink_postgres_data
docker-compose up -d postgres
cd be && make migrate-seed
```

## 🔧 Configuration

### Environment Variables

**Backend (.env)**:

```env
# Server
PORT=9000
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

**Frontend (.env)**:

```env
# Backend API
REACT_CARE_API_URL=http://localhost:9000

# Optional customization
REACT_CUSTOM_DESCRIPTION=DocTheraLink Healthcare System
```

## 🚀 Production Deployment

### Docker Production

```bash
# Build production images
docker-compose -f docker-compose.prod.yml build

# Deploy
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes Ready

The application is designed to be Kubernetes-ready with:

- Separate containers for each service
- Environment-based configuration
- Health check endpoints
- Graceful shutdown handling
- Persistent volume support for database

See `k8s/` directory for Kubernetes manifests (coming soon).

## 📝 Monitoring and Logs

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres
```

### Health Checks

```bash
# Backend health
curl http://localhost:9000/health

# Database health
docker exec care-postgres pg_isready -U care
```

## 🆘 Troubleshooting

### Login not working?

```bash
# Test the API directly
curl -X POST http://localhost:4000/api/v1/auth/login/ \
  -H "Content-Type: application/json" \
  -d '{"username":"doctor","password":"1234@1"}'

# Should return JWT tokens
```

### Changes not showing up?

```bash
# Rebuild the service you changed
docker-compose build frontend  # or backend
docker-compose up -d
```

### Port already in use?

```bash
# Find what's using the port
lsof -i :4000  # or :9000, :5432

# Change ports in docker-compose.yml
```

### Container keeps restarting?

```bash
# Check logs for errors
docker logs care-backend
docker logs klyniva-frontend
```

### Fresh start (nuclear option)

```bash
# ⚠️ This deletes ALL data
docker-compose down -v
docker-compose up --build -d
```

**For more help:** See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)

## 📚 Documentation

- **[Quick Reference](docs/QUICK_REFERENCE.md)** - One-page cheat sheet
- **[Development Guide](docs/DEVELOPMENT.md)** - Detailed development workflow
- **[Troubleshooting Guide](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Repository Structure](docs/REPOSITORY_STRUCTURE.md)** - What to commit to Git
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment
- **[API Documentation](http://localhost:9000/swagger/)** - Interactive API docs (when running)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
