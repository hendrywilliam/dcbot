FROM golang:1.23.2-alpine3.20 AS builder
WORKDIR /app
COPY . .
RUN go build -o main ./cmd/bot

FROM alpine:3.20.3
WORKDIR /app
RUN apk add --no-cache ffmpeg
COPY --from=builder /app/main .
COPY --from=builder /app/media ./media
COPY .env .
EXPOSE 8080
CMD ["/app/main"]
