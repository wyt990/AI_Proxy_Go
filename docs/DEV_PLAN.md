# 五月天AI 代理平台开发计划

## 一、技术框架

### 1.1 技术栈
| 组件         | 技术选型                  | 版本     | 用途说明                 |
|--------------|--------------------------|----------|-------------------------|
| 后端框架     | Go                       | 1.21+    | 高性能服务端开发         |
| Web框架      | Gin                      | 1.9.1    | API路由及中间件处理      |
| 前端框架     | React                    | 18.2+    | 管理界面开发             |
| UI库         | Ant Design Pro           | 6.0.0+   | 企业级后台界面组件       |
| ORM          | GORM                     | 2.0+     | 数据库操作               |
| 缓存         | Redis                    | 7.0+     | 会话及热点数据缓存       |
| 部署         | Docker                   | 24.0+    | 容器化部署               |

### 1.2 目录结构
/AI_Proxy_Go/
├── backend/ # 后端主目录
│ ├── cmd/ # CLI命令入口
│ │ └── main.go # 主程序入口
│ ├── config/ # 配置管理
│ │ ├── config.go # 配置结构定义
│ │ └── config.yaml # 配置文件
│ ├── internal/ # 内部模块
│ │ ├── api/ # API控制器
│ │ │ ├── auth.go # 认证相关接口
│ │ │ ├── proxy.go # 代理相关接口
│ │ │ ├── admin.go # 管理接口
│ │ │ ├── user_handler.go # 用户管理接口
│ │ │ ├── captcha_handler.go # 验证码接口
│ │ │ ├── provider.go # 服务商管理接口
│ │ │ └── model.go # 模型管理接口
│ │ ├── service/ # 业务逻辑层
│ │ │ ├── auth/ # 认证服务
│ │ │ ├── proxy/ # 代理服务
│ │ │ ├── admin/ # 管理服务
│ │ │ ├── user/ # 用户服务
│ │ │ ├── captcha/ # 验证码服务
│ │ │ ├── provider/ # 服务商管理服务
│ │ │ └── model/ # 模型管理服务
│ │ └── model/ # 数据模型
│ │   ├── user.go # 用户模型
│ │   ├── captcha.go # 验证码模型
│ │   ├── AI_provider.go # AI服务商模型
│ │   └── AI_model.go # AI模型
│ └── pkg/ # 公共包
│   └── captcha/ # 验证码工具包
├── frontend/ # 前端主目录
│ ├── static/ # 静态资源
│ │ ├── css/ # 样式文件
│ │ │ ├── home.css # 主页样式
│ │ │ ├── login.css # 登录页样式
│ │ │ ├── install.css # 安装页样式
│ │ │ ├── users.css # 用户管理页样式
│ │ │ ├── providers.css # 服务商管理页样式
│ │ │ └── models.css # 模型管理页样式
│ │ ├── js/ # JavaScript文件
│ │ │ ├── home.js # 主页脚本
│ │ │ ├── login.js # 登录脚本
│ │ │ ├── install.js # 安装脚本
│ │ │ ├── users.js # 用户管理脚本
│ │ │ ├── common.js # 公共脚本
│ │ │ ├── providers.js # 服务商管理脚本
│ │ │ └── models.js # 模型管理脚本
│ │ └── images/ # 图片资源
│ └── templates/ # 模板文件
│   ├── common/ # 公共组件
│   │ ├── header.html # 头部组件
│   │ ├── sidebar.html # 侧边栏组件
│   │ └── footer.html # 页脚组件
│   ├── layouts/ # 布局模板
│   │ └── base.html # 基础布局
│   ├── home/ # 主页模板
│   │ └── index.html # 主页
│   ├── login/ # 登录页模板
│   │ └── index.html # 登录页
│   ├── install/ # 安装页模板
│   │ └── index.html # 安装页
│   ├── users/ # 用户管理页面
│   │ └── index.html # 用户列表页面
│   ├── providers/ # 服务商管理页面
│   │ └── index.html # 服务商管理页面
│   └── models/ # 模型管理页面
│     └── index.html # 模型管理页面
└── docs/ # 文档目录
└── scripts/ # 脚本文件

