# FROM golang:1.23.0-alpine3.20 AS builder

# WORKDIR /usr/local/src
# RUN apk --no-cache add bash make gcc gettext git musl-dev
# COPY go.mod go.sum ./
# RUN go mod download
# COPY ./internal ./internal
# COPY ./cmd ./cmd
# RUN go build -o ./bin/app ./cmd/app/main.go

# FROM alpine
# RUN apk --no-cache add make bash ca-certificates
# WORKDIR /usr/local/src
# COPY --from=builder /usr/local/src/bin/app ./
# COPY ./configs ./configs

# EXPOSE 8080
# CMD ["./app"]


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
# COPY --from=builder /usr/local/src/configs ./configs
COPY --from=builder /usr/local/src/api_v1_mvp.yaml ./api_v1_mvp.yaml
EXPOSE 8080
CMD ["./app"]