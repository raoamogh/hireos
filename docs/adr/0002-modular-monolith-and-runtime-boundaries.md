# ADR 0002: Modular Monolith and Runtime Boundaries

- Status: Accepted for initial implementation
- Date: 2026-11-17

## Context

HIREOS requires multiple business domains, including identity,
recruitment, workflow orchestration, assessments, communication,
interviews, offers and integrations.

The system must maintain clear domain boundaries without introducing
premature distributed-system complexity.

The PRD specifies a modular monolith architecture supported by
separate Temporal worker processes and an isolated AI service.

## Decisions

1. The primary backend application will follow a modular monolith
   architecture.
2. Business capabilities will be organized into domain-oriented
   internal modules.
3. Domain logic must not depend directly on infrastructure
   implementations.
4. External provider integrations must be isolated behind interfaces.
5. The HTTP API will run as a separate executable.
6. Temporal workflows and activities will execute in separate worker
   processes.
7. Database migrations will be managed through a dedicated migration
   executable.
8. AI functionality will remain isolated as specified by the PRD.
9. Microservice decomposition will not be introduced prematurely.

## Runtime Components

The Go backend will support these executables:

- hireos-api
- hireos-worker
- hireos-migrate

Responsibilities:

- API: handles HTTP requests and application interactions.
- Worker: executes Temporal workflows and activities.
- Migrate: applies database schema migrations.

The AI service remains an isolated architectural component.
Its implementation language and deployment topology are not decided
by this ADR.

## Alternatives Considered

### Fully distributed microservices

Not selected because the PRD explicitly directs HIREOS to avoid
premature microservice decomposition.

### Single undifferentiated application process

Not selected because the PRD requires separate Temporal worker
processes and an isolated AI service.

## Consequences

- Domain boundaries must remain clear within the Go codebase.
- Infrastructure dependencies must be isolated from domain logic.
- API and worker processes can be operated independently.
- Workflow execution does not depend on HTTP request lifetimes.
- Future service extraction remains possible without requiring
  premature distributed-system complexity.

## Deferred Decisions

- AI service implementation language.
- AI service deployment topology.
- Production process orchestration.
- Service-to-service communication details.
- Exact domain package dependency enforcement mechanisms.
