# CARE System Deployment Guide

This guide covers the complete deployment of the CARE healthcare management system, including both the React frontend and Go backend.

## System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │    Backend      │    │   Database      │
│   (React)       │◄──►│     (Go)        │◄──►│  (PostgreSQL)   │
│   Port: 4000    │    │   Port: 9000    │    │   Port: 5432    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Prerequisites

### Development Environment
- **Node.js**: 22.11.0 or higher
- **npm**: 10.9.0 or higher
- **Go**: 1.21 or higher
- **PostgreSQL**: 13 or higher
- **Docker**: 20.10 or higher (optional)
- **Docker Compose**: 2.0 or higher (optional)

### Production Environment
- **Linux Server**: Ubuntu 20.04+ or CentOS 8+
- **Reverse Proxy**: Nginx or Apache
- **SSL Certificate**: Let's Encrypt or commercial
- **Process Manager**: PM2 or systemd
- **Monitoring**: Optional but recommended

## Quick Start (Development)

### 1. Clone the Repository
```bash
git clone <repository-url>
cd DocTheraLink
```

### 2. Backend Setup
```bash
cd be

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Edit configuration
nano .env

# Start database with Docker
docker-compose up -d postgres redis

# Run migrations and seed data
make migrate-seed

# Start development server
make dev
```

### 3. Frontend Setup
```bash
cd fe

# Install dependencies
npm run install-all

# Setup plugins
npm run setup

# Copy environment file
cp .env.example .env

# Edit configuration (set REACT_CARE_API_URL=http://localhost:9000)
nano .env

# Start development server
npm run dev
```

### 4. Access the Application
- **Frontend**: http://localhost:4000
- **Backend API**: http://localhost:9000
- **API Documentation**: http://localhost:9000/swagger/
- **Database Admin**: http://localhost:8080 (Adminer)

## Docker Deployment

### Development with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Production Docker Setup

1. **Create production docker-compose file**:
```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: care
      POSTGRES_USER: care
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    restart: unless-stopped

  backend:
    build:
      context: ./be
      dockerfile: Dockerfile
    environment:
      GIN_MODE: release
      DB_PASSWORD: ${DB_PASSWORD}
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  frontend:
    build:
      context: ./fe
      dockerfile: Dockerfile
    environment:
      REACT_CARE_API_URL: https://api.yourdomain.com
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - backend
      - frontend
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

2. **Deploy**:
```bash
# Set environment variables
export DB_PASSWORD=your-secure-password
export JWT_SECRET=your-jwt-secret

# Deploy
docker-compose -f docker-compose.prod.yml up -d
```

## Manual Production Deployment

### 1. Server Setup (Ubuntu 20.04+)

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y nginx postgresql postgresql-contrib redis-server nodejs npm golang-go git

# Create application user
sudo useradd -m -s /bin/bash care
sudo usermod -aG sudo care
```

### 2. Database Setup

```bash
# Switch to postgres user
sudo -u postgres psql

# Create database and user
CREATE DATABASE care;
CREATE USER care WITH PASSWORD 'your-secure-password';
GRANT ALL PRIVILEGES ON DATABASE care TO care;
\q

# Configure PostgreSQL
sudo nano /etc/postgresql/13/main/postgresql.conf
# Set: listen_addresses = 'localhost'

sudo nano /etc/postgresql/13/main/pg_hba.conf
# Add: local   care    care    md5

# Restart PostgreSQL
sudo systemctl restart postgresql
```

### 3. Backend Deployment

```bash
# Switch to care user
sudo su - care

# Clone repository
git clone <repository-url> /home/care/care-app
cd /home/care/care-app/be

# Build application
go build -o care-backend cmd/server/main.go

# Create environment file
cp .env.example .env
nano .env
# Configure production settings

# Run migrations
go run cmd/migrate/main.go -seed

# Create systemd service
sudo nano /etc/systemd/system/care-backend.service
```

**Systemd service file**:
```ini
[Unit]
Description=CARE Backend API
After=network.target postgresql.service

[Service]
Type=simple
User=care
WorkingDirectory=/home/care/care-app/be
ExecStart=/home/care/care-app/be/care-backend
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start service
sudo systemctl enable care-backend
sudo systemctl start care-backend
sudo systemctl status care-backend
```

### 4. Frontend Deployment

```bash
cd /home/care/care-app/fe

# Install dependencies
npm run install-all

# Configure environment
cp .env.example .env
nano .env
# Set REACT_CARE_API_URL=https://api.yourdomain.com

# Build for production
npm run build

# Copy build to web directory
sudo mkdir -p /var/www/care
sudo cp -r build/* /var/www/care/
sudo chown -R www-data:www-data /var/www/care
```

### 5. Nginx Configuration

```bash
# Create Nginx configuration
sudo nano /etc/nginx/sites-available/care
```

**Nginx configuration**:
```nginx
# Frontend
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    
    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;
    
    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
    
    # Frontend
    location / {
        root /var/www/care;
        index index.html;
        try_files $uri $uri/ /index.html;
        
        # Cache static assets
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }
}

# Backend API
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;
    
    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    
    # Proxy to backend
    location / {
        proxy_pass http://localhost:9000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/care /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 6. SSL Certificate (Let's Encrypt)

```bash
# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com -d api.yourdomain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

