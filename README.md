部署私有 Helm 仓库
本文介绍如何使用 Docker 和 Nginx 部署一个私有的 Helm 仓库。私有 Helm 仓库可以用于存储和管理自定义的 Helm Chart，方便团队内部使用。

前提条件
已安装 Docker 和 Docker Compose。

已安装 Helm 客户端。

服务器上已配置 SSL 证书（可选，推荐）。

步骤 1：准备目录结构
在服务器上创建以下目录结构：

bash
复制
mkdir -p /usr/local/chartrepo/pixiuio
mkdir -p /usr/local/charts
mkdir -p /usr/local/nginx/ssl
mkdir -p /usr/local/nginx/conf.d
/usr/local/chartrepo/pixiuio：用于存储 Helm Chart 文件。

/usr/local/charts：用于存储 Helm Chart 的索引文件。

/usr/local/nginx/ssl：用于存储 SSL 证书（可选）。

/usr/local/nginx/conf.d：用于存储 Nginx 配置文件。

步骤 2：配置 Nginx
在 /usr/local/nginx/conf.d/nginx.conf 中创建 Nginx 配置文件：

nginx
复制
server {
    listen 80;
    server_name helm-repo.example.com;  # 替换为你的域名

    location /chartrepo/pixiuio {
        alias /usr/share/nginx/html/chartrepo/pixiuio;
        autoindex on;
    }

    location /charts {
        alias /usr/share/nginx/html/charts;
        autoindex on;
    }
}

server {
    listen 443 ssl;
    server_name helm-repo.example.com;  # 替换为你的域名

    ssl_certificate /etc/nginx/ssl/fullchain.pem;  # SSL 证书路径
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;  # SSL 私钥路径

    location /chartrepo/pixiuio {
        alias /usr/share/nginx/html/chartrepo/pixiuio;
        autoindex on;
    }

    location /charts {
        alias /usr/share/nginx/html/charts;
        autoindex on;
    }
}
将 helm-repo.example.com 替换为你的域名。

如果需要 HTTPS，将 SSL 证书和私钥放入 /usr/local/nginx/ssl 目录。

步骤 3：启动 Nginx 容器
使用以下命令启动 Nginx 容器：

bash
复制
docker run -d \
  --name=helm-repo-new-3 \
  -p 80:80 \
  -p 443:443 \
  -v /usr/local/chartrepo/pixiuio:/usr/share/nginx/html/chartrepo/pixiuio \
  -v /usr/local/charts:/usr/share/nginx/html/charts \
  -v /usr/local/nginx/conf.d/nginx.conf:/etc/nginx/conf.d/nginx.conf \
  -v /usr/local/nginx/ssl:/etc/nginx/ssl \
  nginx
-p 80:80 和 -p 443:443：将容器的 80 和 443 端口映射到主机。

-v：挂载本地目录到容器中。

步骤 4：生成 Helm Chart 索引
将 Helm Chart 文件放入 /usr/local/chartrepo/pixiuio 目录后，生成索引文件：

bash
复制
cd /usr/local/chartrepo/pixiuio
helm repo index . --url https://helm-repo.example.com/chartrepo/pixiuio
--url：指定 Helm 仓库的访问地址。

生成的 index.yaml 文件将用于 Helm 客户端查找 Chart。

步骤 5：添加 Helm 仓库
在 Helm 客户端中添加私有仓库：

bash
复制
helm repo add pixiuio https://helm-repo.example.com/chartrepo/pixiuio
helm repo update
pixiuio：仓库名称，可自定义。

https://helm-repo.example.com/chartrepo/pixiuio：仓库地址。

步骤 6：验证 Helm 仓库
使用以下命令验证仓库是否正常工作：

bash
复制
helm search repo pixiuio
如果能看到上传的 Chart，说明仓库部署成功。
