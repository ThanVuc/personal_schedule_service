# build stage
FROM golang:alpine AS builder
RUN apk add --no-cache

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o personal_schedule_service ./main.go

# stage 2
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/personal_schedule_service .
RUN chmod +x /app/personal_schedule_service
ENTRYPOINT ["./personal_schedule_service"]
