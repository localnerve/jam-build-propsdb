#!/bin/sh
set -e

echo "Running 003-ddl-privileges.sh for database $POSTGRES_DB and $AUTHZ_DATABASE..."

# Grant permissions to jbuser (the secondary/client user) in jam_build
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Ensure jbadmin owns everything created during init
    REASSIGN OWNED BY "$POSTGRES_USER" TO "$DB_APP_USER";
    
    -- Grant everything to jbuser for E2E tests
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "$DB_USER";
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "$DB_USER";
    GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO "$DB_USER";
EOSQL

# Grant permissions in authorizer database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$AUTHZ_DATABASE" <<-EOSQL
    REASSIGN OWNED BY "$POSTGRES_USER" TO "$DB_APP_USER";
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "$DB_USER";
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "$DB_USER";
    GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO "$DB_USER";
EOSQL

echo "003-ddl-privileges.sh complete."
