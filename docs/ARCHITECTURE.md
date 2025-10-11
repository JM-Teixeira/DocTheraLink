# DocTheraLink System Architecture

This document provides a comprehensive overview of the DocTheraLink system architecture, explaining the purpose and function of each component, file, and directory.

## 📁 Project Structure Overview

```
DocTheraLink/
├── fe/                          # Frontend React Application
├── be/                          # Backend Go API Server
├── k8s/                         # Kubernetes Deployment Manifests
├── docs/                        # Documentation
├── docker-compose.yml           # Development Docker Setup
├── docker-compose.prod.yml      # Production Docker Setup
├── README.md                    # Main Project Documentation
├── DEPLOYMENT.md               # Deployment Instructions
└── .gitignore                  # Git Ignore Rules
```

## 🎨 Frontend Architecture (fe/)

The frontend is a React 18.3.1 application built with TypeScript, using modern development practices and tools.

### Core Structure

```
fe/
├── src/                        # Source Code
│   ├── components/            # React Components
│   ├── hooks/                 # Custom React Hooks
│   ├── lib/                   # Core Libraries
│   ├── Utils/                 # Utility Functions
│   ├── CAREUI/               # Custom UI Components
│   ├── Providers/            # React Context Providers
│   ├── Routers/              # Routing Configuration
│   ├── Locale/               # Internationalization
│   ├── common/               # Common Constants & Utils
│   ├── style/                # Global Styles
│   ├── App.tsx               # Main App Component
│   └── index.tsx             # Application Entry Point
├── public/                    # Static Assets
├── apps/                      # Plugin System (Modular Apps)
├── cypress/                   # End-to-End Tests
├── scripts/                   # Build & Utility Scripts
├── package.json              # Dependencies & Scripts
├── vite.config.mts           # Vite Build Configuration
├── tailwind.config.js        # Tailwind CSS Configuration
├── tsconfig.json             # TypeScript Configuration
├── .eslintrc.json            # ESLint Rules
├── .prettierrc.json          # Prettier Formatting
└── Dockerfile                # Container Build Instructions
```

### Key Frontend Files Explained

#### Configuration Files
- **`package.json`**: Defines dependencies, scripts, and project metadata
- **`vite.config.mts`**: Vite bundler configuration with plugins, aliases, and build settings
- **`tsconfig.json`**: TypeScript compiler configuration with path mappings
- **`tailwind.config.js`**: Tailwind CSS configuration with custom colors and themes
- **`.eslintrc.json`**: Code linting rules for TypeScript and React
- **`.prettierrc.json`**: Code formatting rules with import sorting

#### Core Application Files
- **`src/App.tsx`**: Main application component with routing and providers
- **`src/index.tsx`**: Application entry point that renders the React app
- **`src/PluginEngine.tsx`**: Dynamic plugin loading system
- **`src/service-worker.ts`**: PWA service worker for offline functionality

#### Component Architecture
- **`src/components/`**: Feature-based component organization
  - `Auth/`: Authentication components (login, password reset)
  - `Patient/`: Patient management components
  - `Facility/`: Healthcare facility management
  - `Users/`: User management and profiles
  - `Common/`: Shared/reusable components
  - `Form/`: Form components and validation
  - `ui/`: Base UI components (shadcn/ui)

#### Custom Libraries
- **`src/CAREUI/`**: Custom UI component library
  - `display/`: Display components (cards, badges, etc.)
  - `icons/`: Icon components
  - `misc/`: Miscellaneous UI utilities

#### Utilities & Hooks
- **`src/hooks/`**: Custom React hooks for reusable logic
- **`src/Utils/`**: Business logic utilities and helpers
- **`src/lib/`**: Core libraries (API client, utilities)
- **`src/common/`**: Constants, validation, and shared utilities

#### Routing & State
- **`src/Routers/`**: Application routing configuration
- **`src/Providers/`**: React context providers for global state

## 🔧 Backend Architecture (be/)

The backend is a Go application using the Gin framework, following clean architecture principles.

### Core Structure

