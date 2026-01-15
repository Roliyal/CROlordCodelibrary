FROM golang:1.22-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/go-gateway ./main.go

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=build /out/go-gateway /app/go-gateway

EXPOSE 8080
ENTRYPOINT ["/app/go-gateway"]
