#!/bin/sh
set -e

# Postgres initialization script
# Superuser $POSTGRES_USER (postgres) is already created.
# Main database $POSTGRES_DB (jam_build) is also created.

echo "Running 001-ddl-init.sh for databases and users..."

# Create Authorizer database and Roles
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE $AUTHZ_DATABASE;
    
    -- Create Application User
    CREATE USER "$DB_APP_USER" WITH PASSWORD '$DB_APP_PASSWORD';
    ALTER DATABASE "$POSTGRES_DB" OWNER TO "$DB_APP_USER";
    ALTER DATABASE "$AUTHZ_DATABASE" OWNER TO "$DB_APP_USER";
    
    -- Create Client User
    CREATE USER "$DB_USER" WITH PASSWORD '$DB_PASSWORD';
    GRANT CONNECT ON DATABASE "$POSTGRES_DB" TO "$DB_USER";

    -- Grant schema permissions in jam_build
    GRANT ALL ON SCHEMA public TO "$DB_APP_USER";
    GRANT ALL ON SCHEMA public TO "$DB_USER";

    -- Set default privileges for tables created by jbadmin
    ALTER DEFAULT PRIVILEGES FOR ROLE "$DB_APP_USER" IN SCHEMA public GRANT ALL ON TABLES TO "$DB_USER";
    ALTER DEFAULT PRIVILEGES FOR ROLE "$DB_APP_USER" IN SCHEMA public GRANT ALL ON SEQUENCES TO "$DB_USER";
    
    -- Also for tables created by postgres (just in case)
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO "$DB_APP_USER";
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO "$DB_APP_USER";
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO "$DB_USER";
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO "$DB_USER";

    -- Create stub table for foreign key reference in application database
    CREATE TABLE IF NOT EXISTS authorizer_users (id CHAR(36) NOT NULL PRIMARY KEY);
    ALTER TABLE authorizer_users OWNER TO "$DB_APP_USER";
EOSQL

# Create users table in authorizer database and grant permissions
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$AUTHZ_DATABASE" <<-EOSQL
    GRANT ALL ON SCHEMA public TO "$DB_APP_USER";
    GRANT ALL ON SCHEMA public TO "$DB_USER";
    
    -- Set default privileges for authorizer DB
    ALTER DEFAULT PRIVILEGES FOR ROLE "$DB_APP_USER" IN SCHEMA public GRANT ALL ON TABLES TO "$DB_USER";
    ALTER DEFAULT PRIVILEGES FOR ROLE "$DB_APP_USER" IN SCHEMA public GRANT ALL ON SEQUENCES TO "$DB_USER";
    
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO "$DB_APP_USER";
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO "$DB_APP_USER";

    CREATE TABLE IF NOT EXISTS authorizer_users (id CHAR(36) NOT NULL PRIMARY KEY);
    ALTER TABLE authorizer_users OWNER TO "$DB_APP_USER";
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "$DB_APP_USER";
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "$DB_APP_USER";
EOSQL

echo "001-ddl-init.sh complete."
