# 第一阶段：构建 Go 服务器
FROM golang:1.20 AS go-build
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go clean -modcache
RUN go build .

# 第二阶段：构建最终镜像
FROM gcr.io/distroless/static:nonroot
COPY --from=go-build /go/src/app/ /app
EXPOSE 8081
USER nonroot:nonroot
CMD ["/app"]
