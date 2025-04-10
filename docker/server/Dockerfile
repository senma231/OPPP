FROM golang:1.18-alpine AS builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache gcc musl-dev

# 复制 Go 模块定义
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o p3-server ./cmd

# 最终镜像
FROM alpine:3.16

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 复制构建产物
COPY --from=builder /build/p3-server .
COPY config.yaml .

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080 27182/udp 27183/udp 27184

# 运行应用
CMD ["./p3-server", "-config", "config.yaml"]
