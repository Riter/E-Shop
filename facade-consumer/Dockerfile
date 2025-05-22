FROM golang:1.24-alpine

# Установите git
RUN apk add --no-cache git

# Ваши другие шаги, например:
WORKDIR /app
COPY . .

# Загрузите зависимости Go
RUN go mod download

# Копируем все остальное
COPY . .

# Собираем приложение
RUN go build -o app cmd/main.go

# Используем порт по умолчанию
EXPOSE 8080

# Команда запуска
CMD ["./app"]