#!/bin/sh

# 启动后端 API
/app/p3-web-api -port 8080 -mode release &

# 启动 Nginx
nginx -g "daemon off;"
