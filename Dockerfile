FROM golang:1.22 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o plex-overlay main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
  ca-certificates libjpeg62-turbo libpng16-16 && \
  rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/plex-overlay /app/plex-overlay
COPY transparent.png /app/transparent.png

WORKDIR /app
EXPOSE 8080 8081

CMD ["/app/plex-overlay"]
