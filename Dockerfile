# Dockerfile for neighborhood-api

FROM golang:1.23-alpine AS builder

WORKDIR /app

# Instalar dependencias
RUN apk add --no-cache git ca-certificates

# Copiar módulos
COPY go.mod go.sum ./
RUN go mod download

# Copiar código
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o neighborhood-api ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/neighborhood-api .

EXPOSE 8080

CMD ["./neighborhood-api"]
