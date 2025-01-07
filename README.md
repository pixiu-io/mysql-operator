# 部署 Helm 私有仓库指南

## 前置条件

1. 已安装 Docker 并正常运行。
2. 已配置好需要挂载的目录和文件，确保以下路径存在并有正确的权限：
   - `/usr/local/chartrepo/pixiuio`
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

### 3. 上传 `charts`

将 `charts` 上传到 `/usr/local/chartrepo/pixiuio 


### 4. 构建 `index.yaml` 文件

确保以下目录和文件存在：
```bash
helm repo index  /usr/local/chartrepo/pixiuio  --url https://harbor.cloud.pixiuio.com/chartrepo/pixiuio
```
### 5. 启动 Helm 私有仓库容器

```bash
docker run -d --name=helm-repo \
    -p 80:80 \
    -p 443:443 \
    -v /usr/local/chartrepo/pixiuio:/usr/share/nginx/html/chartrepo/pixiuio \
    -v /usr/local/nginx/nginx.conf:/etc/nginx/conf.d/nginx.conf \
    -v /usr/local/nginx/ssl:/etc/nginx/ssl \
    nginx

```

### 6. 验证 Helm 私有仓库

```bash
helm   repo  add  pixiuio  https://harbor.cloud.pixiuio.com/chartrepo/pixiuio
helm search repo pixiuio
```

得到如此回显表明部署成功

```bash
[root@pixiu-server pixiuio]# helm search repo pixiuio
NAME                           	CHART VERSION	APP VERSION	DESCRIPTION
pixiuio-1/gpu-operator         	v24.6.2      	v24.6.2    	NVIDIA GPU Operator creates/configures/manages ...
pixiuio-1/grafana              	7.2.4        	10.2.3     	The leading tool for querying and visualizing t...
pixiuio-1/jenkins              	4.12.0       	2.426.2    	Jenkins - Build great things at any scale! The ...
pixiuio-1/kube-prometheus-stack	65.1.1       	v0.77.1    	kube-prometheus-stack collects Kubernetes manif...
pixiuio-1/kubernetes-dashboard 	6.0.0        	2.7.0      	General-purpose web UI for Kubernetes clusters
pixiuio-1/loki                 	5.41.8       	2.9.3      	Helm chart for Grafana Loki in simple, scalable...
pixiuio-1/loki-stack           	2.10.2       	v2.9.3     	Loki: like Prometheus, but for logs.
pixiuio-1/prometheus           	25.10.0      	v2.49.1    	Prometheus is a monitoring system and time seri...
pixiuio-1/prometheus-adapter   	4.11.0       	v0.12.0    	A Helm chart for k8s prometheus adapter
pixiuio-1/promtail             	6.15.4       	2.9.3      	Promtail is an agent which ships the contents o...
pixiuio-1/xgnginx              	0.1.0        	1.16.0     	A Helm chart for Kubernetes
pixiuio-1/zookeeper            	11.4.9       	3.8.2      	Apache ZooKeeper provides a reliable, centraliz...
```
