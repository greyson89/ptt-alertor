FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ptt-alertor .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /

RUN mkdir -p /storage
COPY public ./public
COPY --from=builder /ptt-alertor /ptt-alertor

EXPOSE 9090

ENTRYPOINT ["/ptt-alertor"]
