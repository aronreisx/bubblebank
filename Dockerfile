FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main main.go

FROM alpine:3.19
WORKDIR /root/
COPY --from=builder /app/main .
ENV DB_IMAGE=${DB_IMAGE} \
    DB_VERSION=${DB_VERSION} \
    DB_PORT=${DB_PORT} \
    DB_USER=${DB_USER} \
    DB_PASS=${DB_PASS} \
    DB_NAME=${DB_NAME} \
    DB_HOST=${DB_HOST} \
    SERVER_PORT=${SERVER_PORT}
EXPOSE ${SERVER_PORT}
CMD ["./main"]
