# Build stage
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/out/server ./agentservice/

# Runtime stage
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /app/out .

ENTRYPOINT ["/app/server"]