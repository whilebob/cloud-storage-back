# Cloud Storage 云储存

### 自己写的屎山，将来会重构
## 项目结构(不是标准目录，没有按照Gin标准来写)

```
cloud-storage/
├── api/              # API 路由层
├── common/           # 公共组件
├── config/           # 配置文件
├── gorm-logger/      # 数据库日志
├── middleware/       # 中间件
├── model/            # 数据模型
├── service/          # 业务逻辑层
├── utils/            # 工具函数
├── main.go           # 入口文件
└── config/config.yaml       # 配置文件
```

## 快速开始

### 1. 配置环境

修改 `config/config.yaml` 中的数据库、Redis、MinIO 配置：

```yaml
mysql:
  host: localhost
  port: 3306
  username: root
  password: 123456
  dbname: cloud_storage

redis:
  host: localhost
  port: 6379
  password:
  db: 0

minio:
  host: localhost
  port: 9000
  access_key: admin
  secret_key: 12345678

HOST: localhost
PORT: 8080
```

### 2. 启动服务

```bash
go run main.go
```

服务将在 `http://localhost:3000` 启动。

## API 接口文档

### 认证接口

#### 注册用户
```
POST /user/register
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

#### 用户登录
```
POST /user/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

### 文件接口

#### 获取文件列表
```
GET /file/list
Authorization: Bearer <token>
```

#### 删除文件
```
DELETE /file/delete?id=<file_id>
Authorization: Bearer <token>
```

#### 下载文件
```
GET /file/download?id=<file_id>
Authorization: Bearer <token>
```

#### 预览文件
```
GET /file/preview?filename=<filename>
Authorization: Bearer <token>
```
### 上传接口

#### 单文件上传
```
POST /upload/file
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

#### 分片上传
```
POST /upload/chunk
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

## 许可证

MIT License

Copyright © 2026 Cloud Storage