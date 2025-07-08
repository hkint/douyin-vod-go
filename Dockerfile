# --- 构建阶段 ---
# 使用官方的 Golang 镜像作为构建环境
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 Go 模块文件
COPY go.mod ./
# 下载依赖
RUN go mod download

# 复制所有源代码
COPY . .

# 构建 Go 应用。
# CGO_ENABLED=0 禁用 CGO，使其成为静态二进制文件
# -ldflags="-w -s" 减小二进制文件大小
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o douyin-server .

# --- 运行阶段 ---
# 使用一个非常小的基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/douyin-server .

# 复制前端静态文件
# COPY --from=builder /app/static /app/static

# 暴露服务端口
EXPOSE 8080

# 设置容器启动时执行的命令
ENTRYPOINT ["/app/douyin-server"]