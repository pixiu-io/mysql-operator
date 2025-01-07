# 部署 Helm 私有仓库指南

## 前置条件

1. 已安装 Docker 并正常运行。
2. 已配置好需要挂载的目录和文件，确保以下路径存在并有正确的权限：
   - `/usr/local/chartrepo/pixiuio`
   - `/usr/local/charts`
   - `/usr/local/nginx/nginx.conf`
   - `/usr/local/nginx/ssl`

## 操作步骤

### 1. 准备目录和文件

确保以下目录和文件存在：
```bash
# 创建所需目录
mkdir -p /usr/local/chartrepo/pixiuio
mkdir -p /usr/local/charts
mkdir -p /usr/local/nginx/ssl
```

### 2. 准备 nginx 配置文件
vim  /usr/local/nginx/nginx.conf

```bash
server {
    listen 80;
    server_name localhost;

    # 重定向所有 HTTP 请求到 HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name harbor.cloud.pixiuio.com;

    # SSL 配置
    ssl_certificate /etc/nginx/ssl/helm-harbor.pem;
    ssl_certificate_key /etc/nginx/ssl/helm-harbor.key;

    # SSL 强化配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:...';
    ssl_prefer_server_ciphers on;

    # 网站根目录配置
    location / {
        root /usr/share/nginx/html;
        index index.html;
    }
}
```
### 3. 构建 `index.yaml` 文件

确保以下目录和文件存在：
```bash
# 创建所需目录
mkdir -p /usr/local/chartrepo/pixiuio
mkdir -p /usr/local/charts
mkdir -p /usr/local/nginx/ssl
```
