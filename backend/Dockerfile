# Multi-stage build for the ORBIT backend (Go + PostgreSQL). Builds
# both the API server and the migration runner (cmd/migrate), so a
# single image can migrate the database then start serving.
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/orbit-backend .
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/orbit-backend /app/orbit-backend
COPY --from=build /out/migrate /app/migrate
COPY migrations /app/migrations
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
