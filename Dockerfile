FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /matching ./cmd/matching

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /matching .

CMD ["./matching"]