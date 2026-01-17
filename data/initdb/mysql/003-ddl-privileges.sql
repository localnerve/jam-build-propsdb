--
-- Jam-build database privilege grants.
--
-- Jam-build, a web application practical reference.
-- Copyright (c) 2025 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC
-- 
-- This file is part of Jam-build.
-- Jam-build is free software: you can redistribute it and/or modify it
-- under the terms of the GNU Affero General Public License as published by the Free Software
-- Foundation, either version 3 of the License, or (at your option) any later version.
-- Jam-build is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
-- without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
-- See the GNU Affero General Public License for more details.
-- You should have received a copy of the GNU Affero General Public License along with Jam-build.
-- If not, see <https://www.gnu.org/licenses/>.
-- Additional terms under GNU AGPL version 3 section 7:
-- a) The reasonable legal notice of original copyright and author attribution must be preserved
--    by including the string: "Copyright (c) 2025 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC"
--    in this material, copies, or source code of derived works.
--

-- Grant SELECT, INSERT, UPDATE, DELETE permissions on the application_documents and user_documents tables to jbadmin
-- Grant SELECT permissions on application_documents to jbuser
-- Grant SELECT, INSERT, UPDATE, DELETE on user_documents to jbuser
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.application_documents TO 'jbadmin'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.user_documents TO 'jbadmin'@'%';
GRANT SELECT ON jam_build.application_documents TO 'jbuser'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.user_documents TO 'jbuser'@'%';

-- Grant SELECT, INSERT, UPDATE, DELETE permissions on the application_collections and user_collections tables to jbadmin
-- Grant SELECT permissions on application_collections to jbuser
-- Grant SELECT, INSERT, UPDATE, DELETE permissions on user_collections to jbuser
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.application_collections TO 'jbadmin'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.user_collections TO 'jbadmin'@'%';
GRANT SELECT ON jam_build.application_collections TO 'jbuser'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.user_collections TO 'jbuser'@'%';

-- Grant SELECT, INSERT, UPDATE, DELETE permissions on the application_properties and user_properties tables to jbadmin
-- Grant SELECT permissions on application_properties to jbuser
-- Grant SELECT, INSERT, UPDATE, DELETE permissions on user_properties to jbuser
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.application_properties TO 'jbadmin'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.user_properties TO 'jbadmin'@'%';
GRANT SELECT ON jam_build.application_properties TO 'jbuser'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.user_properties TO 'jbuser'@'%';

-- Grant SELECT, INSERT, UPDATE, DELETE permissions on junction tables to jbadmin
-- Grant SELECT permissions on application junction tables to jbuser
-- Grant SELECT, INSERT, UPDATE, DELETE permissions on user junction tables to jbuser
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.application_documents_collections TO 'jbadmin'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.application_collections_properties TO 'jbadmin'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.user_documents_collections TO 'jbadmin'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.user_collections_properties TO 'jbadmin'@'%';
GRANT SELECT ON jam_build.application_documents_collections TO 'jbuser'@'%';
GRANT SELECT ON jam_build.application_collections_properties TO 'jbuser'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.user_documents_collections TO 'jbuser'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON jam_build.user_collections_properties TO 'jbuser'@'%';

-- Grant SELECT permissions on authorizer_users table to both jbadmin and jbuser
-- This is required for foreign key validation on user_documents
GRANT SELECT ON authorizer.authorizer_users TO 'jbadmin'@'%';
GRANT SELECT ON authorizer.authorizer_users TO 'jbuser'@'%';

-- Apply the changes immediately
FLUSH PRIVILEGES;
