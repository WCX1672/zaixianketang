#!/bin/bash

# 创建日志目录
mkdir -p logs

# 启动媒体服务器
echo "Starting media server..."
cd media
go run main.go -config config.toml > ../logs/media.log 2>&1 &
MEDIA_PID=$!
cd ..

# 启动信令服务器
echo "Starting signaling server..."
cd signaling
go run main.go -config config.toml > ../logs/signaling.log 2>&1 &
SIGNALING_PID=$!
cd ..

# 启动课堂管理服务
echo "Starting classroom service..."
cd classroom
go run main.go -config config.toml > ../logs/classroom.log 2>&1 &
CLASSROOM_PID=$!
cd ..

# 启动API网关
echo "Starting API gateway..."
cd gateway
go run main.go -config config.toml > ../logs/gateway.log 2>&1 &
GATEWAY_PID=$!
cd ..

echo "All services started:"
echo " - Media server PID: $MEDIA_PID"
echo " - Signaling server PID: $SIGNALING_PID"
echo " - Classroom service PID: $CLASSROOM_PID"
echo " - API gateway PID: $GATEWAY_PID"

# 保存PID到文件
echo $MEDIA_PID > .pids
echo $SIGNALING_PID >> .pids
echo $CLASSROOM_PID >> .pids
echo $GATEWAY_PID >> .pids

echo "System is running. Access at: http://localhost:8000"