```
be/
├── cmd/                       # Application Entry Points
│   ├── server/               # Main Server Application
│   └── migrate/              # Database Migration Tool
├── internal/                 # Private Application Code
│   ├── api/                  # HTTP API Layer
│   ├── auth/                 # Authentication & Authorization
│   ├── config/               # Configuration Management
│   ├── database/             # Database Connection & Migrations
│   ├── models/               # Data Models
│   ├── services/             # Business Logic Layer
│   └── utils/                # Internal Utilities
├── pkg/                      # Public Packages (if any)
├── docs/                     # API Documentation
├── scripts/                  # Build & Deployment Scripts
├── migrations/               # Database Migration Files
├── go.mod                    # Go Module Definition
├── go.sum                    # Go Module Checksums
├── Dockerfile                # Container Build Instructions
├── docker-compose.yml        # Development Environment
├── Makefile                  # Build Automation
└── .env.example              # Environment Variables Template
```

### Key Backend Files Explained

#### Application Entry Points (`cmd/`)
- **`cmd/server/main.go`**: Main HTTP server with middleware, routing, and graceful shutdown
- **`cmd/migrate/main.go`**: Database migration tool for schema updates and seeding

#### Configuration (`internal/config/`)
- **`config.go`**: Environment-based configuration management with validation

#### Database Layer (`internal/database/`)
- **`database.go`**: Database connection, migration runner, and seeding logic

#### Data Models (`internal/models/`)
- **`user.go`**: User model with roles, authentication, and relationships
- **`facility.go`**: Healthcare facility model with geographic data
- **`patient.go`**: Patient model with medical history and treatment data
- **`consultation.go`**: Medical consultation and daily rounds models
- **`therapy.go`**: Physical therapy specific models (sessions, schedules)
- **`common.go`**: Shared models (pagination, responses, file uploads)

#### Authentication (`internal/auth/`)
- **`jwt.go`**: JWT token generation, validation, and refresh logic
- **`middleware.go`**: Authentication middleware and role-based access control

#### API Layer (`internal/api/`)
- **`routes.go`**: HTTP route definitions and middleware setup
- **`handlers/`**: HTTP request handlers organized by domain
  - `auth.go`: Authentication endpoints
  - `user.go`: User management endpoints
  - `facility.go`: Facility management endpoints
  - `patient.go`: Patient management endpoints
  - `therapy.go`: Physical therapy endpoints

#### Business Logic (`internal/services/`)
- **`user_service.go`**: User management business logic
- **`therapy_service.go`**: Physical therapy scheduling and management
- **`notification_service.go`**: Notification system logic

## 🗄️ Database Architecture

### PostgreSQL Schema Design

The system uses PostgreSQL with the following key tables:

#### Core Tables
- **`users`**: System users with roles and authentication
- **`facilities`**: Healthcare facilities with geographic relationships
- **`patients`**: Patient records with medical history
- **`consultations`**: Medical consultations and treatments
- **`daily_rounds`**: Daily medical observations and notes

#### Physical Therapy Tables
- **`therapy_sessions`**: Individual therapy sessions
- **`therapy_schedules`**: Therapist schedules and availability
- **`facility_sections`**: Configurable facility sections (pool, gym, etc.)
- **`patient_treatments`**: Treatment plans and progress tracking

#### Geographic Tables
- **`states`**: State/province data
- **`districts`**: District/county data
- **`local_bodies`**: Local administrative bodies
- **`wards`**: Ward/neighborhood data

#### Supporting Tables
- **`skills`**: Available skills for healthcare workers
- **`user_skills`**: User-skill relationships
- **`notifications`**: System notifications
- **`file_uploads`**: File and document management

### Database Features
- **UUID Primary Keys**: For external references and security
- **Soft Deletes**: Important records are marked as deleted, not removed
- **Audit Trails**: Created/updated timestamps and user tracking
- **Relationships**: Proper foreign key constraints with cascading
- **Indexes**: Optimized for common query patterns

## 🐳 Containerization Architecture

### Multi-Container Setup

The application is designed as a microservices architecture with separate containers:

