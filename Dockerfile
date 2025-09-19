# --- builder ---
FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go mod download && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# --- runtime ---
FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /app/server /server
ENV APP_PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
