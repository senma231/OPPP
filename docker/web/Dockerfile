# 构建阶段
FROM node:16-alpine AS builder

WORKDIR /app

# 复制前端代码
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# 构建后端
FROM golang:1.18-alpine AS backend-builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache gcc musl-dev

# 复制 Go 模块定义
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制源代码
COPY backend/ ./

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o p3-web-api .

# 最终镜像
FROM nginx:1.21-alpine

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 复制前端构建产物
COPY --from=builder /app/build /usr/share/nginx/html

# 复制后端构建产物
COPY --from=backend-builder /build/p3-web-api /app/

# 复制 Nginx 配置
COPY docker/web/nginx.conf /etc/nginx/conf.d/default.conf

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 80

# 启动脚本
COPY docker/web/start.sh /app/
RUN chmod +x /app/start.sh

CMD ["/app/start.sh"]
