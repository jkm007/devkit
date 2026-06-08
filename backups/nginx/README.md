# Nginx 配置备份

**备份时间**: 2026-06-08
**备份来源**: 远程服务器 123.57.201.44

## 文件说明

| 文件 | 来源 | 用途 |
|------|------|------|
| `remote-nginx.conf` | 远程服务器 `/etc/nginx/nginx.conf` | Nginx 主配置 |
| `remote-devkit.conf` | 远程服务器 `/etc/nginx/conf.d/devkit.conf` | DevKit 站点配置 |
| `frontend-nginx.conf` | 本地代码 | 前端 Docker 部署用 |
| `production-nginx.conf` | 本地 DEPLOY.md | 部署文档中的配置（已过时） |

## ⚠️ 注意

`production-nginx.conf` 是从 DEPLOY.md 提取的，与服务器实际配置有差异：

| 差异项 | DEPLOY.md | 实际服务器 |
|--------|-----------|-----------|
| WebSocket | 无单独配置 | `/api/ws/` 独立 location |
| MinIO 代理 | 无 | `/minio/` 代理到 10.0.50.108:9000 |
| 上传限制 | 默认 | `client_max_body_size 200m` |
| 代理缓冲 | 默认 | 256k-512k 缓冲区 |
| 静态资源缓存 | 无 | 1天缓存 + immutable |
| Gzip | 未配置 | 已开启 |

## 恢复命令

```bash
# 恢复主配置
scp backups/nginx/remote-nginx.conf root@123.57.201.44:/etc/nginx/nginx.conf

# 恢复站点配置
scp backups/nginx/remote-devkit.conf root@123.57.201.44:/etc/nginx/conf.d/devkit.conf

# 重启 nginx
ssh root@123.57.201.44 "nginx -t && systemctl reload nginx"
```
