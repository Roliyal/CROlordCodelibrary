# 第一阶段：构建 Go 服务器
FROM golang:1.19-alpine AS go-build
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .


# 第二阶段：构建最终镜像
FROM gcr.io/distroless/static:nonroot
COPY --from=go-build /go/src/app/app /app
EXPOSE 8081
USER nonroot:nonroot
CMD ["/app"]
