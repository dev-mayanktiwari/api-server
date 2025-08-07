-- Create databases for microservices
CREATE DATABASE auth_db;
CREATE DATABASE user_db;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE auth_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE user_db TO postgres;