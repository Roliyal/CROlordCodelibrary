# Build + run go-service
FROM golang:1.21 AS build
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/go-service ./...

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/go-service /app/go-service
# logs path is configurable; default is logs/app.log
ENTRYPOINT ["/app/go-service"]
