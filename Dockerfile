# Build a static, cgo-free binary (the SQLite driver is pure Go).
FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/whook ./cmd/whook

FROM alpine:3.20
RUN apk add --no-cache wget \
    && adduser -D -u 10001 whook \
    && mkdir -p /data \
    && chown whook /data

USER whook
WORKDIR /data
COPY --from=build /out/whook /usr/local/bin/whook

ENV WHOOK_ADDR=":8080" \
    WHOOK_DATABASE_URL="file:/data/whook.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"

EXPOSE 8080
ENTRYPOINT ["whook"]
