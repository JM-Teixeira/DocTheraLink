# Deployment Instructions - Phase 3-5

## Quick Start

### For Windows Users

1. **Open Command Prompt or PowerShell** in the project root directory

2. **Run the deployment script:**
   ```cmd
   deploy_phase3_5.bat
   ```

3. **Wait for completion** (approximately 5-10 minutes)

4. **Verify deployment:**
   - Frontend: http://localhost:4000
   - Backend: http://localhost:9000
   - API Docs: http://localhost:9000/swagger/

### For Linux/Mac Users

1. **Open Terminal** in the project root directory

2. **Make script executable:**
   ```bash
   chmod +x deploy_phase3_5.sh
   ```

3. **Run the deployment script:**
   ```bash
   ./deploy_phase3_5.sh
   ```

4. **Wait for completion** (approximately 5-10 minutes)

5. **Verify deployment:**
   - Frontend: http://localhost:4000
   - Backend: http://localhost:9000
   - API Docs: http://localhost:9000/swagger/

## What Gets Deployed

### Phase 3: Facility Types & Features
- ✅ 5 new facility types (Clinic, Physiotherapy Clinic, etc.)
- ✅ 45+ new facility features
- ✅ Phone number default changed to +351 (Portugal)

### Phase 4: Clock In/Out System
- ✅ Staff can clock in/out
- ✅ Automatic duration tracking
- ✅ Time log history
- ✅ Admin oversight

### Phase 5: Timetable Management
- ✅ Create facility sections
- ✅ Assign staff to sections
- ✅ Visual timetable grid
- ✅ Capacity management

## Manual Deployment (If Scripts Fail)

### Step 1: Backup Database

```bash
docker exec care-postgres pg_dump -U care care > backup_before_phase3_5.sql
```

**Important:** Keep this backup file safe!

### Step 2: Run Database Migration

```bash
docker exec -i care-postgres psql -U care -d care < be/scripts/phase3_5_migration.sql
```

### Step 3: Update Backend Dependencies

```bash
cd be
go mod tidy
cd ..
```

### Step 4: Rebuild Containers

```bash
docker-compose build backend frontend
```

### Step 5: Restart Services

```bash
docker-compose up -d backend frontend
```

### Step 6: Verify Services

```bash
# Check backend health
curl http://localhost:9000/health

# Check frontend
curl http://localhost:4000
```

## Testing After Deployment

### Quick Test 1: Login

1. Go to http://localhost:4000
2. Login with: `doctor` / `1234@1`
3. Verify you can access the dashboard

### Quick Test 2: Clock In/Out

1. Look for the Clock In/Out button (usually in header or profile)
2. Click "Clock In"
3. Verify success message
4. Click "Clock Out"
5. Verify duration is calculated

### Quick Test 3: Facility Features

1. Go to Facilities → Create Facility
2. Select "Physiotherapy Clinic" as type
3. Open Features dropdown
4. Verify you see 45+ features
5. Select a few features and save

### Quick Test 4: Timetable

1. Go to a Facility page
2. Navigate to Timetable tab
3. Create a section (e.g., "PT Room 1")
4. Create a timetable slot
5. Verify it appears in the grid

## Troubleshooting

### Issue: Script Fails to Run

**Windows:**
```cmd
# Try running as administrator
# Right-click Command Prompt → Run as Administrator
deploy_phase3_5.bat
```

**Linux/Mac:**
```bash
# Ensure script is executable
chmod +x deploy_phase3_5.sh
./deploy_phase3_5.sh
```

### Issue: Docker Not Running

**Solution:**
1. Start Docker Desktop
2. Wait for Docker to fully start
3. Run the deployment script again

### Issue: Database Migration Fails

**Solution:**
1. Check if PostgreSQL container is running:
   ```bash
   docker ps | grep postgres
   ```
2. If not running, start it:
   ```bash
   docker-compose up -d postgres
   ```
3. Wait 10 seconds, then retry migration

### Issue: Backend Won't Start

**Solution:**
1. Check logs:
   ```bash
   docker logs care-backend
   ```
2. Look for error messages
3. Common fixes:
   - Restart: `docker-compose restart backend`
   - Rebuild: `docker-compose build backend && docker-compose up -d backend`

### Issue: Frontend Won't Start

**Solution:**
1. Check logs:
   ```bash
   docker logs klyniva-frontend
   ```
2. Common fixes:
   - Restart: `docker-compose restart frontend`
   - Rebuild: `docker-compose build frontend && docker-compose up -d frontend`

### Issue: "Port Already in Use"

**Solution:**
1. Check what's using the port:
   ```bash
   # Windows
   netstat -ano | findstr :4000
   netstat -ano | findstr :9000
   
   # Linux/Mac
   lsof -i :4000
   lsof -i :9000
   ```
2. Stop the conflicting service or change ports in docker-compose.yml

### Issue: Changes Not Showing

