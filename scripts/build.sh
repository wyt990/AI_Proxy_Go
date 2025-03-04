#!/bin/bash

# 获取构建时间
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')

# 构建程序
go build -ldflags "-X 'AI_Proxy_Go/backend/internal/version.BuildTime=${BUILD_TIME}'" -o AI_Proxy_Go backend/cmd/main.go 