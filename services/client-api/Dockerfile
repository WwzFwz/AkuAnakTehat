FROM golang:1.24.2-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /service ./cmd/client-api
FROM alpine:3.21
RUN addgroup -g 10001 app && adduser -D -u 10001 -G app app
COPY --from=build /service /service
USER app
EXPOSE 8080
ENTRYPOINT ["/service"]
