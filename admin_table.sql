-- Tabel untuk admin login
CREATE TABLE IF NOT EXISTS admins (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default admin
INSERT INTO admins (username, password) VALUES ('admin', 'admin123') ON CONFLICT (username) DO NOTHING;

-- Grant privileges
GRANT ALL PRIVILEGES ON TABLE admins TO amconsole;
GRANT USAGE, SELECT ON SEQUENCE admins_id_seq TO amconsole;