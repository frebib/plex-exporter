FROM alpine:latest as certificates

RUN apk update && \
    apk add --no-cache ca-certificates \
    && update-ca-certificates

# ~~~~~~~~~~~~~~

FROM golang as builder
WORKDIR /build
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/plex_exporter .

# ~~~~~~~~~~~~~~

FROM scratch
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/plex_exporter /bin/plex_exporter

ENTRYPOINT ["/bin/plex_exporter"]
