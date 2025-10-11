-- Initialize CARE database with required extensions and basic setup

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable PostGIS extension (if needed for location data)
-- CREATE EXTENSION IF NOT EXISTS postgis;

-- Create indexes for better performance
-- These will be created automatically by GORM, but can be added here for reference

-- Example: Create index on users table
-- CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
-- CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
-- CREATE INDEX IF NOT EXISTS idx_users_user_type ON users(user_type);

-- Example: Create index on facilities table
-- CREATE INDEX IF NOT EXISTS idx_facilities_name ON facilities(name);
-- CREATE INDEX IF NOT EXISTS idx_facilities_facility_type ON facilities(facility_type);

-- Example: Create index on patients table
-- CREATE INDEX IF NOT EXISTS idx_patients_name ON patients(name);
-- CREATE INDEX IF NOT EXISTS idx_patients_phone_number ON patients(phone_number);

-- Set timezone
SET timezone = 'UTC';

-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE care TO care;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO care;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO care;