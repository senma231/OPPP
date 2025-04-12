FROM golang:1.20-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git make

# 复制源代码
COPY . .

# 构建客户端
RUN make client

# 最终镜像
FROM alpine:3.18

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 复制二进制文件和配置
COPY --from=builder /app/bin/p3-client /app/
COPY --from=builder /app/client/config.example.yaml /app/config.yaml

# 创建数据目录
RUN mkdir -p /app/data

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 27182/udp 27183/udp

# 设置入口点
ENTRYPOINT ["/app/p3-client"]
CMD ["-config", "/app/config.yaml"]
