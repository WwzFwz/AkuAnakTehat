FROM golang:1.24.2-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/service ./cmd/pemda-portal

FROM alpine:3.21
RUN mkdir -p /data && chown 10001:10001 /data
USER 10001:10001
COPY --from=builder /out/service /usr/local/bin/service
EXPOSE 8093
ENTRYPOINT ["/usr/local/bin/service"]