### 1.3 文件说明
| 文件路径 | 用途说明 | 依赖关系 |
|---------|---------|----------|
| cmd/main.go | 程序入口点 | config, internal/api |
| config/config.go | 配置定义 | - |
| internal/api/auth.go | 认证接口 | service/auth |
| internal/api/proxy.go | 代理接口 | service/proxy |
| internal/api/user_handler.go | 用户管理接口 | model/user, service/user |
| internal/api/captcha_handler.go | 验证码接口 | pkg/captcha, model/captcha |
| internal/api/provider.go | 服务商管理接口 | model/AI_provider |
| internal/api/model.go | 模型管理接口 | model/AI_model |
| internal/model/user.go | 用户数据模型 | - |
| internal/model/captcha.go | 验证码数据模型 | - |
| internal/model/AI_provider.go | AI服务商模型 | - |
| internal/model/AI_model.go | AI模型 | - |
| pkg/captcha/captcha.go | 验证码生成工具 | - |
| frontend/templates/users/index.html | 用户管理页面 | layouts/base |
| frontend/static/js/users.js | 用户管理脚本 | common.js |
| frontend/static/css/users.css | 用户管理样式 | - |
| frontend/static/js/common.js | 公共功能脚本 | - |
| frontend/templates/providers/index.html | 服务商管理页面 | layouts/base |
| frontend/static/js/providers.js | 服务商管理脚本 | common.js |
| frontend/static/css/providers.css | 服务商管理样式 | - |
| frontend/templates/models/index.html | 模型管理页面 | layouts/base |
| frontend/static/js/models.js | 模型管理脚本 | common.js |
| frontend/static/css/models.css | 模型管理样式 | - |

## 二、数据库设计

### 2.1 表结构设计
1. **系统配置表 (system_settings)**
```sql
CREATE TABLE system_settings (
id BIGINT PRIMARY KEY,
config_key VARCHAR(50),
value TEXT,
description TEXT COMMENT '配置项说明',
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
deleted_at DATETIME
);
```

系统配置表说明：
- id: 配置项ID
- config_key: 配置项键名
- value: 配置项值
- description: 配置项说明
- created_at: 创建时间
- updated_at: 更新时间
- deleted_at: 软删除时间

2. **用户表 (users)**
```sql
CREATE TABLE users (
id BIGINT PRIMARY KEY,
username VARCHAR(50),
password_hash VARCHAR(100),
name VARCHAR(50),
email VARCHAR(100),
role VARCHAR(20) DEFAULT 'user',
is_active TINYINT(1) DEFAULT 1,
last_login DATETIME,
created_at DATETIME(3),
updated_at DATETIME(3)
);
```

用户表说明：
- id: 用户ID
- username: 用户名
- password_hash: 密码哈希
- name: 用户姓名
- email: 电子邮箱
- role: 用户角色，默认为'user'
- is_active: 是否激活，1表示激活，0表示未激活
- last_login: 最后登录时间
- created_at: 创建时间，精确到毫秒
- updated_at: 更新时间，精确到毫秒

3. **验证码表 (captcha)**
```sql
CREATE TABLE captcha (
id BIGINT PRIMARY KEY,
code VARCHAR(10) NOT NULL,
image_data TEXT NOT NULL,
ip VARCHAR(45) NOT NULL,
user_agent VARCHAR(255),
session_id VARCHAR(64),
type VARCHAR(20) DEFAULT 'LOGIN',
expired_at DATETIME NOT NULL,
used TINYINT(1) DEFAULT 0,
used_at DATETIME(3),
fail_count BIGINT,
created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP
);
```

