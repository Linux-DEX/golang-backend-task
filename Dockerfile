FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy && go mod download && go mod verify && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-server cmd/api/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser
COPY --from=builder /app/api-server .
RUN chown -R appuser:appuser /home/appuser

USER appuser

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./api-server"]
