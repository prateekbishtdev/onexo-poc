# onexo-poc

Onexo proof of concept.

## Hypothesis

_What we believe is true and want to validate._

## Success criteria

_Measurable outcomes, agreed before building, that decide go / no-go._

- [ ] Criterion 1
- [ ] Criterion 2

## Scope

**In scope**

- _TBD_

**Out of scope**

- _TBD_

## Getting started

Go 1.25 service backed by PostgreSQL 17 and Redis 7.

### Run everything in Docker

```sh
cp .env.example .env
make up          # builds the app and starts app + postgres + redis
curl localhost:8080/health
```

### Run the app on the host

```sh
cp .env.example .env
make deps        # starts only postgres + redis
make run
```

### Endpoints

| Method | Path           | Purpose                                                        |
| ------ | -------------- | -------------------------------------------------------------- |
| GET    | `/health`      | Readiness: pings Postgres and Redis; `503` if either is down   |
| GET    | `/health/live` | Liveness: process is up; never touches dependencies            |

### API tests (Bruno)

Open the `bruno/` folder as a collection in [Bruno](https://www.usebruno.com/), select the `local` environment and run the `health` folder. Or from the CLI:

```sh
cd bruno && npx @usebruno/cli run --env local
```

### Project layout

```
cmd/server/          entrypoint, wiring, graceful shutdown
internal/config/     env-based configuration
internal/database/   Postgres connection pool (pgx)
internal/cache/      Redis client (go-redis)
internal/handler/    HTTP handlers
internal/server/     router and middleware
bruno/               Bruno API collection
```

## Findings

_Results against each success criterion._

## Recommendation

_Go / no-go, with the estimated cost to productionize._
