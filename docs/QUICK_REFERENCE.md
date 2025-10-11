# Quick Reference Card

## 🚀 Start/Stop

```bash
# Start everything
docker-compose up -d

# Stop everything
docker-compose down

# Restart specific service
docker-compose restart backend
```

## 🔄 After Making Changes

| What Changed | Command |
|--------------|---------|
| Frontend code | `docker-compose build frontend && docker-compose up -d` |
| Backend code | `docker-compose build backend && docker-compose up -d` |
| Both | `docker-compose up --build -d` |
| Config only | `docker-compose up -d` |

## 📊 Monitoring

```bash
# View all logs
docker-compose logs -f

# View specific service
docker logs care-backend -f

# Check status
docker ps
```

## 🔑 Test Credentials

All users: password `1234@1`
- `doctor`
- `nurse`
- `physiotherapist`
- `stateadmin`

## 🌐 URLs

- Frontend: http://localhost:4000
- Backend: http://localhost:9000
- API Docs: http://localhost:9000/swagger/
- DB Admin: http://localhost:8080

## 🆘 Quick Fixes

```bash
# Not working? Try rebuild
docker-compose up --build -d

# Still not working? Fresh start
docker-compose down -v
docker-compose up --build -d

# Check what's wrong
docker-compose logs -f
```

## 📝 Database Access

```bash
# Web UI: http://localhost:8080
# Server: postgres
# User: care
# Pass: care
# DB: care

# CLI
docker exec -it care-postgres psql -U care -d care
```

## 🧪 Test API

```bash
# Login
curl -X POST http://localhost:4000/api/v1/auth/login/ \
  -H "Content-Type: application/json" \
  -d '{"username":"doctor","password":"1234@1"}'
```

## 📚 More Help

- [README.md](README.md) - Full documentation
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) - Development guide
- [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) - Common issues
