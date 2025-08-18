# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go

# Final stage
FROM gcr.io/distroless/static:nonroot

USER nonroot:nonroot

COPY --from=builder /app/main .

EXPOSE ${SERVER_PORT}

CMD ["./main"]
