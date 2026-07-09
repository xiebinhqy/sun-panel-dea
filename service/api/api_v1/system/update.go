package system

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/lib/cmn"

	"github.com/gin-gonic/gin"
)

type UpdateApi struct{}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	} `json:"assets"`
}

type UpdateCheckResult struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseUrl     string `json:"releaseUrl"`
	DownloadUrl    string `json:"downloadUrl"`
	ReleaseNotes   string `json:"releaseNotes"`
}

const (
	GITHUB_OWNER = "xiebinhqy"
	GITHUB_REPO  = "sun-panel-dea"
	DOCKER_IMAGE = "xiebinhqy/sun-panel-dea:latest"
)

// 检查是否运行在 Docker 环境中
func isDockerEnv() bool {
	// 检查环境变量
	if os.Getenv("ISDOCKER") != "" || os.Getenv("SUN_PANEL_DOCKER") != "" {
		return true
	}
	// 检查 Docker 环境标识文件
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

// 检查GitHub最新版本
func (a *UpdateApi) CheckUpdate(c *gin.Context) {
	currentVersion := cmn.GetSysVersionInfo()

	// 请求GitHub Releases API
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GITHUB_OWNER, GITHUB_REPO)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		apiReturn.Error(c, "Failed to create request: "+err.Error())
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Sun-Panel/"+currentVersion.Version)

	resp, err := client.Do(req)
	if err != nil {
		apiReturn.Error(c, "Failed to check update: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		apiReturn.Error(c, "Failed to read response: "+err.Error())
		return
	}

	// GitHub API rate limit check
	if resp.StatusCode == 403 {
		apiReturn.Error(c, "GitHub API rate limit exceeded, please try again later")
		return
	}

	if resp.StatusCode != 200 {
		apiReturn.Error(c, fmt.Sprintf("GitHub API returned status %d", resp.StatusCode))
		return
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		apiReturn.Error(c, "Failed to parse GitHub response: "+err.Error())
		return
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersionStr := strings.TrimPrefix(currentVersion.Version, "v")

	// 版本比较
	hasUpdate := compareVersions(latestVersion, currentVersionStr)

	// 查找下载URL (优先找linux-amd64或musl版本)
	downloadUrl := ""
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if (strings.Contains(name, "linux") || strings.Contains(name, "musl")) &&
			(strings.Contains(name, "amd64") || strings.Contains(name, "x86_64")) &&
			(strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip")) {
			downloadUrl = asset.BrowserDownloadUrl
			break
		}
	}
	// 如果没找到特定平台，使用第一个可下载的压缩包
	if downloadUrl == "" {
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip") {
				downloadUrl = asset.BrowserDownloadUrl
				break
			}
		}
	}

	result := UpdateCheckResult{
		HasUpdate:      hasUpdate,
		CurrentVersion: currentVersion.Version,
		LatestVersion:  release.TagName,
		ReleaseUrl:     fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", GITHUB_OWNER, GITHUB_REPO, release.TagName),
		DownloadUrl:    downloadUrl,
		ReleaseNotes:   release.Body,
	}

	apiReturn.SuccessData(c, result)
}

// 执行更新
func (a *UpdateApi) PerformUpdate(c *gin.Context) {
	// Docker 环境使用 Docker 更新逻辑
	if isDockerEnv() {
		a.performDockerUpdate(c)
		return
	}

	// 非 Docker 环境使用传统更新逻辑
	currentVersion := cmn.GetSysVersionInfo()

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GITHUB_OWNER, GITHUB_REPO)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		apiReturn.Error(c, "Failed to create request: "+err.Error())
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Sun-Panel/"+currentVersion.Version)

	resp, err := client.Do(req)
	if err != nil {
		apiReturn.Error(c, "Failed to check update: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		apiReturn.Error(c, "Failed to read response: "+err.Error())
		return
	}

	if resp.StatusCode == 403 {
		apiReturn.Error(c, "GitHub API rate limit exceeded, please try again later")
		return
	}

	if resp.StatusCode != 200 {
		apiReturn.Error(c, fmt.Sprintf("GitHub API returned status %d", resp.StatusCode))
		return
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		apiReturn.Error(c, "Failed to parse GitHub response: "+err.Error())
		return
	}

	// 查找下载URL
	downloadUrl := ""
	assetName := ""
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if (strings.Contains(name, "linux") || strings.Contains(name, "musl")) &&
			(strings.Contains(name, "amd64") || strings.Contains(name, "x86_64")) &&
			(strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip")) {
			downloadUrl = asset.BrowserDownloadUrl
			assetName = asset.Name
			break
		}
	}
	if downloadUrl == "" {
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip") {
				downloadUrl = asset.BrowserDownloadUrl
				assetName = asset.Name
				break
			}
		}
	}

	if downloadUrl == "" {
		apiReturn.Error(c, "No suitable download asset found in the latest release")
		return
	}

	// 创建临时目录
	tmpDir := filepath.Join(os.TempDir(), "sun-panel-update")
	os.RemoveAll(tmpDir)
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// 下载文件
	apiReturn.SuccessData(c, gin.H{
		"message":     "Update started, downloading...",
		"downloadUrl": downloadUrl,
	})

	// 异步执行更新（在返回响应后执行）
	go func() {
		downloadPath := filepath.Join(tmpDir, assetName)
		if err := downloadFile(downloadPath, downloadUrl); err != nil {
			fmt.Println("Download failed:", err.Error())
			return
		}

		// 解压文件
		extractDir := filepath.Join(tmpDir, "extract")
		os.MkdirAll(extractDir, 0755)

		if strings.HasSuffix(assetName, ".tar.gz") {
			cmd := exec.Command("tar", "-xzf", downloadPath, "-C", extractDir)
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Println("Extract failed:", err.Error(), string(output))
				return
			}
		} else if strings.HasSuffix(assetName, ".zip") {
			cmd := exec.Command("unzip", "-o", downloadPath, "-d", extractDir)
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Println("Extract failed:", err.Error(), string(output))
				return
			}
		}

		// 查找解压后的sun-panel二进制文件和web目录
		var binaryPath string
		var webPath string

		filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.Name() == "sun-panel" {
				binaryPath = path
			}
			if info.IsDir() && info.Name() == "web" {
				webPath = path
			}
			return nil
		})

		// 获取当前程序路径
		execPath, err := os.Executable()
		if err != nil {
			fmt.Println("Cannot get executable path:", err.Error())
			return
		}

		// 获取当前工作目录
		workDir, _ := os.Getwd()

		// 替换binary
		if binaryPath != "" {
			// 备份旧binary
			backupPath := execPath + ".bak"
			os.Rename(execPath, backupPath)

			// 复制新binary
			if err := copyFile(execPath, binaryPath); err != nil {
				fmt.Println("Failed to copy binary:", err.Error())
				// 恢复备份
				os.Rename(backupPath, execPath)
				return
			}
			os.Chmod(execPath, 0755)
			// 删除备份
			os.Remove(backupPath)

			// 更新版本文件
			versionStr := fmt.Sprintf("%d|%s", 0, strings.TrimPrefix(release.TagName, "v"))
			versionFilePath := filepath.Join(workDir, "service", "assets", "version")
			os.WriteFile(versionFilePath, []byte(versionStr), 0644)
		}

		// 替换web目录
		if webPath != "" {
			webDir := filepath.Join(workDir, "web")
			os.RemoveAll(webDir)
			copyDir(webDir, webPath)
		}

		fmt.Println("Update completed, restarting...")

		// 如果是Docker环境，退出容器让Docker自动重启
		if os.Getenv("ISDOCKER") != "" || os.Getenv("SUN_PANEL_DOCKER") != "" {
			os.Exit(0)
		}

		// 重新启动自身
		cmd := exec.Command(execPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
		os.Exit(0)
	}()

	// 立即返回响应，不阻塞
	// 注意：由于gin的异步特性，这里实际上还是会等goroutine启动后才返回
	// 但更新操作是异步的，所以响应会先返回给前端
}

