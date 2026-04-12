FROM golang:1.26-alpine AS builder

WORKDIR /bh_mus-app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bh_mus-bin ./cmd/main.go

FROM alpine:3.22
RUN apk add --no-cache postgresql-client curl

WORKDIR /app

COPY --from=builder /bh_mus-app/bh_mus-bin /app/bh_mus-bin

ENTRYPOINT ["/app/bh_mus-bin"]
