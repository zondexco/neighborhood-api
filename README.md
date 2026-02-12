# Neighborhood API

Backend service for the Neighborhood platform, written in Go with Gin.

## Requirements

- Go 1.22+
- PostgreSQL (Supabase-compatible)
- Make (optional)

## Configuration

Copy `.env.example` to `.env` and fill the variables:

```bash
cp .env.example .env
```

Key database settings:

- `DB_DISABLE_PREPARED_STATEMENTS=true` should stay enabled when connecting through PgBouncer/Supabase (transaction pooling). This forces the driver to use the simple protocol and avoids `pq: bind message ...` errors.
- If you host Postgres yourself with session pooling, you can set it to `false` for better throughput.

Logging settings:

- `LOG_ENABLE_FILE=true` enables local file logging with rotation.
- `LOG_FILE_PATH=logs/neighborhood-api.log` sets the output file path.
- `LOG_MAX_SIZE_MB`, `LOG_MAX_BACKUPS`, `LOG_MAX_AGE_DAYS`, `LOG_COMPRESS` control rotation policy.
- `LOG_ALSO_STDOUT=true` keeps logs in console in addition to file output.

## Running locally

```bash
make run
```

or

```bash
go run ./cmd/api
```

## Testing

```bash
make test
```

## Troubleshooting

- `pq: bind message ...` errors: ensure `DB_DISABLE_PREPARED_STATEMENTS` is `true` or switch your pool to session mode.
- CORS errors: update `CORS_ALLOWED_ORIGINS`.
