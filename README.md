# HIREOS

**HIREOS** is a recruitment operating system designed to manage and orchestrate the candidate hiring lifecycle.

The project uses a modular monolith architecture with Go as its primary backend language, PostgreSQL for persistent data, and Temporal for durable workflow orchestration.

## Technology Stack

| Component | Technology |
|---|---|
| Backend | Go |
| Database | PostgreSQL |
| Database driver | pgx |
| SQL generation | sqlc |
| Database migrations | golang-migrate |
| Cache | Redis |
| Object storage | MinIO |
| Workflow orchestration | Temporal |
| Containers | Docker Compose |
| CI | GitHub Actions |

## Prerequisites

Install the following tools before setting up the project:

- Git
- Go (version specified in `go.mod`)
- Docker Desktop
- Docker Compose
- Temporal CLI
- sqlc
- golang-migrate CLI

Docker Desktop must be running before starting the infrastructure.

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/raoamogh/hireos.git
cd hireos
```

### 2. Configure environment variables

Create your local environment file:

```bash
cp .env.example .env
```

Edit `.env` and replace the placeholder passwords with secure local development passwords.

**Important:** Never commit `.env` or share its credentials. The file is excluded from Git.

### 3. Start infrastructure

Start PostgreSQL, Redis and MinIO:

```bash
make infra-up
```

Check the services:

```bash
docker compose ps
```

Wait until PostgreSQL and Redis report healthy status.

### 4. Run database migrations

Apply pending migrations:

```bash
make migrate-up
```

Check the current migration version:

```bash
make migrate-version
```

### 5. Start Temporal

Temporal runs separately from the Docker Compose services.

Open a dedicated terminal:

```bash
cd "$HOME/Developer/hireos"

mkdir -p data/temporal

temporal server start-dev \
  --db-filename "$PWD/data/temporal/temporal.db"
```

The Temporal UI is available at:

http://localhost:8233

Keep this terminal running while developing features that depend on Temporal.

### 6. Start the Go API

Open another terminal:

```bash
cd "$HOME/Developer/hireos"

set -a
source .env
set +a

go run ./cmd/api
```

The API listens on port `8080`.

Keep the terminal running while using the API.

## Health Checks

The API exposes two initial health endpoints.

| Endpoint | Purpose |
|---|---|
| `GET /health` | Checks whether the API is running |
| `GET /ready` | Checks whether the API and PostgreSQL are ready |

Test the API:

```bash
curl -i http://127.0.0.1:8080/health
```

Test database readiness:

```bash
curl -i http://127.0.0.1:8080/ready
```

A successful readiness response indicates that PostgreSQL is reachable.

## Development Commands

The project includes a Makefile for common development operations.

| Command | Description |
|---|---|
| `make help` | Display available commands |
| `make test` | Run Go tests |
| `make build` | Compile Go packages |
| `make fmt` | Format Go source files |
| `make vet` | Run Go static analysis |
| `make sqlc-generate` | Generate database code |
| `make migrate-up` | Apply pending migrations |
| `make migrate-down` | Roll back one migration |
| `make migrate-version` | Display migration version |
| `make infra-up` | Start Docker infrastructure |
| `make infra-down` | Stop Docker infrastructure without deleting volumes |

## Local Service Addresses

| Service | Address |
|---|---|
| Go API | http://localhost:8080 |
| PostgreSQL | localhost:5434 |
| Redis | localhost:6379 |
| MinIO API | http://localhost:9000 |
| MinIO Console | http://localhost:9001 |
| Temporal | localhost:7233 |
| Temporal UI | http://localhost:8233 |

## Project Structure

```text
hireos/
├── cmd/
│   ├── api/
│   ├── migrate/
│   └── worker/
├── internal/
│   ├── application/
│   ├── candidate/
│   ├── identity/
│   ├── job/
│   ├── organization/
│   └── platform/
├── db/
│   ├── migrations/
│   └── queries/
├── docs/
│   └── adr/
├── .github/
│   └── workflows/
├── docker-compose.yml
├── Makefile
├── sqlc.yaml
└── go.mod
```

## Development Workflow

The repository uses the following branch conventions:

- `main` — stable branch.
- `develop` — integration branch.
- `feature/*` — feature development.
- `fix/*` — bug fixes.
- `chore/*` — maintenance tasks.
- `docs/*` — documentation changes.

Changes should be developed on dedicated branches and submitted through pull requests targeting `develop`.

## Security Notes

- Never commit secrets, credentials, tokens or local environment files.
- Do not use real candidate information in local development or test fixtures.
- Treat external input as untrusted.
- Keep sensitive application data out of logs.

## Project Status

HIREOS is under active development.

Refer to the Product Requirements Document (`PRD.md`) for the approved product scope, architecture, functional requirements and engineering roadmap.
