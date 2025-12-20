# ---------- builder ----------
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o api ./cmd/api

# ---------- api (distroless) ----------
FROM gcr.io/distroless/base-debian12 AS api

WORKDIR /app

COPY --from=builder /app/api /app/api

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT [ "/app/api" ]

# ---------- migrate ----------
FROM alpine:3.19 AS migrate

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/migrations /app/migrations

ENTRYPOINT ["migrate"]


# ---------- test ----------
FROM golang:1.25.5-alpine AS test

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENTRYPOINT ["go", "test", "./..."]
