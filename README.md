# User Metadata Service

Go service that consumes events from RabbitMQ, evaluates metadata rules, and applies atomic JSONB updates to `users.meta_data` with idempotency and failure recording.

## Architecture (Union)

- **Domain** (`internal/domain`): Business models — `Event`, `User`, `MetadataRule`, `MetadataRuleAction`, `ProcessedEvent`, `FailedEvent`, `EventSource`.
- **Application** (`internal/application`): Orchestrator, rule engine (condition + value resolution), metadata executor (set/increment/append/remove/merge/max/min), rule cache.
- **Infrastructure** (`internal/infrastructure`): GORM entities and repositories, RabbitMQ consumer, logger, DB and config.

## Event flow

1. Consumer receives a message from a queue bound to a topic (from `event_sources`).
2. **Idempotency**: Check `processed_events` by `event_id`; skip if already processed.
3. Load matching rules from cache/DB (`metadata_rules` + `metadata_rule_actions`) by `event_type` (and optional `event_version`).
4. For each rule action: evaluate `condition_expression`, resolve `value_template` (static/event/metadata), collect key-level operations.
5. Compute new `meta_data` from current JSONB and operations.
6. **Atomic transaction**: update `users.meta_data` and insert into `processed_events`.
7. On any processing error: insert into `failed_events` and Nack (or record and Ack, depending on policy).

## Database (MySQL)

Tables (GORM AutoMigrate):

- `event_sources` — topic name, enabled.
- `metadata_rules` — event_type, event_version, enabled, priority, description.
- `metadata_rule_actions` — rule_id, operation, metadata_key, value_source, value_template, condition_expression, execution_order.
- `processed_events` — event_id (PK), event_json (JSON), processed_at.
- `failed_events` — id, event_type, payload (JSON), error_message, created_at, processed_at.
- `users` — id (CHAR(36)), meta_data (JSON), updated_at.

Expected event JSON shape: `id`, `type`, `version`, `user_id`, `data` (optional). `user_id` is required for metadata updates.

## Configuration (env)

**Database** — use `DATABASE_URL` (full MySQL DSN), or set:

| Variable | Default |
|----------|---------|
| `DB_USER` | `user` |
| `DB_PASSWORD` | `password` |
| `DB_HOST` | `localhost` |
| `DB_PORT` | `3306` |
| `DB_NAME` | `users_metadata` |
| `DB_CHARSET` | `utf8mb4` |
| `DB_MAX_OPEN_CONNS` | 25 |
| `DB_MAX_IDLE_CONNS` | 5 |
| `DB_CONN_MAX_LIFETIME` | 5m |

**RabbitMQ** — use `RABBITMQ_URL` (full AMQP URL), or set:

| Variable | Default |
|----------|---------|
| `RABBITMQ_USER` | `guest` |
| `RABBITMQ_PASSWORD` | `guest` |
| `RABBITMQ_HOST` | `localhost` |
| `RABBITMQ_PORT` | `5672` |
| `RABBITMQ_VHOST` | `/` |
| `RABBITMQ_PREFETCH` | 10 |

**Super Stream** (optional, stream protocol on port 5552):

| Variable | Default |
|----------|---------|
| `RABBITMQ_STREAM_ENABLED` | `false` |
| `RABBITMQ_STREAM_PORT` | `5552` (host = `RABBITMQ_HOST`) |

When `RABBITMQ_STREAM_ENABLED=true`, the service creates one super stream consumer per enabled row in **event_sources** (stream name = `topic_name`). Add a row with e.g. `topic_name = 'users-metadata'` to consume from that super stream.

**Other**

| Variable | Default |
|----------|---------|
| `WORKERS` | 5 |
| `RULE_CACHE_TTL` | 1m |
| `SHUTDOWN_TIMEOUT` | 30s |

## Run

**Option A — Local (Docker)**  
Start MySQL and RabbitMQ, seed an event source, then run the service:

```bash
make run
```

Or manually: `docker compose up -d`, wait ~15s, then `go run ./cmd`. Add an event source: `INSERT INTO event_sources (id, topic_name, enabled, created_at) VALUES (UUID(), 'user.events', 1, NOW());`

**Option B — Direct**

```bash
go run ./cmd
```

Ensure MySQL and RabbitMQ are up and that `event_sources` has at least one row (topic_name, enabled=true) so the service creates queues and consumes.

### Staging / Super Stream

For staging, set `RABBITMQ_URL` (or `RABBITMQ_HOST`, `RABBITMQ_PORT`, `RABBITMQ_USER`, `RABBITMQ_PASSWORD`). Example (replace password):

```bash
export RABBITMQ_URL="amqp://4sale-rabbitmq-admin:YOUR_PASSWORD@rabbitmq-stg.rabbitmq-system.svc.cluster.local:5672/%2f"
# Or from outside the cluster (e.g. NLB):
# export RABBITMQ_HOST=rabbitmq-stg-nlb-24d4c1440fa2a16a.elb.eu-west-1.amazonaws.com
# export RABBITMQ_PORT=5672
# export RABBITMQ_USER=4sale-rabbitmq-admin
# export RABBITMQ_PASSWORD=YOUR_PASSWORD
```

To test with a **super stream**: create the super stream in RabbitMQ, add a row to **event_sources** with `topic_name = 'users-metadata'` (or your stream name), then run:

```bash
export RABBITMQ_STREAM_ENABLED=true
export RABBITMQ_STREAM_PORT=5552
go run ./cmd
```

The service will consume from each super stream named in `event_sources` (same handler as AMQP).

## Operations

- **Metadata operations**: `set`, `increment`, `append`, `remove`, `merge`, `max`, `min`.
- **Value sources**: `static`, `event` (e.g. `value_template`: `data.amount`), `metadata`.
- **Condition expression**: simple comparisons, e.g. `event.data.amount > 100`, `event.data.status == "active"`.
- Rule cache is refreshed every 1 minute (cron); rules are loaded dynamically from DB without redeploy.
