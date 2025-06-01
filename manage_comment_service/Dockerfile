
FROM golang:1.24-alpine AS builder

WORKDIR /app


COPY go.mod go.sum ./


RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o comments_service ./cmd/main.go


FROM alpine:latest

WORKDIR /app


RUN apk add --no-cache curl


COPY --from=builder /app/comments_service .

COPY environment/ ./environment/


EXPOSE 30333


CMD ["./comments_service"] 