FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/verifier ./cmd/verifier

FROM alpine:3.22

RUN adduser -D -H -u 10001 verifier
COPY --from=build /out/verifier /usr/local/bin/verifier

USER verifier
EXPOSE 8081
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:8081/health >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/verifier"]
