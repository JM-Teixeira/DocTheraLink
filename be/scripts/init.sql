-- Runs once when the PostgreSQL volume is first initialised (docker-compose).
-- Tables are created by the backend's migrations; the database owner is the
-- configured DB_USER, so no extra grants are needed.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
