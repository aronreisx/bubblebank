# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go

# Final stage
FROM gcr.io/distroless/static

USER nonroot:nonroot

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
