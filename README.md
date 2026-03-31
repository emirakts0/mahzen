# Mahzen

A personal knowledge management platform with semantic search, built with Go, Gin, and React.

## Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26, Gin, HTTP/3 (QUIC via quic-go) |
| Frontend | React 19, TypeScript, Vite 7, Tailwind CSS v4, React Router v7, TanStack Query v5 |
| Database | PostgreSQL 18 |
| Search | Meilisearch v1.41 (keyword + semantic/vector search) |
| AI | OpenAI `text-embedding-3-small` + `gpt-4o-mini` |

---

## Development

### 1. Start infrastructure

```bash
make docker-up
```

Starts PostgreSQL and Meilisearch. Wait for containers to be healthy.

### 2. Apply database migrations

```bash
make migrate-up
```

Runs the SQL directly inside the Postgres container via `docker exec` — no external migration tool required.

### 3. Configure

Edit `config.yaml` as needed. Defaults match the Docker Compose credentials so no changes are required for local development.

To enable AI features (embeddings + summarisation), set your OpenAI key:

```yaml
openai:
  api_key: "sk-..."
```

Leave it empty to run without AI — the app falls back to keyword-only search.

### 4. Install frontend dependencies

```bash
make web-install
```

### 5. Generate TLS certificates (required for HTTP/3)

HTTP/3 (QUIC) mandates TLS. Generate a self-signed certificate for local development:

```bash
make gen-certs
```

This creates `certs/server.crt` and `certs/server.key` (already set in `config.yaml`).
Browsers will show a security warning for self-signed certs — dismiss it once, or add `certs/server.crt` to your system trust store to suppress it permanently.

> To disable HTTP/3 and run plain HTTP instead, clear the `cert_file` and `key_file` values in `config.yaml`.

### 6. Start the backend

```bash
make run
```

REST API: `https://localhost:8080` (HTTP/2 + HTTP/3 when TLS is configured, plain HTTP otherwise)

### 7. Start the frontend dev server

In a second terminal:

```bash
make web-dev
```

Open `http://localhost:3000`.

---

## Production build

`make dist` produces a single self-contained binary with the React SPA embedded.

```bash
make dist
./mahzen -config config.yaml
```

`http://localhost:8080` serves both the REST API (`/v1/*`) and the React SPA (everything else).

---

## Infrastructure ports

| Service | Port |
|---|---|
| PostgreSQL | 5432 |
| Meilisearch | 7700 |
| Backend REST | 8080 |
| Frontend dev | 3000 |