#### Frontend Container
```dockerfile
# Multi-stage build for optimized production image
FROM node:18-alpine AS builder
# Build React application
FROM nginx:alpine AS production
# Serve static files with Nginx
```

#### Backend Container
```dockerfile
# Multi-stage build for minimal Go binary
FROM golang:1.21-alpine AS builder
# Build Go application
FROM alpine:latest AS production
# Run single binary with minimal dependencies
```

#### Database Container
```dockerfile
# PostgreSQL with initialization scripts
FROM postgres:15-alpine
# Custom initialization and extensions
```

### Container Communication
- **Frontend → Backend**: HTTP API calls via environment-configured URL
- **Backend → Database**: PostgreSQL connection via Docker network
- **Shared Network**: All containers communicate via Docker network
- **Volume Persistence**: Database data persisted via Docker volumes

## 🔄 Data Flow Architecture

### Request Flow
```
User Browser → Frontend (React) → Backend API (Go) → Database (PostgreSQL)
     ↓              ↓                    ↓                    ↓
  UI Updates ← JSON Response ← Business Logic ← Data Query
```

### Authentication Flow
```
1. User Login → Backend validates credentials
2. Backend generates JWT tokens (access + refresh)
3. Frontend stores tokens securely
4. Subsequent requests include Bearer token
5. Backend validates token and extracts user context
6. Role-based access control applied
```

### Physical Therapy Workflow
```
1. Coordinator creates facility sections
2. Therapists are assigned to sections
3. Patients are registered and assigned treatments
4. Schedules are created with therapist-patient-section mapping
5. Treatment sessions are recorded with notes
6. Progress tracking and notifications
```

## 🔐 Security Architecture

### Authentication & Authorization
- **JWT Tokens**: Stateless authentication with access/refresh pattern
- **Role-Based Access Control**: Different permissions per user type
- **Password Security**: Bcrypt hashing with configurable cost
- **Token Expiration**: Configurable token lifetimes

### API Security
- **CORS Configuration**: Controlled cross-origin access
- **Request Validation**: Input validation and sanitization
- **Rate Limiting**: Protection against abuse
- **Security Headers**: XSS, CSRF, and other protections

### Data Security
- **Database Encryption**: Connection encryption and secure storage
- **Audit Logging**: User action tracking
- **Soft Deletes**: Data recovery capabilities
- **Access Logging**: Request/response logging

## 🚀 Deployment Architecture

### Development Environment
- **Docker Compose**: Multi-container development setup
- **Hot Reload**: Automatic code reloading during development
- **Database Seeding**: Test data for development
- **Debug Tools**: Logging, profiling, and debugging support

### Production Environment
- **Container Orchestration**: Docker Swarm or Kubernetes ready
- **Load Balancing**: Nginx reverse proxy
- **SSL Termination**: HTTPS encryption
- **Health Checks**: Application and database monitoring
- **Backup Strategy**: Automated database backups

### Kubernetes Readiness
The application is designed for Kubernetes deployment with:
- **Stateless Services**: Frontend and backend are stateless
- **ConfigMaps**: Environment-based configuration
- **Secrets**: Secure credential management
- **Persistent Volumes**: Database data persistence
- **Health Probes**: Liveness and readiness checks
- **Horizontal Scaling**: Multiple replicas support

## 📊 Monitoring & Observability

### Logging
- **Structured Logging**: JSON format for log aggregation
- **Log Levels**: Configurable verbosity (debug, info, warn, error)
- **Request Logging**: HTTP request/response tracking
- **Error Tracking**: Comprehensive error reporting

### Metrics
- **Health Endpoints**: Application health status
- **Performance Metrics**: Response times and throughput
- **Database Metrics**: Connection pool and query performance
- **Business Metrics**: User activity and system usage

### Alerting
- **System Alerts**: Service availability and performance
- **Business Alerts**: Critical workflow notifications
- **User Notifications**: In-app notification system

This architecture provides a scalable, maintainable, and secure foundation for the DocTheraLink healthcare management system, with clear separation of concerns and modern development practices.