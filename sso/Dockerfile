# Этап сборки
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Статическая сборка (важно для Alpine)
ENV CGO_ENABLED=0 GOOS=linux

RUN go build -o sso ./cmd/sso
RUN go build -o migrate ./cmd/migrator

# Этап запуска
FROM alpine:3.18

WORKDIR /app

# Установка CA-сертификатов (опционально)
RUN apk --no-cache add ca-certificates

# Копируем бинарники и конфиги
COPY --from=builder /app/sso .
COPY --from=builder /app/migrate .
COPY migrations ./migrations
COPY environment environment
COPY config config

# Устанавливаем переменную окружения (опционально)
ENV CONFIG_PATH=./config/local.yaml

# Значение по умолчанию — запуск основного приложения
CMD ["./sso", "--config=./config/local.yaml"]
