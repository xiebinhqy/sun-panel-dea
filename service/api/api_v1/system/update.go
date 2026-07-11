package system

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateApi struct{}

// GitHub Release 信息
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	HtmlURL    string `json:"html_url"`
	Prerelease bool   `json:"prerelease"`
	CreatedAt  string `json:"created_at"`
}

// 版本比较结果
type VersionCheckResult struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseNotes   string `json:"releaseNotes"`
	DownloadUrl    string `json:"downloadUrl"`
	IsDocker       bool   `json:"isDocker"`
	IsUpdating     bool   `json:"isUpdating"`
}

const (
	// 更新状态文件路径（用于标记正在更新中）
	updateStatusFile = "conf/.updating"
	// GitHub API 地址（你的仓库）
	githubApiUrl = "https://api.github.com/repos/xiebinhqy/sun-panel-dea/releases/latest"
)

// CheckUpdate 检查是否有新版本
func (a *UpdateApi) CheckUpdate(c *gin.Context) {
	currentVer := cmn.GetSysVersionInfo()

	result := VersionCheckResult{
		CurrentVersion: currentVer.Version,
		IsDocker:       global.ISDOCKER == "docker",
	}

	// 检查是否正在更新中
	if _, err := os.Stat(updateStatusFile); err == nil {
		result.IsUpdating = true
		apiReturn.SuccessData(c, result)
		return
	}

	// 请求 GitHub API 获取最新版本
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", githubApiUrl, nil)
	if err != nil {
		apiReturn.Error(c, fmt.Sprintf("请求GitHub API失败: %v", err))
		return
	}
	req.Header.Set("Accept", "application/json")
	// 使用 User-Agent 避免被 GitHub 限制
	req.Header.Set("User-Agent", "sun-panel-dea-update-checker")

	resp, err := client.Do(req)
	if err != nil {
		apiReturn.Error(c, fmt.Sprintf("连接GitHub失败: %v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		apiReturn.Error(c, fmt.Sprintf("读取GitHub响应失败: %v", err))
		return
	}

	// 处理 API 限流等情况
	if resp.StatusCode == 404 {
		// 没有 Release 发布，视为当前已是最新
		result.LatestVersion = currentVer.Version
		result.ReleaseNotes = ""
		result.HasUpdate = false
		apiReturn.SuccessData(c, result)
		return
	}

	if resp.StatusCode != 200 {
		apiReturn.Error(c, fmt.Sprintf("GitHub API返回错误状态码: %d", resp.StatusCode))
		return
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		apiReturn.Error(c, fmt.Sprintf("解析GitHub响应失败: %v", err))
		return
	}

	// 清理版本号（去掉前缀 v）
	latestVer := strings.TrimPrefix(release.TagName, "v")
	currentVerStr := currentVer.Version

	// 比较版本号
	result.LatestVersion = release.TagName
	result.ReleaseNotes = release.Body
	result.DownloadUrl = release.HtmlURL
	result.HasUpdate = compareVersions(latestVer, currentVerStr) > 0

	apiReturn.SuccessData(c, result)
}

// PerformUpdate 执行在线更新
func (a *UpdateApi) PerformUpdate(c *gin.Context) {
	// 验证是否在 Docker 环境中运行
	if global.ISDOCKER != "docker" {
		apiReturn.Error(c, "非Docker环境不支持在线更新，请手动更新")
		return
	}

	// 检查是否正在更新中
	if _, err := os.Stat(updateStatusFile); err == nil {
		apiReturn.Error(c, "正在更新中，请勿重复操作")
		return
	}

	// 创建更新状态标记文件
	if err := os.WriteFile(updateStatusFile, []byte(time.Now().Format("2006-01-02 15:04:05")), 0644); err != nil {
		apiReturn.Error(c, fmt.Sprintf("创建更新标记失败: %v", err))
		return
	}

	// 在后台执行更新脚本
	go performUpdateAsync()

	apiReturn.SuccessData(c, gin.H{"message": "更新已启动，系统将在更新完成后自动重启"})
}

// GetUpdateStatus 获取更新状态
func (a *UpdateApi) GetUpdateStatus(c *gin.Context) {
	if _, err := os.Stat(updateStatusFile); err == nil {
		apiReturn.SuccessData(c, gin.H{"status": "updating"})
		return
	}
	apiReturn.SuccessData(c, gin.H{"status": "idle"})
}

// performUpdateAsync 后台异步执行更新
func performUpdateAsync() {
	// 延迟执行以确保 API 响应已返回
	time.Sleep(2 * time.Second)

	global.Logger.Info("开始在线更新流程...")

	// 构建更新脚本路径
	updateScriptPath := filepath.Join("conf", "update.sh")
	updateLogPath := filepath.Join("conf", "update.log")

	// 创建更新脚本内容
	scriptContent := `#!/bin/bash
# Sun-Panel-DEA 在线更新脚本
# 由系统自动生成

LOG_FILE="` + updateLogPath + `"

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
`

	// 写入更新脚本
	if err := os.WriteFile(updateScriptPath, []byte(scriptContent), 0755); err != nil {
		global.Logger.Errorf("写入更新脚本失败: %v", err)
		// 清理更新标记
		os.Remove(updateStatusFile)
		return
	}

	// 创建日志文件
	os.WriteFile(updateLogPath, []byte{}, 0644)

	// 检查 docker 是否可用
	dockerCheck := exec.Command("docker", "info")
	if err := dockerCheck.Run(); err != nil {
		global.Logger.Errorf("Docker 不可用: %v", err)
		os.Remove(updateStatusFile)
		os.Remove(updateScriptPath)
		return
	}

	// 使用 nohup 后台执行更新脚本（分离父进程）
	cmd := exec.Command("/bin/bash", updateScriptPath)
	// 设置独立的进程组，使得脚本在父进程结束后继续运行
	cmd.SysProcAttr = getSysProcAttr()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		global.Logger.Errorf("启动更新脚本失败: %v", err)
		os.Remove(updateStatusFile)
		return
	}

	// 分离子进程，不等待其完成
	go func() {
		// 等待一段时间后清理更新标记（实际上下面的脚本会重启容器，这个标记文件会被销毁）
		time.Sleep(30 * time.Second)
		// 如果脚本执行完毕但容器还在（意味着更新失败），清理标记
		if err := cmd.Wait(); err != nil {
			global.Logger.Errorf("更新脚本执行异常: %v", err)
			// 检查容器是否还在运行
			checkCmd := exec.Command("docker", "ps", "--filter", "name=sun-panel-dea", "--format", "{{.Names}}")
			if output, err := checkCmd.Output(); err != nil || !strings.Contains(string(output), "sun-panel-dea") {
				// 容器不在运行，说明更新部分完成但有问题
				global.Logger.Errorf("更新后容器未正常运行")
			}
		}
		os.Remove(updateStatusFile)
		os.Remove(updateScriptPath)
	}()

	global.Logger.Info("更新脚本已在后台启动")
}

// compareVersions 比较两个版本号
// 返回: 1 (v1 > v2), 0 (v1 == v2), -1 (v1 < v2)
func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(v1Parts) {
			fmt.Sscanf(v1Parts[i], "%d", &num1)
		}
		if i < len(v2Parts) {
			fmt.Sscanf(v2Parts[i], "%d", &num2)
		}
		if num1 > num2 {
			return 1
		} else if num1 < num2 {
			return -1
		}
	}
	return 0
}

// getSysProcAttr 获取系统进程属性
func getSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}
