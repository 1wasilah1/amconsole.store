-- Connect to PostgreSQL as postgres user
-- Run: sudo -u postgres psql

-- 1. Create database
CREATE DATABASE amconsole;

-- 2. Create user with password
CREATE USER amconsole WITH PASSWORD 'amconsole123';

-- 3. Grant privileges to user
GRANT ALL PRIVILEGES ON DATABASE amconsole TO amconsole;

-- 4. Connect to the new database
\c amconsole;

-- 5. Grant schema privileges
GRANT ALL ON SCHEMA public TO amconsole;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO amconsole;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO amconsole;

-- 6. Create tables
CREATE TABLE IF NOT EXISTS tvs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'ready',
    start_time TIMESTAMP,
    duration INTEGER DEFAULT 0,
    end_time TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admins (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL
);

-- 7. Grant table privileges
GRANT ALL PRIVILEGES ON TABLE tvs TO amconsole;
GRANT ALL PRIVILEGES ON TABLE admins TO amconsole;
GRANT USAGE, SELECT ON SEQUENCE tvs_id_seq TO amconsole;
GRANT USAGE, SELECT ON SEQUENCE admins_id_seq TO amconsole;

-- 8. Insert default admin (optional)
INSERT INTO admins (username, password) VALUES ('admin', 'admin123') ON CONFLICT DO NOTHING;