FROM golang:1.25-alpine AS builder

WORKDIR /usr/local/src

COPY go.mod go.sum ./
RUN go mod download

# Устанавливаем goose для миграций
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app ./cmd/app/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata bash

WORKDIR /app

COPY --from=builder /bin/app ./app
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /usr/local/src/configs ./configs
COPY --from=builder /usr/local/src/migrations ./migrations
COPY --from=builder /usr/local/src/api_v1_mvp.yaml ./api_v1_mvp.yaml
COPY --from=builder /usr/local/src/scripts/docker-entrypoint.sh ./docker-entrypoint.sh
RUN chmod +x ./docker-entrypoint.sh

EXPOSE 9090
ENTRYPOINT ["./docker-entrypoint.sh"]