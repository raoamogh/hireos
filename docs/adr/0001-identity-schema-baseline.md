
# ADR 0001: Identity and Organization Schema Baseline

- Status: Accepted for initial implementation
- Date: 2026-11-12

## Context

HIREOS requires a multi-tenant identity and organization foundation.
Users can belong to multiple organizations. Organization access must
be isolated, and authorization must support role-based permissions.

The PRD specifies the initial identity tables and requires UUID
primary keys, foreign keys, organization isolation and PostgreSQL RLS.

## Decisions

1. Users are global identities with globally unique normalized email
   addresses.
2. Organization membership is represented by organization_members.
3. Roles and permissions are global catalogs.
4. Role assignments are scoped to organization memberships.
5. Candidate accounts reference users independently of organization
   membership.
6. Invitation tokens are stored as hashes.
7. Tenant-scoped tables use organization_id and PostgreSQL RLS.
8. Candidate accounts use user-scoped PostgreSQL RLS.
9. Password hashes are stored; plaintext passwords are never stored.
10. UUID primary keys and timestamptz timestamps are used.

## Initial role catalog

- organization_owner
- administrator
- recruiter
- hiring_manager
- interviewer
- approver
- candidate

Permission definitions are intentionally not seeded until the
authorization matrix is designed.

## RLS context

Tenant-scoped application transactions must set:

app.current_organization_id

Candidate-account transactions must set:

app.current_user_id

These settings must be transaction-local, not persistent session
settings.

The production application database role must not be a superuser
or have BYPASSRLS.

## Deferred decisions

- Session storage and refresh-token lifecycle.
- Password-reset token storage.
- Email-verification token storage.
- Detailed permission catalogue and role-permission assignments.
- Candidate-specific authorization semantics.
- Organization invitation role assignment.
- Runtime database roles and grants.

These will be addressed in subsequent implementation decisions
and migrations.

## Consequences

The initial schema establishes identity, organization membership,
RBAC catalogs, candidate accounts and invitations.

It does not implement the complete authentication or authorization
system. Those remain separate implementation tasks.