验证码表说明：
- id: 验证码ID
- code: 验证码内容
- image_data: 验证码图片数据(Base64)
- ip: 请求IP
- user_agent: 用户代理
- session_id: 会话ID
- type: 验证码类型(LOGIN/REGISTER/RESET)
- expired_at: 过期时间
- used: 是否已使用
- used_at: 使用时间
- fail_count: 验证失败次数
- created_at: 创建时间
- updated_at: 更新时间

4. **AI服务商表 (ai_providers)**
```sql
CREATE TABLE ai_providers (
id BIGINT PRIMARY KEY,
name VARCHAR(50) NOT NULL,
type VARCHAR(20) NOT NULL,
base_url VARCHAR(255) NOT NULL,
models TEXT NOT NULL COMMENT '支持的模型列表',
credentials TEXT NOT NULL COMMENT '加密存储的凭证信息',
auth_type VARCHAR(20) DEFAULT 'API_KEY',
headers TEXT,
rate_limit BIGINT DEFAULT 60,
token_limit BIGINT DEFAULT 0,
token_used BIGINT DEFAULT 0,
priority BIGINT DEFAULT 0,
timeout BIGINT DEFAULT 30,
retry_count BIGINT DEFAULT 3,
retry_interval BIGINT DEFAULT 1,
is_active BOOLEAN DEFAULT TRUE,
status VARCHAR(20) DEFAULT 'NORMAL',
last_error TEXT,
last_check_time DATETIME,
config TEXT,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
deleted_at DATETIME
);
```

AI服务商表说明：
- id: 服务商ID
- name: 服务商名称
- type: 服务商类型(OPENAI/ANTHROPIC/GOOGLE/BAIDU/CUSTOM)
- base_url: API基础URL
- models: 支持的模型列表(JSON格式)
- credentials: 加密存储的凭证信息(JSON格式)
- auth_type: 认证类型(API_KEY/OAUTH/JWT)
- headers: 自定义请求头(JSON格式)
- rate_limit: 每分钟请求限制
- token_limit: 令牌使用限制(0表示无限制)
- token_used: 已使用的令牌数
- priority: 优先级(用于负载均衡)
- timeout: 请求超时时间(秒)
- retry_count: 重试次数
- retry_interval: 重试间隔(秒)
- is_active: 是否启用
- status: 状态(NORMAL/RATE_LIMITED/ERROR)
- last_error: 最后一次错误信息
- last_check_time: 最后一次健康检查时间
- config: 额外配置(JSON格式)
- created_at: 创建时间
- updated_at: 更新时间
- deleted_at: 软删除时间

## 三、功能开发计划

### 3.1 后端开发阶段（6周）
1. **基础框架搭建**（1周）
   - 项目结构初始化
   - 配置管理实现
   - 数据库连接配置
   - 基础中间件开发

2. **核心功能开发**（2周）
   - 用户认证系统
   - AI代理功能
   - 服务商管理
   - 请求转发实现

3. **安装程序开发**（1周）
   - 安装检测逻辑
   - 数据库初始化
   - 配置文件生成
   - 管理员账户创建

4. **管理功能开发**（2周）
   - 用户管理接口
   - 服务商配置
   - 系统监控
   - 日志记录
   - 模型管理接口

### 3.2 前端开发阶段（4周）
1. **登录界面**（1周）
   - 登录表单
   - 验证码集成
   - 记住密码
   - 忘记密码

2. **管理界面**（2周）
   - 布局设计
   - 导航菜单
   - 用户管理
   - 服务商管理
   - 模型管理

3. **监控界面**（1周）
   - 数据统计
   - 图表展示
   - 实时监控
   - 日志查看

## 四、进度检查对照表

### 第一阶段：基础框架搭建

#### 4.1.1 第1周：项目初始化 [进行中]
1. ✅ 项目结构搭建
   - ✅ 创建目录结构
   - ✅ 配置开发环境
   - ✅ 初始化Git仓库
   - ✅ 添加基础依赖

2. ✅ 配置系统实现
   - ✅ 配置文件结构设计
   - ✅ 配置加载机制
   - ✅ 默认配置支持
   - ✅ 配置验证器

