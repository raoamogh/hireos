
-- Roll back HIREOS identity and organization foundation.
-- Migration: 000001

DROP TABLE member_roles;
DROP TABLE role_permissions;
DROP TABLE permissions;
DROP TABLE roles;
DROP TABLE invitations;
DROP TABLE candidate_accounts;
DROP TABLE organization_members;
DROP TABLE users;
DROP TABLE organizations;
