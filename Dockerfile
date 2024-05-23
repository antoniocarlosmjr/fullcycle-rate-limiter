FROM golang:1.22.1 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags '-s -w' \
    -o rate-limiter-app cmd/server/main.go

FROM scratch
COPY --from=build /app/rate-limiter-app .
CMD ["./rate-limiter-app"]