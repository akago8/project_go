## Wallet Service

REST API for depositing and withdrawing funds from wallets backed by PostgreSQL with pessimistic locking and retry logic to survive heavy contention.

### Requirements

- Go 1.22
- Docker and docker-compose

### Configuration

Copy `.env.example` to `config.env` and adjust values if needed. The service always reads settings from `config.env`. For docker Compose the bundled file already points to the bundled database service.

### Running locally with Docker

```
docker-compose up --build
```

Once both containers are healthy the API listens on `http://localhost:8080`.

### Running locally without Docker

1. Start PostgreSQL with matching credentials.
2. Ensure `config.env` has the correct connection settings.
3. Run `go run ./cmd/server`.

### API

- `POST /api/v1/wallet` body:
  ```
  {
    "walletId": "UUID",
    "operationType": "DEPOSIT" | "WITHDRAW",
    "amount": 1000
  }
  ```
  Response contains the updated balance.

- `GET /api/v1/wallets/{walletId}` returns the current balance.

### Tests

```
go test ./...
```

### Decisions

- Balance stored as `BIGINT` with cents semantics to avoid floating point errors.
- Each balance change runs inside a serializable transaction with `SELECT ... FOR UPDATE` and retry loop for serialization or deadlock errors, allowing safe processing of 1000 RPS on a single wallet without 5xx responses.

