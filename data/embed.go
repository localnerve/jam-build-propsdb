package data

import (
	_ "embed"
)

//go:embed initdb/mariadb/002-ddl-tables.sql
var InitdbMariaDBTables string

//go:embed initdb/mariadb/003-ddl-privileges.sql
var InitdbMariaDBPrivileges string
