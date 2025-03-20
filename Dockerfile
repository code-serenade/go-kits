# 使用 Alpine 3.21 作为基础镜像
FROM golang:alpine3.21 AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具（git、curl）和 swag
RUN apk add --no-cache git curl && \
    go install github.com/swaggo/swag/cmd/swag@latest

# 复制项目文件
COPY . .

# 生成 Swagger 文档
RUN swag init

# 编译 Go 应用
RUN go build -o main .

# 创建一个小体积运行时镜像
FROM alpine:3.21

# 设置工作目录
WORKDIR /app

# 复制编译后的二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs

# 运行应用
CMD ["./main"]
