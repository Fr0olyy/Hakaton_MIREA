FROM golang:1.22-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/backend ./cmd/api

FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/backend /app/backend
COPY migrations /app/migrations
EXPOSE 8080

CMD ["/app/backend"]
