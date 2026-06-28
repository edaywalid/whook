# Build a static, cgo-free binary (the SQLite driver is pure Go). Building on
# BUILDPLATFORM and cross-compiling to TARGET* keeps multi-arch builds fast
# (no QEMU for the compile step).
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w -X main.version=${VERSION}" -o /out/whook ./cmd/whook

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