## Environment Configuration

### Backend Environment Variables (.env)

```env
# Server
PORT=9000
HOST=0.0.0.0
GIN_MODE=release

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=care
DB_PASSWORD=your-secure-password
DB_NAME=care
DB_SSLMODE=require
DB_TIMEZONE=UTC

# JWT
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# CORS
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com

# Security
BCRYPT_COST=12
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1h

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# External APIs
GOOGLE_MAPS_API_KEY=your-google-maps-api-key
GOV_DATA_API_KEY=your-gov-data-api-key

# Email (optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@yourdomain.com
```

### Frontend Environment Variables (.env)

```env
# Backend API URL
REACT_CARE_API_URL=https://api.yourdomain.com

# Optional customization
REACT_CUSTOM_DESCRIPTION=Your healthcare system description
REACT_HEADER_LOGO={"light": "https://cdn.yourdomain.com/logo.png", "dark": "https://cdn.yourdomain.com/logo-dark.png"}

# External APIs
REACT_GMAPS_API_KEY=your-google-maps-api-key
REACT_GOV_DATA_API_KEY=your-gov-data-api-key
REACT_RECAPTCHA_SITE_KEY=your-recaptcha-site-key

# Analytics (optional)
REACT_PLAUSIBLE_SITE_DOMAIN=yourdomain.com
REACT_PLAUSIBLE_SERVER_URL=https://plausible.yourdomain.com

# Error tracking (optional)
REACT_SENTRY_DSN=your-sentry-dsn
REACT_SENTRY_ENVIRONMENT=production
```

## Monitoring and Maintenance

### 1. Log Management

```bash
# Backend logs
sudo journalctl -u care-backend -f

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-13-main.log
```

### 2. Database Backup

```bash
# Create backup script
sudo nano /home/care/backup.sh
```

```bash
#!/bin/bash
BACKUP_DIR="/home/care/backups"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# Database backup
pg_dump -h localhost -U care care > $BACKUP_DIR/care_db_$DATE.sql

# Keep only last 7 days
find $BACKUP_DIR -name "care_db_*.sql" -mtime +7 -delete

# Upload to cloud storage (optional)
# aws s3 cp $BACKUP_DIR/care_db_$DATE.sql s3://your-backup-bucket/
```

```bash
# Make executable and schedule
chmod +x /home/care/backup.sh
crontab -e
# Add: 0 2 * * * /home/care/backup.sh
```

### 3. Health Monitoring

```bash
# Create health check script
sudo nano /home/care/health-check.sh
```

```bash
#!/bin/bash
# Check backend health
if ! curl -f http://localhost:9000/health > /dev/null 2>&1; then
    echo "Backend is down, restarting..."
    sudo systemctl restart care-backend
fi

# Check database connection
if ! pg_isready -h localhost -U care > /dev/null 2>&1; then
    echo "Database connection failed"
    # Send alert
fi
```

### 4. Updates and Maintenance

```bash
# Create update script
sudo nano /home/care/update.sh
```

```bash
#!/bin/bash
cd /home/care/care-app

# Pull latest changes
git pull origin main

# Update backend
cd be
go build -o care-backend cmd/server/main.go
sudo systemctl restart care-backend

# Update frontend
cd ../fe
npm run install-all
npm run build
sudo cp -r build/* /var/www/care/
sudo chown -R www-data:www-data /var/www/care

echo "Update completed"
```

## Security Checklist

- [ ] Use strong passwords for database and JWT secrets
- [ ] Enable SSL/TLS with valid certificates
- [ ] Configure firewall (UFW) to allow only necessary ports
- [ ] Regular security updates for the operating system
- [ ] Database access restricted to localhost
- [ ] Proper CORS configuration
- [ ] Rate limiting enabled
- [ ] Security headers configured in Nginx
- [ ] Regular backups with encryption
- [ ] Log monitoring and alerting
- [ ] User access controls and role-based permissions

## Troubleshooting

### Common Issues

1. **Backend won't start**
   - Check database connection
   - Verify environment variables
   - Check port availability
   - Review logs: `sudo journalctl -u care-backend`

2. **Frontend build fails**
   - Check Node.js version
   - Clear npm cache: `npm cache clean --force`
   - Delete node_modules and reinstall

3. **Database connection issues**
   - Verify PostgreSQL is running
   - Check connection parameters
   - Ensure user has proper permissions

4. **CORS errors**
   - Verify CORS_ALLOWED_ORIGINS includes frontend domain
   - Check protocol (http vs https)

### Performance Optimization

1. **Database**
   - Add indexes for frequently queried fields
   - Use connection pooling
   - Regular VACUUM and ANALYZE

2. **Backend**
   - Enable Gzip compression
   - Use Redis for caching
   - Optimize database queries

3. **Frontend**
   - Enable Nginx gzip compression
   - Use CDN for static assets
   - Implement service worker caching

This deployment guide provides a comprehensive setup for both development and production environments. Adjust configurations based on your specific requirements and infrastructure.