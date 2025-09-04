# PostgreSQL stage
FROM postgres AS postgres
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=mypassword
ENV POSTGRES_DB=user_service_db

# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main ./cmd/main.go

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates postgresql-client
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
EXPOSE 8080
EXPOSE 5432

CMD ["./main"]
