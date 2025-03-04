#!/bin/bash

# 获取构建时间
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')

# 运行程序
go run -ldflags "-X 'AI_Proxy_Go/backend/internal/version.BuildTime=${BUILD_TIME}'" backend/cmd/main.go 