3. ⏳ 安装程序与数据库初始化
   - ✅ 安装检测逻辑
     * ✅ install.lock文件检查
     * ✅ 环境依赖检测
     * ✅ 权限检查
   - ✅ 数据库配置与初始化
     * ✅ 数据库连接配置
     * ✅ 表结构创建
     * ✅ 初始数据准备
     * ✅ 连接池配置
   - ✅ 系统初始化
     * ✅ 管理员账户创建
     * ✅ 基础配置生成
     * ✅ 安装锁创建
   - ✅ 事务管理
     * ✅ 安装回滚机制
     * ✅ 错误处理
     * ✅ 日志记录

4. ✅ Web服务器
   - ✅ 基础路由配置
   - ✅ 中间件集成
   - ✅ 安装检测中间件
   - ✅ 静态文件服务

5. ✅ 前端基础框架
   - ✅ 基础布局模板
   - ✅ 公共组件封装
   - ✅ 响应式布局
   - ✅ 主题样式定义
   - ✅ 静态资源组织

6. ✅ 用户认证基础
   - ✅ 登录页面
   - ✅ Token管理
   - ✅ 会话控制
   - ✅ 权限验证

### 已完成功能列表
1. 基础框架
   - ✅ 项目目录结构
   - ✅ Go模块初始化
   - ✅ 依赖管理

2. 配置系统
   - ✅ 配置文件结构
   - ✅ 配置加载机制
   - ✅ 默认配置支持

3. Web服务器
   - ✅ 服务器配置
   - ✅ 基础路由
   - ✅ 中间件框架
   - ✅ 安装检测

4. 安装系统
   - ✅ 安装状态检测
   - ✅ 安装页面路由
   - ✅ API路由定义
   - ✅ 数据库初始化
   - ✅ 管理员配置


### 第二阶段：前端开发

#### 4.3.1 第6-7周：管理界面 [未开始]
1. ⏳ 登录模块
   - ✅ 登录表单
   - ✅ 验证码集成
   - ⏳ 权限控制
   - ⏳ 记住密码

2. ⏳ 管理功能
   - ✅用户管理
   - 服务商管理
   - ⏳ 系统配置
   - 日志查看

#### 4.3.2 第8-9周：监控系统 [未开始]
1. ⏳ 监控功能
   - 性能监控
   - ⏳ 请求统计
   - 错误追踪
   - 系统日志

2. ⏳ 数据可视化
   - ⏳ 统计图表
   - 实时监控
   - 报表导出
   - 告警设置

### 第三阶段：核心功能开发

#### 4.2.1 第3-4周：代理服务 [未开始]
1. ⏳ 代理核心功能
   - 请求转发器
   - 响应处理器
   - 错误重试机制
   - 限流器实现

2. ⏳ AI服务商管理
   - 服务商配置
   - 负载均衡
   - 健康检查
   - 模型路由


### 第四阶段：测试与优化

#### 4.4.1 第10周：测试 [未开始]
1. ⏳ 测试用例
   - 单元测试
   - 集成测试
   - 性能测试
   - 安全测试

2. ⏳ 文档完善
   - API文档
   - 部署文档
   - 使用手册
   - 开发文档

## 五、注意事项

### 5.1 代码规范
1. **命名规范**
   - Go代码：使用驼峰命名
   - SQL：使用下划线命名
   - 前端组件：大写开头

2. **注释要求**
   - 所有导出函数必须有注释
   - 复杂逻辑需要说明
   - 配置项需要说明用途

3. **安全规范**
   - 密码必须加密存储
   - API需要认证和鉴权
   - 敏感信息必须脱敏

### 5.2 开发流程
1. **版本控制**
   - 使用Git管理代码
   - 遵循Git Flow工作流
   - 每个功能独立分支

2. **代码审查**
   - 提交前自测
   - 团队代码审查
   - 合并前确认测试

3. **测试要求**
   - 单元测试覆盖率>80%
   - 必须包含集成测试
   - 提供性能测试报告

