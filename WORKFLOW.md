# Sun-Panel-Dea 开发工作流文档

## 概述

本文档描述了 sun-panel-dea 项目的开发、构建和部署工作流。

## 仓库信息

- **SSH**: git@github.com:xiebinhqy/sun-panel-dea.git
- **HTTPS**: https://github.com/xiebinhqy/sun-panel-dea.git
- **Docker 镜像**: xiebinhqy/sun-panel-dea:latest

---

## 完整工作流

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  本地开发   │───▶│  Push 到    │───▶│ GitHub      │───▶│ Docker      │───▶│ 在线更新    │
│  修改代码   │    │  GitHub     │    │ Actions     │    │ Hub 拉取    │    │ 点击更新    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

---

## 步骤 1: 本地开发

在本地修改代码，开发新功能或修复 bug。

```bash
# 查看修改状态
git status

# 查看具体修改
git diff

# 本地测试（如需要）
npm run dev
```

---

## 步骤 2: 提交并推送代码到 GitHub

```bash
# 添加修改的文件
git add .

# 提交修改
git commit -m "feat: 添加新功能描述"

# 推送到 GitHub
git push origin main
```

推送后，GitHub Actions 会自动触发 Docker 镜像构建。

---

## 步骤 3: GitHub Actions 自动构建

GitHub Actions 会在以下情况触发构建：
- 推送到 `main` 或 `master` 分支
- 发布新的 Release

### 配置要求

在 GitHub 仓库设置中添加以下 Secrets：
- `DOCKERHUB_USERNAME`: Docker Hub 用户名
- `DOCKERHUB_TOKEN`: Docker Hub 访问令牌（从 Docker Hub → Account Settings → Security 获取）

### 查看构建状态

```bash
# 方法1: 查看 GitHub Actions 页面
# 访问: https://github.com/xiebinhqy/sun-panel-dea/actions

# 方法2: 使用 gh CLI
gh run list
gh run watch
```

构建成功后，镜像会推送到 `xiebinhqy/sun-panel-dea:latest`

---

## 步骤 4: 在线更新

### Docker 环境（推荐）

1. **检测更新**
   - 在网页中点击"检查更新"按钮
   - 系统会连接 GitHub API 检查最新版本

2. **执行更新**
   - 如果有新版本，点击"在线更新"
   - 系统会自动：
     1. 拉取最新 Docker 镜像
     2. 停止当前容器
     3. 移除旧容器
     4. 使用 docker-compose 重新启动

### 传统环境

系统会下载 GitHub Release 中的压缩包，解压并替换程序文件。

---

## 配置文件

### docker-compose.yml

```yaml
version: '3.2'

services:
  sun-panel:
    image: 'sun-panel-dea:latest'
    container_name: sun-panel
    volumes:
      - ./conf:/app/conf
      - ./uploads:/app/uploads
      - ./database:/app/database
    ports:
      - 3002:3002
    restart: always
```

### GitHub Actions (.github/workflows/docker-build.yml)

```yaml
name: Build and Push Docker Image

on:
  push:
    branches: [main, master]
  release:
    types: [published]

env:
  IMAGE_NAME: xiebinhqy/sun-panel-dea

jobs:
  build:
    runs-on: ubuntu-latest
    # ... (完整配置见文件)
```

---

## 注意事项

1. **Docker Socket 挂载**: 如果需要在容器内执行 Docker 命令，需要挂载 Docker socket：
   ```yaml
   volumes:
     - /var/run/docker.sock:/var/run/docker.sock
   ```

2. **环境变量**: Docker 环境通过以下方式识别：
   - `ISDOCKER` 环境变量
   - `SUN_PANEL_DOCKER` 环境变量
   - `/.dockerenv` 文件存在

3. **数据持久化**: 配置文件、数据库、上传文件通过 volume 挂载，更新不会影响数据。

---

## 常见问题

### Q: GitHub Actions 构建失败怎么办？
A: 检查 Actions 日志，常见问题包括：
- Docker Hub 凭据错误
- 构建资源不足
- 代码编译错误

### Q: 在线更新失败怎么办？
A: 检查容器日志：
```bash
docker logs sun-panel
```

### Q: 如何回滚版本？
A: 使用之前的镜像标签：
```bash
docker pull xiebinhqy/sun-panel-dea:<版本号>
docker-compose down
# 修改 docker-compose.yml 中的 image 为指定版本
docker-compose up -d
```

---

## 更新日志

| 日期 | 变更 |
|------|------|
| 2026-07-09 | 初始版本，建立完整工作流 |