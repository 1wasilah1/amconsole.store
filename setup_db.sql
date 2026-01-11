-- Setup database dan user untuk PS Rental
-- Jalankan sebagai postgres user di VPS

-- 1. Buat database
CREATE DATABASE amconsole;

-- 2. Buat user
CREATE USER amconsole WITH PASSWORD 'amconsole123';

-- 3. Grant privileges
GRANT ALL PRIVILEGES ON DATABASE amconsole TO amconsole;

-- 4. Connect ke database amconsole
\c amconsole;

-- 5. Grant schema privileges
GRANT ALL ON SCHEMA public TO amconsole;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO amconsole;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO amconsole;

-- 6. Buat tabel tvs
CREATE TABLE IF NOT EXISTS tvs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'ready',
    start_time TIMESTAMP,
    duration INTEGER DEFAULT 0,
    end_time TIMESTAMP
);

-- 7. Grant privileges pada tabel
GRANT ALL PRIVILEGES ON TABLE tvs TO amconsole;
GRANT USAGE, SELECT ON SEQUENCE tvs_id_seq TO amconsole;