## 功能模块

### 1. 用户管理
- [x] 用户列表
  - 支持分页显示
  - 可配置每页显示数量(20/50/100)
  - 显示用户基本信息、状态、创建时间等
- [x] 用户添加/编辑/删除
  - 支持创建新用户，设置用户名、密码、姓名、邮箱等信息
  - 支持编辑现有用户信息，可选择是否修改密码
  - 支持删除用户，删除前有确认提示
- [x] 用户权限管理
  - 支持设置用户角色(普通用户/管理员)
  - 支持启用/禁用用户状态
  - 使用 bcrypt 加密存储密码，保证安全性

### 2. 服务商管理
- [x] 服务商列表
  - 支持分页显示
  - 显示服务商名称、类型、状态等信息
- [x] 服务商添加/编辑/删除
  - 支持创建新服务商，设置名称、类型、API URL等信息
  - 支持编辑现有服务商信息
  - 支持删除服务商，删除前有确认提示
- [x] 服务商配置
  - 支持设置服务商的认证类型、请求头、限流等配置

### 3. 模型管理
- [x] 模型列表
  - 支持分页显示
  - 显示模型名称、服务商、状态等信息
- [x] 模型添加/编辑/删除
  - 支持创建新模型，设置名称、服务商、配置等信息
  - 支持编辑现有模型信息
  - 支持删除模型，删除前有确认提示

## 技术实现

### 前端实现
#### 用户管理模块
- 使用原生 JavaScript 实现，无需额外框架
- 采用模块化设计，分离业务逻辑和 UI 交互
- 实现功能：
  1. 用户列表展示和分页
  2. 用户表单验证和提交
  3. 模态框交互
  4. 用户状态和角色管理
- 关键文件：
  - frontend/templates/users/index.html: 用户管理页面模板
  - frontend/static/js/users.js: 用户管理相关的 JavaScript 代码
  - frontend/static/css/users.css: 用户管理页面样式

#### 服务商管理模块
- 使用原生 JavaScript 实现，无需额外框架
- 采用模块化设计，分离业务逻辑和 UI 交互
- 实现功能：
  1. 服务商列表展示和分页
  2. 服务商表单验证和提交
  3. 模态框交互
  4. 服务商配置管理
- 关键文件：
  - frontend/templates/providers/index.html: 服务商管理页面模板
  - frontend/static/js/providers.js: 服务商管理相关的 JavaScript 代码
  - frontend/static/css/providers.css: 服务商管理页面样式

#### 模型管理模块
- 使用原生 JavaScript 实现，无需额外框架
- 采用模块化设计，分离业务逻辑和 UI 交互
- 实现功能：
  1. 模型列表展示和分页
  2. 模型表单验证和提交
  3. 模态框交互
  4. 模型配置管理
- 关键文件：
  - frontend/templates/models/index.html: 模型管理页面模板
  - frontend/static/js/models.js: 模型管理相关的 JavaScript 代码
  - frontend/static/css/models.css: 模型管理页面样式

### 后端实现
#### 用户管理模块
- 采用 RESTful API 设计
- 实现功能：
  1. 用户 CRUD 操作
  2. 分页查询
  3. 密码加密存储
  4. 数据验证
- 关键文件：
  - backend/internal/api/user_handler.go: 用户管理 API 处理器
  - backend/internal/model/user.go: 用户模型定义

#### 服务商管理模块
- 采用 RESTful API 设计
- 实现功能：
  1. 服务商 CRUD 操作
  2. 配置管理
  3. 数据验证
- 关键文件：
  - backend/internal/api/provider.go: 服务商管理 API 处理器
  - backend/internal/model/AI_provider.go: AI服务商模型定义

#### 模型管理模块
- 采用 RESTful API 设计
- 实现功能：
  1. 模型 CRUD 操作
  2. 配置管理
  3. 数据验证
- 关键文件：
  - backend/internal/api/model.go: 模型管理 API 处理器
  - backend/internal/model/AI_model.go: AI模型定义

