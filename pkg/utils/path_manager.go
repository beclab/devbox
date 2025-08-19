package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

type PathManager struct {
	baseDir string
}

var (
	globalPathMgr *PathManager
)

func NewPathManager(baseDir string) *PathManager {
	return &PathManager{
		baseDir: baseDir,
	}
}

func init() {
	baseDir := os.Getenv("BASE_DIR")
	if baseDir == "" {
		baseDir = "/tmp"
	}
	globalPathMgr = NewPathManager(baseDir)
}

func (pm *PathManager) GetAppPath(username, appName string) string {
	return filepath.Join(pm.baseDir, username, appName)
}

func (pm *PathManager) GetUserBaseDir(username string) string {
	return filepath.Join(pm.baseDir, username)
}

func (pm *PathManager) GetChartmuseumURL(username, endpoint string) string {
	return fmt.Sprintf("http://127.0.0.1:8888/%s%s", username, endpoint)
}

func (pm *PathManager) GetChartVersionsURL(username, chartName string) string {
	return pm.GetChartmuseumURL(username, fmt.Sprintf("/api/charts/%s", chartName))
}

func (pm *PathManager) GetDeleteChartVersionURL(username, chartName, version string) string {
	return pm.GetChartmuseumURL(username, fmt.Sprintf("/api/charts/%s/%s", chartName, version))
}

func GetAppPath(owner, appName string) string {
	return globalPathMgr.GetAppPath(owner, appName)
}

func GetUserBaseDir(username string) string {
	return globalPathMgr.GetUserBaseDir(username)
}
