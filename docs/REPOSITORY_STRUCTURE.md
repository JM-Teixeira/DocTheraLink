# Repository Structure

## 📁 Directory Layout

```
DocTheraLink/
├── README.md                    # Main documentation
├── docker-compose.yml           # Docker orchestration
├── .gitignore                   # Git ignore rules
│
├── be/                          # Backend (Go)
│   ├── cmd/server/             # Application entry point
│   ├── internal/               # Internal packages
│   │   ├── api/               # API handlers & routes
│   │   ├── auth/              # Authentication
│   │   ├── config/            # Configuration
│   │   ├── database/          # Database & migrations
│   │   └── models/            # Data models
│   ├── scripts/               # SQL scripts
│   ├── Dockerfile             # Backend Docker build
│   ├── go.mod                 # Go dependencies
│   ├── go.sum                 # Go checksums
│   └── .env.example           # Environment template
│
├── fe/                         # Frontend (React + TypeScript)
│   ├── src/                   # Source code
│   │   ├── components/       # React components
│   │   ├── hooks/            # Custom hooks
│   │   ├── Utils/            # Utilities
│   │   ├── Routers/          # Routing
│   │   └── CAREUI/           # UI components
│   ├── public/               # Static assets
│   ├── apps/                 # Plugin applications
│   ├── scripts/              # Build scripts
│   ├── nginx.conf            # Nginx configuration
│   ├── Dockerfile            # Frontend Docker build
│   ├── docker-entrypoint.sh  # Runtime config script
│   ├── package.json          # Node dependencies
│   ├── vite.config.mts       # Vite configuration
│   └── .env.example          # Environment template
│
└── docs/                      # Documentation
    ├── QUICK_REFERENCE.md    # Cheat sheet
    ├── DEVELOPMENT.md        # Development guide
    ├── TROUBLESHOOTING.md    # Common issues
    ├── DEPLOYMENT.md         # Deployment guide
    └── REPOSITORY_STRUCTURE.md  # This file
```

## 📝 What Should Be Committed to Git

### ✅ Always Commit

**Source Code:**
- `be/cmd/`, `be/internal/` - Backend code
- `fe/src/` - Frontend code
- `fe/public/` - Static assets (images, fonts)

**Configuration:**
- `docker-compose.yml` - Docker orchestration
- `Dockerfile` (both be/ and fe/) - Docker builds
- `nginx.conf` - Web server config
- `.gitignore` - Git ignore rules
- `.env.example` - Environment templates

**Dependencies:**
- `package.json`, `package-lock.json` - Node dependencies
- `go.mod`, `go.sum` - Go dependencies

**Documentation:**
- `README.md` - Main docs
- `docs/*.md` - All documentation

**Scripts:**
- `be/scripts/*.sql` - Database scripts
- `fe/scripts/*.js` - Build scripts
- `fe/docker-entrypoint.sh` - Runtime scripts

**Legal:**
- `LICENSE` - License file
- `NOTICE` - Attribution notices

### ❌ Never Commit

**Secrets & Environment:**
- `.env` - Environment variables
- `*.pem`, `*.key`, `*.crt` - Certificates
- `secrets/` - Secret files

**Dependencies:**
- `node_modules/` - Node packages
- `vendor/` - Go vendor directory

**Build Artifacts:**
- `fe/build/`, `fe/dist/` - Frontend builds
- `be/main` - Backend binary
- `*.exe`, `*.dll`, `*.so` - Compiled binaries

**IDE & OS:**
- `.vscode/`, `.idea/` - IDE settings
- `.DS_Store`, `Thumbs.db` - OS files
- `.kiro/` - Kiro IDE files

**Logs & Temp:**
- `*.log` - Log files
- `tmp/`, `temp/` - Temporary files
- `uploads/` - User uploads (dev)

**Database:**
- `*.sql` (except scripts) - Database dumps
- `*.sqlite`, `*.db` - Database files

## 🔍 Finding Files

### Backend Code
```bash
# API handlers
be/internal/api/handlers/

# Routes
be/internal/api/routes.go

# Models
be/internal/models/

# Database
be/internal/database/
```

### Frontend Code
```bash
# Components
fe/src/components/

# Pages/Routes
fe/src/Routers/

# API calls
fe/src/Utils/request/

# Hooks
fe/src/hooks/
```

### Configuration
```bash
# Docker
docker-compose.yml
be/Dockerfile
fe/Dockerfile

# Environment
be/.env.example
fe/.env.example

# Web server
fe/nginx.conf
```

### Documentation
```bash
# Main docs
README.md

# Detailed guides
docs/DEVELOPMENT.md
docs/TROUBLESHOOTING.md
docs/QUICK_REFERENCE.md
```

## 📦 What Gets Built

### Frontend Build Process
```
fe/src/ → Vite Build → fe/build/ → Docker Image → Nginx serves static files
```

### Backend Build Process
```
be/cmd/, be/internal/ → Go Build → be/main → Docker Image → Binary runs
```

## 🚀 Deployment Artifacts

**Docker Images:**
- `doctheralink-backend:latest` - Backend service
- `doctheralink-frontend:latest` - Frontend service

**Volumes (Data):**
- `postgres_data` - Database data
- `redis_data` - Cache data
- `backend_uploads` - File uploads

## 🔄 Git Workflow

```bash
# Check what's changed
git status

# Stage files
git add .

# Commit
git commit -m "feat: your feature"

# Push
git push origin main
```

## 📊 Repository Size

**Typical sizes:**
- Source code: ~50MB
- node_modules: ~500MB (not committed)
- Docker images: ~1GB (not committed)
- Database: varies (not committed)

## 🧹 Cleaning Up

```bash
# Remove build artifacts
rm -rf fe/build fe/dist be/main

# Remove dependencies (can reinstall)
rm -rf fe/node_modules be/vendor

# Remove Docker artifacts
docker system prune -a
```

## 📚 Related Documentation

- [README.md](../README.md) - Getting started
- [DEVELOPMENT.md](./DEVELOPMENT.md) - Development workflow
- [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - Quick commands