### 数据库设计
#### 用户表 (users)
| 字段          | 类型         | 说明                    |
|--------------|--------------|------------------------|
| id           | bigint       | 主键                    |
| username     | varchar(50)  | 用户名                  |
| password_hash| varchar(100) | 密码哈希                |
| name         | varchar(50)  | 姓名                    |
| email        | varchar(100) | 邮箱                    |
| role         | varchar(20)  | 角色(user/admin)        |
| is_active    | tinyint(1)   | 是否启用                |
| last_login   | datetime     | 最后登录时间             |
| created_at   | datetime     | 创建时间                |
| updated_at   | datetime     | 更新时间                |

## API 接口
OpenAI: https://api.openai.com/v1/chat/completions
Anthropic: https://api.anthropic.com/v1/complete
Google: https://generativelanguage.googleapis.com/v1/models/gemini-pro:generateContent
### 用户管理接口
#### 获取用户列表
- 请求方式：GET
- 路径：/api/users
- 参数：
  - page: 页码
  - pageSize: 每页数量
- 返回：用户列表和总数

#### 创建用户
- 请求方式：POST
- 路径：/api/users
- 参数：用户信息(username, password, name, email, role, is_active)
- 返回：创建结果

#### 更新用户
- 请求方式：PUT
- 路径：/api/users/:id
- 参数：用户信息(可选 password)
- 返回：更新结果

#### 删除用户
- 请求方式：DELETE
- 路径：/api/users/:id
- 返回：删除结果

# 开发计划

## 一、基础设施

### 1. 系统安装
- [x] 安装向导
- [x] 数据库配置
- [x] Redis配置
- [x] 系统初始化

### 2. 用户认证
- [x] 登录功能
- [x] 验证码
- [x] JWT认证
- [x] 权限控制

## 二、服务商管理

### 1. 服务商配置
- [x] 服务商增删改查
- [x] 服务商类型管理
- [x] 基础URL配置
- [x] 请求头配置
- [x] 连接测试功能

### 2. 模型管理
- [x] 模型增删改查
- [x] 模型参数配置
- [x] 模型与服务商关联
- [ ] 上下文长度配置

### 3. 密钥管理
- [x] 密钥增删改查
- [x] 公钥/私钥类型管理
- [x] 密钥与服务商关联
- [x] 密钥使用限制
  - [x] 速率限制（每分钟请求数）
  - [x] 配额限制（总使用量）
  - [x] 过期时间设置
- [x] 密钥状态管理（启用/禁用）
- [x] 创建者信息记录
- [x] 重复密钥检查

## 三、代理服务（待开发）

### 1. 请求路由
- [ ] API路由规则配置
- [ ] 请求转发处理
- [ ] 响应处理规则

### 2. 中间件系统
- [ ] 密钥验证中间件
- [ ] 速率限制中间件
- [ ] 配额检查中间件
- [ ] 请求日志中间件

### 3. 响应处理
- [ ] 统一错误处理
- [ ] 响应格式化
- [ ] 计费统计处理

### 4. 监控统计
- [ ] 请求成功/失败统计
- [ ] 响应时间监控
- [ ] 使用量统计
- [ ] 费用统计

## 四、系统管理（待开发）

### 1. 系统配置
- [ ] 系统参数配置
- [ ] 日志级别配置
- [ ] 缓存配置

### 2. 监控面板
- [ ] 系统状态监控
- [ ] 性能监控
- [ ] 资源使用监控

### 3. 数据管理
- [ ] 数据备份
- [ ] 数据恢复
- [ ] 数据清理

## 五、其他功能（待开发）

### 1. 计费系统
- [ ] 计费规则配置
- [ ] 费用统计
- [ ] 账单生成

### 2. 报表系统
- [ ] 使用量报表
- [ ] 费用报表
- [ ] 性能报表

### 3. 通知系统
- [ ] 异常通知
- [ ] 额度预警
- [ ] 系统公告