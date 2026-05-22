# 部署说明

生产环境使用根目录的 `docker-compose.prod.yml`。这套配置包含：

- `postgres`：PostgreSQL 16，数据保存在 `postgres_data` volume。
- `migrate`：一次性迁移任务，按顺序执行 `migrations/*.up.sql`。
- `backend`：Gin API 服务，容器内监听 `3212`。
- `frontend`：Nginx 静态前端，只在 compose 内部暴露。
- `caddy`：公网入口，占用宿主机 `80/443`，自动申请 HTTPS 证书。

## VertexAI ADC 凭证

和 WcsTransfer 保持一致，默认使用宿主机路径：

```bash
sudo mkdir -p /secrets
sudo cp gcp-credentials.json /secrets/gcp-credentials.json
sudo chown 65532:65532 /secrets/gcp-credentials.json
sudo chmod 600 /secrets/gcp-credentials.json
```

后端容器内固定读取：

```text
/secrets/gcp-credentials.json
```

如果你想换宿主机路径，修改 `configs/.env.prod`：

```env
GCP_CREDENTIALS_FILE=/your/path/gcp-credentials.json
GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp-credentials.json
```

## 首次部署

```bash
git clone <repo-url> /home/wangcs/code/llm-gateway
cd /home/wangcs/code/llm-gateway
cp configs/.env.prod.example configs/.env.prod
vim configs/.env.prod
docker compose --env-file configs/.env.prod -f docker-compose.prod.yml up -d --build
```

上线前必须修改：

- `DOMAIN`
- `FRONTEND_ORIGIN`
- `ACME_EMAIL`
- `POSTGRES_PASSWORD`
- `AUTH_TOKEN_SECRET`
- `PROVIDER_KEY_SECRET`

## 从 WcsTransfer 切换

因为 WcsTransfer 的 Caddy 也占用 `80/443`，先停掉 WcsTransfer：

```bash
cd /home/wangcs/code/WcsTransfer
docker compose --env-file .env.prod -f docker-compose.prod.yml down
```

再启动本项目：

```bash
cd /home/wangcs/code/llm-gateway
docker compose --env-file configs/.env.prod -f docker-compose.prod.yml up -d --build
```

## 日常更新

```bash
cd /home/wangcs/code/llm-gateway
git pull
docker compose --env-file configs/.env.prod -f docker-compose.prod.yml up -d --build
```

## 查看状态和日志

```bash
docker compose --env-file configs/.env.prod -f docker-compose.prod.yml ps
docker compose --env-file configs/.env.prod -f docker-compose.prod.yml logs -f backend
docker compose --env-file configs/.env.prod -f docker-compose.prod.yml logs -f caddy
```

## 停止服务

```bash
docker compose --env-file configs/.env.prod -f docker-compose.prod.yml stop
```

如果要删除容器但保留数据卷：

```bash
docker compose --env-file configs/.env.prod -f docker-compose.prod.yml down
```
