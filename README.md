# MoonTV Server

影视资源搜索聚合后端服务，提供多源搜索、用户管理和 API Key 鉴权。

## 功能

- 多源影视资源搜索与聚合（支持 SSE 流式返回）
- 用户注册/登录（JWT 鉴权）+ 邀请码机制
- API Key 管理（AES-256-GCM 加密）
- 后台管理面板（用户管理、全局源管理、统计）
- 内嵌静态 Web 管理界面

## 快速开始

### 环境变量

复制并修改环境配置：

```bash
cp .env.example .env
```

### Docker 部署（推荐）

```bash
docker compose up -d
```

### 从源码构建

```bash
go build -o moontv-server ./cmd/server
./moontv-server
```

## API

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/auth/login` | 无 | 登录 |
| POST | `/api/auth/register` | 无 | 注册（需邀请码） |
| POST | `/api/user/apikey` | JWT | 生成 API Key |
| GET | `/api/search?wd=关键词` | API Key | 搜索 |
| GET | `/api/search/sse?wd=关键词` | API Key | SSE 流式搜索 |
| GET | `/api/detail?url=...` | API Key | 获取详情 |
| GET | `/api/sources` | API Key | 源列表 |

## Docker 镜像

每次推送到 `main` 分支或创建 `v*` tag 时，GitHub Actions 自动构建并推送镜像到 GHCR：

```bash
docker pull ghcr.io/caigq99/moontv-server:main
```

## 技术栈

- Go + Gin
- SQLite (GORM)
- JWT + AES-256-GCM
