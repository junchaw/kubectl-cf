# syntax=docker/dockerfile:1.2
FROM golang:1-alpine as builder

RUN apk --no-cache --no-progress add git ca-certificates make \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

# syntax=docker/dockerfile:1.2
# Create a minimal container to run a Golang static binary
FROM alpine:3.21.3

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY kubectl-cf /

ENTRYPOINT ["/kubectl-cf"]
EXPOSE 80
