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

# 准备 nginx 配置文件
vim  /usr/local/nginx/nginx.conf

```bash

```

