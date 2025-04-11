#!/bin/bash

# 设置颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 创建测试目录
mkdir -p test/server test/client1 test/client2

# 复制配置文件
echo -e "${YELLOW}复制配置文件...${NC}"
cp server/config.example.yaml test/server/config.yaml
cp client/config.example.yaml test/client1/config.yaml
cp client/config.example.yaml test/client2/config.yaml

# 修改客户端配置
echo -e "${YELLOW}修改客户端配置...${NC}"
sed -i '' 's/my-node/client1/g' test/client1/config.yaml
sed -i '' 's/your-node-token/test-token1/g' test/client1/config.yaml
sed -i '' 's/my-node/client2/g' test/client2/config.yaml
sed -i '' 's/your-node-token/test-token2/g' test/client2/config.yaml

# 启动测试数据库
echo -e "${YELLOW}启动测试数据库...${NC}"
docker run -d --name p3-test-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=p3_test \
  -p 5432:5432 \
  postgres:14

# 等待数据库启动
echo -e "${YELLOW}等待数据库启动...${NC}"
sleep 5

# 运行服务端单元测试
echo -e "${YELLOW}运行服务端单元测试...${NC}"
cd server
go test -v ./...
if [ $? -eq 0 ]; then
  echo -e "${GREEN}服务端单元测试通过${NC}"
else
  echo -e "${RED}服务端单元测试失败${NC}"
  exit 1
fi
cd ..

# 运行客户端单元测试
echo -e "${YELLOW}运行客户端单元测试...${NC}"
cd client
go test -v ./...
if [ $? -eq 0 ]; then
  echo -e "${GREEN}客户端单元测试通过${NC}"
else
  echo -e "${RED}客户端单元测试失败${NC}"
  exit 1
fi
cd ..

# 构建服务端和客户端
echo -e "${YELLOW}构建服务端和客户端...${NC}"
go build -o bin/p3-server server/main.go
go build -o bin/p3-client client/main.go

# 启动服务端
echo -e "${YELLOW}启动服务端...${NC}"
cd test/server
../../bin/p3-server -config config.yaml &
SERVER_PID=$!
cd ../..

# 等待服务端启动
echo -e "${YELLOW}等待服务端启动...${NC}"
sleep 5

# 启动客户端 1
echo -e "${YELLOW}启动客户端 1...${NC}"
cd test/client1
../../bin/p3-client -config config.yaml &
CLIENT1_PID=$!
cd ../..

# 启动客户端 2
echo -e "${YELLOW}启动客户端 2...${NC}"
cd test/client2
../../bin/p3-client -config config.yaml &
CLIENT2_PID=$!
cd ../..

# 等待客户端启动
echo -e "${YELLOW}等待客户端启动...${NC}"
sleep 5

# 运行集成测试
echo -e "${YELLOW}运行集成测试...${NC}"
cd test
go test -v ./integration/...
INTEGRATION_TEST_RESULT=$?
cd ..

# 清理
echo -e "${YELLOW}清理...${NC}"
kill $SERVER_PID
kill $CLIENT1_PID
kill $CLIENT2_PID
docker stop p3-test-db
docker rm p3-test-db

# 输出结果
if [ $INTEGRATION_TEST_RESULT -eq 0 ]; then
  echo -e "${GREEN}集成测试通过${NC}"
  exit 0
else
  echo -e "${RED}集成测试失败${NC}"
  exit 1
fi
