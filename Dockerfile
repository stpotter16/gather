FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o /app/server ./cmd/server && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o /app/migrate ./cmd/migrate

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrate .
EXPOSE 8080
CMD ["/app/server"]
