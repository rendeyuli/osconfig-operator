FROM golang:1.24.4 as builder

WORKDIR /app
COPY . .
RUN go build -o agent main.go configmap_watcher.go applier.go

FROM debian:bookworm-slim
COPY --from=builder /app/agent /usr/local/bin/agent

ENTRYPOINT ["/usr/local/bin/agent"]