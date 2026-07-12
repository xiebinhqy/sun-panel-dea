#!/bin/bash
# Sun-Panel-DEA 在线更新脚本
# 由系统自动生成

LOG_FILE="conf/update.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

log "=== 开始在线更新 ==="

# 1. 获取当前容器信息
log "获取当前容器信息..."
CURRENT_CONTAINER=$(hostname)
log "当前容器ID: $CURRENT_CONTAINER"

# 2. 拉取最新镜像
log "拉取最新镜像 deattorsss/sun-panel-dea:latest ..."
docker pull deattorsss/sun-panel-dea:latest >> "$LOG_FILE" 2>&1
if [ $? -ne 0 ]; then
    log "拉取镜像失败！"
    exit 1
fi
log "镜像拉取成功"

# 3. 获取当前容器的挂载卷和端口配置
log "获取当前容器配置..."
# 获取容器信息
CONTAINER_INFO=$(docker inspect "$CURRENT_CONTAINER" 2>/dev/null)

# 提取挂载卷信息 (排除 docker.sock)
VOLUMES=$(echo "$CONTAINER_INFO" | python3 -c "
import json, sys
info = json.load(sys.stdin)
mounts = info[0]['Mounts']
volumes = []
for m in mounts:
    if m['Type'] == 'bind' and 'docker.sock' not in m['Source']:
        volumes.append('-v ' + m['Source'] + ':' + m['Destination'])
    elif m['Type'] == 'volume':
        volumes.append('-v ' + m['Name'] + ':' + m['Destination'])
print(' '.join(volumes))
" 2>/dev/null)

# 提取端口映射
PORTS=$(echo "$CONTAINER_INFO" | python3 -c "
import json, sys
info = json.load(sys.stdin)
host_config = info[0]['HostConfig']
port_bindings = host_config.get('PortBindings', {})
ports = []
for container_port, bindings in port_bindings.items():
    for b in bindings:
        host_port = b.get('HostPort', '')
        ports.append('-p ' + host_port + ':' + container_port.split('/')[0])
print(' '.join(ports))
" 2>/dev/null)

# 提取网络模式
NETWORK_MODE=$(echo "$CONTAINER_INFO" | python3 -c "
import json, sys
info = json.load(sys.stdin)
print(info[0]['HostConfig']['NetworkMode'])
" 2>/dev/null)

# 提取重启策略
RESTART_POLICY=$(echo "$CONTAINER_INFO" | python3 -c "
import json, sys
info = json.load(sys.stdin)
print(info[0]['HostConfig']['RestartPolicy']['Name'])
" 2>/dev/null)

log "VOLUMES: $VOLUMES"
log "PORTS: $PORTS"
log "NETWORK_MODE: $NETWORK_MODE"
log "RESTART_POLICY: $RESTART_POLICY"

# 4. 优雅停止旧容器
log "停止旧容器..."
docker stop "$CURRENT_CONTAINER" >> "$LOG_FILE" 2>&1
docker rm "$CURRENT_CONTAINER" >> "$LOG_FILE" 2>&1
log "旧容器已移除"

# 5. 启动新容器
log "启动新容器..."
RESTART_FLAG=""
if [ "$RESTART_POLICY" = "always" ] || [ "$RESTART_POLICY" = "unless-stopped" ]; then
    RESTART_FLAG="--restart $RESTART_POLICY"
fi

NET_FLAG=""
if [ "$NETWORK_MODE" != "default" ] && [ -n "$NETWORK_MODE" ]; then
    NET_FLAG="--network $NETWORK_MODE"
fi

# 构建启动命令
RUN_CMD="docker run -d \
    --name sun-panel-dea \
    $RESTART_FLAG \
    $NET_FLAG \
    $VOLUMES \
    $PORTS \
    deattorsss/sun-panel-dea:latest"

log "执行命令: $RUN_CMD"
eval "$RUN_CMD" >> "$LOG_FILE" 2>&1

if [ $? -ne 0 ]; then
    log "启动新容器失败！"
    exit 1
fi

log "=== 在线更新完成 ==="
log "新容器已启动成功！"
