# Используем официальный golang-образ
FROM golang:1.24-alpine

# Устанавливаем зависимости
RUN apk add --no-cache git

# Создаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

ENV GOPROXY=direct
ENV GOSUMDB=off

# Скачиваем зависимости
RUN go mod download

# Копируем все остальное
COPY . .

# Собираем приложение
RUN go build -o app cmd/main.go

# Используем порт по умолчанию
EXPOSE 8080

# Команда запуска
CMD ["./app"]