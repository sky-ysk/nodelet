#!/bin/bash
while true
do
    # 执行 TestCreateJson1 程序
    go run ./test/scene3/createTaskForScene3.go

    # 等待一分钟（60秒）
    sleep 10
done