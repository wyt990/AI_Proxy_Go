## 构建说明

## 运行说明

开发时运行程序：

```bash
# Linux/Mac
chmod +x scripts/run.sh
./scripts/run.sh

# Windows
go run -ldflags "-X 'AI_Proxy_Go/backend/internal/version.BuildTime=%date% %time%'" backend/cmd/main.go
```

生产环境运行已构建的程序：

```bash
# Linux/Mac
./AI_Proxy_Go

# Windows
AI_Proxy_Go.exe
```

要构建程序，请运行：

```bash
# Linux/Mac
chmod +x scripts/build.sh
./scripts/build.sh

# Windows
go build -ldflags "-X 'AI_Proxy_Go/backend/internal/version.BuildTime=%date% %time%'" -o AI_Proxy_Go.exe backend/cmd/main.go
```

这将生成一个包含版本号和构建时间的可执行文件。 