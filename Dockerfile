FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/file-server .

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/file-server /app/file-server
COPY public /app/public

ENV PORT=8080
ENV FILE_STORAGE_DIR=/app/storage

RUN mkdir -p /app/storage && chown -R app:app /app

USER app
EXPOSE 8080
VOLUME ["/app/storage"]

CMD ["/app/file-server"]
