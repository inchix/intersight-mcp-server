FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /intersight-mcp-server ./cmd/intersight-mcp-server

FROM alpine:3.21
RUN apk --no-cache add ca-certificates && \
    adduser -D -h /app appuser
USER appuser
WORKDIR /app
COPY --from=builder /intersight-mcp-server .
ENTRYPOINT ["./intersight-mcp-server"]
