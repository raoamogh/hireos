# ADR 0003: PostgreSQL Data Access and Migration Strategy

- Status: Accepted for initial implementation
- Date: 2026-11-17

## Context

HIREOS requires persistent relational data for identity, organization
membership, recruitment operations, workflow execution tracking,
audit history and integrations.

The database must support transactions, foreign keys, organization
isolation, Row-Level Security and reliable schema evolution.

The PRD specifies PostgreSQL, pgx, sqlc and golang-migrate.

## Decisions

1. PostgreSQL will be the primary relational database.
2. pgx will provide Go database connectivity.
3. sqlc will generate type-safe Go database access code from SQL.
4. golang-migrate will manage database schema migrations.
5. Migration files will be maintained in db/migrations.
6. SQL queries used for code generation will be maintained in
   db/queries.
7. Generated database code will be maintained separately from
   handwritten domain logic.
8. Database constraints will enforce important data invariants.
9. Tenant-owned queries must enforce organization scope.
10. PostgreSQL Row-Level Security will provide an additional layer
    of tenant isolation.

## Alternatives Considered

### An ORM-based data access layer

Not selected. The PRD explicitly specifies pgx and sqlc.

### Application-managed schema changes

Not selected. The PRD explicitly specifies golang-migrate.

### A non-relational primary database

Not selected because the PRD requires relational constraints,
transactions, foreign keys and PostgreSQL Row-Level Security.

## Consequences

- SQL remains explicit and reviewable.
- sqlc generates typed database access code.
- Schema changes are versioned and reversible through migration
  files.
- Database constraints enforce important invariants.
- Tenant isolation must be enforced by application query design
  and PostgreSQL Row-Level Security.
- Generated code must remain synchronized with SQL queries and
  schema definitions.

## Deferred Decisions

- Production database topology.
- Database replication and failover.
- Connection pooling limits for production.
- Production database role and grant design.
- Backup and recovery configuration.
- Database partitioning strategy.
