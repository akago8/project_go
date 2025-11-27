FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o wallet ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/wallet /app/wallet
COPY config.env /app/config.env
COPY internal/db/migrations /app/migrations
ENV CONFIG_FILE=/app/config.env
EXPOSE 8080
ENTRYPOINT ["/app/wallet"]

