FROM golang:1.24.2-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOWORK=off go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/pvmbg-mock
FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/service /usr/local/bin/service
USER app
EXPOSE 8082
ENTRYPOINT ["/usr/local/bin/service"]
