FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /bin/this-is-pun ./cmd/server

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /bin/this-is-pun /bin/this-is-pun
EXPOSE 8080
CMD ["/bin/this-is-pun"]