// Docker 环境下的更新逻辑
func (a *UpdateApi) performDockerUpdate(c *gin.Context) {
	fmt.Println("Docker environment detected, pulling latest image...")

	// 返回响应，然后异步拉取镜像
	apiReturn.SuccessData(c, gin.H{
		"message": "Docker update started, pulling latest image...",
		"image":   DOCKER_IMAGE,
	})

	// 异步执行 Docker 更新
	go func() {
		fmt.Println("Step 1: Pulling latest Docker image:", DOCKER_IMAGE)

		// 拉取最新镜像
		pullCmd := exec.Command("docker", "pull", DOCKER_IMAGE)
		if output, err := pullCmd.CombinedOutput(); err != nil {
			fmt.Println("Failed to pull Docker image:", err.Error(), string(output))
			return
		}
		fmt.Println("Successfully pulled Docker image")

		// 停止当前容器 (sun-panel)
		fmt.Println("Step 2: Stopping current container...")
		stopCmd := exec.Command("docker", "stop", "sun-panel")
		if output, err := stopCmd.CombinedOutput(); err != nil {
			fmt.Println("Warning: Failed to stop container:", err.Error(), string(output))
		}

		// 移除当前容器
		fmt.Println("Step 3: Removing old container...")
		rmCmd := exec.Command("docker", "rm", "sun-panel")
		if output, err := rmCmd.CombinedOutput(); err != nil {
			fmt.Println("Warning: Failed to remove container:", err.Error(), string(output))
		}

		// 使用 docker-compose 重新启动
		fmt.Println("Step 4: Restarting with docker-compose...")
		// 寻找 docker-compose.yml 文件
		composePath := findDockerComposeFile()
		if composePath != "" {
			// 使用 docker-compose 启动
			upCmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")
			if output, err := upCmd.CombinedOutput(); err != nil {
				fmt.Println("Failed to restart with docker-compose:", err.Error(), string(output))
				// 尝试 docker compose (新命令格式)
				upCmd2 := exec.Command("docker", "compose", "-f", composePath, "up", "-d")
				if output2, err2 := upCmd2.CombinedOutput(); err2 != nil {
					fmt.Println("Failed to restart with docker compose:", err2.Error(), string(output2))
				}
			}
		} else {
			fmt.Println("docker-compose.yml not found, container needs manual restart")
		}

		fmt.Println("Docker update completed!")
	}()
}

// 查找 docker-compose.yml 文件
func findDockerComposeFile() string {
	// 常见路径
	paths := []string{
		"./docker-compose.yml",
		"./docker-compose.yaml",
		"/app/docker-compose.yml",
		"/app/docker-compose.yaml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// 简单的版本比较函数
// 格式: "1.2.3" 或 "v1.2.3"
func compareVersions(v1, v2 string) bool {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			num1, _ = strconv.Atoi(strings.TrimSpace(parts1[i]))
		}
		if i < len(parts2) {
			num2, _ = strconv.Atoi(strings.TrimSpace(parts2[i]))
		}
		if num1 > num2 {
			return true
		} else if num1 < num2 {
			return false
		}
	}
	return false
}

// 下载文件
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// 复制文件
func copyFile(dst, src string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// 复制目录
func copyDir(dst string, src string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(dstPath, srcPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(dstPath, srcPath); err != nil {
				return err
			}
		}
	}

	return nil
}
