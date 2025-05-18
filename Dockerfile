FROM golang:1-alpine as builder

RUN apk --no-cache --no-progress add git ca-certificates make \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

WORKDIR /app

# Download go modules
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN make build

# Create a minimal container to run a Golang static binary
FROM alpine:3.21.3

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/kubectl-cf .

ENTRYPOINT ["/kubectl-cf"]
EXPOSE 8080
