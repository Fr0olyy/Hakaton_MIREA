FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o /agent-server ./cmd/server

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /agent-server /agent-server
EXPOSE 8090
CMD ["/agent-server"]
