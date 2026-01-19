#!/bin/sh
set -e

# Use mariadb or mysql
M_CMD=$(command -v mariadb || command -v mysql)

echo "Running 001-ddl-init.sh for databases and users..."

$M_CMD -u root --password="${MYSQL_ROOT_PASSWORD}" <<EOF
CREATE DATABASE IF NOT EXISTS \`${AUTHZ_DATABASE}\`;
CREATE DATABASE IF NOT EXISTS \`${MYSQL_DATABASE}\`;
CREATE USER IF NOT EXISTS '${DB_USER}'@'%' IDENTIFIED BY '${DB_PASSWORD}';
CREATE USER IF NOT EXISTS '${MYSQL_USER}'@'%' IDENTIFIED BY '${MYSQL_PASSWORD}';
CREATE TABLE IF NOT EXISTS \`${AUTHZ_DATABASE}\`.authorizer_users (id CHAR(36) NOT NULL PRIMARY KEY);
GRANT SELECT ON \`${AUTHZ_DATABASE}\`.* TO '${MYSQL_USER}'@'%';
GRANT SELECT ON \`${AUTHZ_DATABASE}\`.* TO '${DB_USER}'@'%';
FLUSH PRIVILEGES;
EOF

echo "001-ddl-init.sh complete."