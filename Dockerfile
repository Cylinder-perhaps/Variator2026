FROM golang:1.25-alpine AS builder

WORKDIR /usr/local/src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app ./cmd/app/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata bash

WORKDIR /app
COPY --from=builder /bin/app ./app
COPY --from=builder /usr/local/src/configs ./configs
COPY --from=builder /usr/local/src/api_v1_mvp.yaml ./api_v1_mvp.yaml

EXPOSE 9090
CMD ["./app"]