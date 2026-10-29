# Stage 1: Build do binário estático em Go
FROM golang:alpine AS builder

WORKDIR /app

# Baixa as dependências primeiro para aproveitar o cache de camadas do Docker
COPY go.mod go.sum ./
RUN go mod download

# Copia o código-fonte e compila o binário otimizado
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/gopay ./cmd/api

# Stage 2: Imagem final enxuta e segura
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/gopay .

EXPOSE 8080

CMD ["./gopay"]