**Solution:**
1. Clear browser cache (Ctrl+Shift+Delete)
2. Hard refresh (Ctrl+F5)
3. Restart containers:
   ```bash
   docker-compose restart
   ```

## Rollback Instructions

If something goes wrong and you need to rollback:

### Step 1: Stop Services

```bash
docker-compose down
```

### Step 2: Restore Database

```bash
docker-compose up -d postgres
docker exec -i care-postgres psql -U care -d care < backup_before_phase3_5.sql
```

### Step 3: Revert Code

```bash
git checkout main
# or
git checkout <previous-commit-hash>
```

### Step 4: Rebuild and Restart

```bash
docker-compose up --build -d
```

## Post-Deployment Checklist

- [ ] All services are running (check with `docker ps`)
- [ ] Frontend is accessible at http://localhost:4000
- [ ] Backend is accessible at http://localhost:9000
- [ ] Can login with test credentials
- [ ] Clock In/Out button is visible
- [ ] New facility types appear in dropdown
- [ ] New features appear in multi-select
- [ ] Timetable tab is accessible
- [ ] No errors in browser console (F12)
- [ ] No errors in backend logs (`docker logs care-backend`)

## Next Steps After Deployment

### 1. Test Thoroughly

Follow the testing guide: [docs/TESTING_PHASE3_5.md](docs/TESTING_PHASE3_5.md)

### 2. Train Staff

- Show staff how to use Clock In/Out
- Demonstrate timetable features
- Explain new facility types and features

### 3. Monitor

```bash
# Watch backend logs
docker logs -f care-backend

# Watch frontend logs
docker logs -f klyniva-frontend

# Watch all logs
docker-compose logs -f
```

### 4. Gather Feedback

- Ask users about the new features
- Note any issues or suggestions
- Plan improvements

### 5. Backup Regularly

```bash
# Create daily backups
docker exec care-postgres pg_dump -U care care > backup_$(date +%Y%m%d).sql
```

## Documentation

After deployment, review these documents:

1. **[Implementation Guide](docs/PHASE3_5_IMPLEMENTATION.md)** - Complete feature documentation
2. **[Testing Guide](docs/TESTING_PHASE3_5.md)** - How to test everything
3. **[Quick Reference](docs/QUICK_REFERENCE_PHASE3_5.md)** - Quick command reference
4. **[Implementation Summary](IMPLEMENTATION_SUMMARY.md)** - What was implemented
5. **[Changelog](CHANGELOG_PHASE3_5.md)** - Detailed changes

## Support

### Getting Help

1. **Check Documentation**
   - Review the guides in the `docs/` folder
   - Check the troubleshooting section

2. **Check Logs**
   ```bash
   docker logs care-backend
   docker logs klyniva-frontend
   docker logs care-postgres
   ```

3. **Check API Documentation**
   - Go to http://localhost:9000/swagger/
   - Test endpoints directly

4. **Contact Development Team**
   - Provide error messages
   - Include steps to reproduce
   - Share relevant logs

## Important Notes

### Data Safety
- ✅ Backup is created automatically before migration
- ✅ Keep backups for at least 30 days
- ✅ Test rollback procedure in staging first

### Performance
- ✅ First load may be slower (building containers)
- ✅ Subsequent starts are faster
- ✅ Database indexes improve query performance

### Security
- ✅ All endpoints require authentication
- ✅ Change default passwords in production
- ✅ Use HTTPS in production
- ✅ Keep backups secure

### Maintenance
- ✅ Monitor disk space (logs and database)
- ✅ Regular backups (daily recommended)
- ✅ Update dependencies periodically
- ✅ Review logs for errors

## Production Deployment

For production deployment:

1. **Use production docker-compose file**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

2. **Set environment variables**
   - Update `.env` files
   - Set secure JWT_SECRET
   - Configure CORS properly
   - Set production database credentials

3. **Enable HTTPS**
   - Configure SSL certificates
   - Update nginx configuration
   - Redirect HTTP to HTTPS

4. **Monitor**
   - Set up monitoring (Prometheus, Grafana)
   - Configure alerts
   - Log aggregation

5. **Backup Strategy**
   - Automated daily backups
   - Off-site backup storage
   - Test restore procedures

## Success Criteria

Deployment is successful when:

- ✅ All services start without errors
- ✅ Frontend loads correctly
- ✅ Backend API responds
- ✅ Database migrations complete
- ✅ Can login and use the system
- ✅ New features are accessible
- ✅ No errors in logs
- ✅ All tests pass

## Congratulations!

If you've reached this point and all checks pass, you've successfully deployed Phase 3-5 features! 🎉

Your system now has:
- ✅ Enhanced facility management
- ✅ Staff time tracking
- ✅ Timetable scheduling

Enjoy the new features!

---

**Deployment Date:** _____________  
**Deployed By:** _____________  
**Version:** 1.0.0  
**Status:** ⬜ Success ⬜ Failed ⬜ Rolled Back
