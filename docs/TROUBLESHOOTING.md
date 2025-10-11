# Troubleshooting Guide

## Authentication Issues

### Login Returns "Invalid Credentials"

**Symptoms:**
- Users exist in database but login fails
- Backend logs show "Password verification failed"

**Solution:**
The bcrypt password hash in the database may be incorrect. All test users should use password `1234@1`.

To verify and fix:
```bash
# Check if backend is responding
curl -X POST http://localhost:4000/api/v1/auth/login/ \
  -H "Content-Type: application/json" \
  -d '{"username":"doctor","password":"1234@1"}'

# If it fails, the password hashes need to be regenerated
# This is already fixed in the seed data, so rebuild:
docker-compose down
docker-compose up --build -d
```

### Double `/api/api` in Request URLs

**Symptoms:**
- Network tab shows requests to `/api/api/v1/...`
- 404 errors on all API calls

**Solution:**
Check nginx configuration and environment variables:

1. Verify `fe/nginx.conf` has:
```nginx
location /api/ {
    proxy_pass http://backend:9000/api/;
}
```

2. Verify `docker-compose.yml` has:
```yaml
environment:
  REACT_CARE_API_URL: http://localhost:4000
```

3. Rebuild frontend:
```bash
docker-compose build frontend
docker-compose up -d
```

## Build Issues

### Frontend Build Fails: "REACT_CARE_API_URL environment variable is required"

**Solution:**
The Dockerfile needs a placeholder value for build time. Check `fe/Dockerfile` has:
```dockerfile
ENV REACT_CARE_API_URL="http://placeholder"
```

Then rebuild:
```bash
docker-compose build frontend
docker-compose up -d
```

## Database Issues

### Database Connection Refused

**Solution:**
```bash
# Check if postgres is running
docker ps | grep postgres

# Check logs
docker logs care-postgres

# Restart services
docker-compose restart postgres backend
```

### Database Not Seeding

**Solution:**
```bash
# Remove volumes and rebuild
docker-compose down -v
docker-compose up --build -d
```

## Port Conflicts

### Port Already in Use

**Symptoms:**
- Error: "bind: address already in use"

**Solution:**
```bash
# Find what's using the port (e.g., 4000)
lsof -i :4000

# Kill the process or change port in docker-compose.yml
```

## Container Issues

### Container Keeps Restarting

**Solution:**
```bash
# Check logs
docker logs care-backend
docker logs klyniva-frontend

# Check health status
docker ps -a

# Restart specific service
docker-compose restart backend
```

## Development Issues

### Changes Not Reflecting in UI

See [Development Workflow](./DEVELOPMENT.md#hot-reload-and-rebuilding)

### CORS Errors

**Solution:**
Check `be/cmd/server/main.go` CORS configuration includes your frontend URL:
```go
AllowOrigins: []string{"http://localhost:4000"}
```

## Getting Help

If you encounter issues not covered here:

1. Check container logs:
```bash
docker-compose logs -f
```

2. Check specific service:
```bash
docker logs care-backend --tail 100
docker logs klyniva-frontend --tail 100
```

3. Verify all services are healthy:
```bash
docker ps
```

4. Check environment variables:
```bash
docker exec klyniva-frontend cat /usr/share/nginx/html/config.js
```
