# --- Build Stage ---
FROM golang:1.23.1 AS builder

ENV CGO_ENABLED=0
WORKDIR /app
COPY . .

RUN go build -o flox main.go

# --- Runtime Stage ---
FROM gcr.io/distroless/static

COPY --from=builder /app/flox /flox
COPY pipeline.yaml /etc/flox/pipeline.yaml

ENTRYPOINT ["/flox"]
CMD ["--config", "/etc/flox/pipeline.yaml"]
