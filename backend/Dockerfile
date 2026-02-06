# Multi-stage Go build
FROM golang:1.21-alpine AS builder
WORKDIR /src

# install git for go modules that require it
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w" -o /app ./

# Final image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app /app
EXPOSE 8080
USER nobody:nobody
ENTRYPOINT ["/app"]
