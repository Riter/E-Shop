# Билд стадии
FROM golang:1.24 as builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
# RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"
RUN go env -w GOPROXY=direct
RUN go env -w GOSUMDB=off
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# Финальная стадия
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main /app/main

# Установка зависимостей для alpine (если нужны)
RUN apk --no-cache add ca-certificates

ENTRYPOINT ["/app/main"]