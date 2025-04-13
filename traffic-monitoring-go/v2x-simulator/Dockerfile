# v2x-simulator/Dockerfile
FROM golang:1.19 AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o v2x-simulator ./app

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/v2x-simulator .

ENTRYPOINT ["./v2x-simulator"]
