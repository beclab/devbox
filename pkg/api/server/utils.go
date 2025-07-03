package server

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/utils"
	"github.com/beclab/oachecker"

	"github.com/mholt/archiver/v3"
	"k8s.io/klog/v2"
)

func getAppPath(owner, app string) string {
	return filepath.Join(BaseDir, owner, app)
}

func UnArchive(src, dstDir string) error {
	err := CheckDir(dstDir)
	if err != nil {
		return err
	}
	err = archiver.Unarchive(src, dstDir)
	return err
}

func CheckDir(dirname string) error {
	fi, err := os.Stat(dirname)
	if (err == nil || os.IsExist(err)) && fi.IsDir() {
		return nil
	}
	if os.IsExist(err) {
		return err
	}

	err = os.MkdirAll(dirname, 0755)
	return err
}

func readCfgFromFile(owner, chartDir string) (*oachecker.AppConfiguration, error) {
	cfgFile := findAppCfgFile(chartDir)
	klog.Infof("readCfgFromFile: %s", cfgFile)
	if len(cfgFile) == 0 {
		return nil, errors.New("not found OlaresManifest.yaml file")
	}
	appcfg, err := readAppInfo(owner, cfgFile)
	if err != nil {
		return nil, err
	}
	return appcfg, nil
}

func readAppInfo(owner, cfgFile string) (*oachecker.AppConfiguration, error) {
	f, err := os.Open(cfgFile)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	admin, err := utils.GetAdminUsername(context.TODO())
	if err != nil {
		return nil, err
	}
	opts := []func(map[string]interface{}){
		oachecker.WithAdmin(admin),
		oachecker.WithOwner(owner),
	}
	appcfg, err := oachecker.GetAppConfigurationFromContent(data, opts...)
	return appcfg, nil
}

// findAppCfgFile find app.cfg path in untar path
func findAppCfgFile(chartDirPath string) string {
	charts, err := os.ReadDir(chartDirPath)
	if err != nil {
		return ""
	}

	for _, c := range charts {
		if !c.IsDir() || strings.HasPrefix(c.Name(), ".") {
			continue
		}
		appCfgFullName := path.Join(chartDirPath, c.Name(), constants.AppCfgFileName)
		if PathExists(appCfgFullName) {
			return appCfgFullName
		}
	}

	return ""
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}
	return false
}

func findChartPath(chartDirPath string) string {
	charts, err := os.ReadDir(chartDirPath)
	if err != nil {
		return ""
	}

	for _, c := range charts {
		if !c.IsDir() || strings.HasPrefix(c.Name(), ".") {
			continue
		}
		appCfgFullName := path.Join(chartDirPath, c.Name(), constants.AppCfgFileName)
		if PathExists(appCfgFullName) {
			return path.Join(chartDirPath, c.Name())
		}
	}

	return ""
}

func removeSpecialCharsMap(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '.' || r == '_' || r == '-' {
			return -1
		}
		return r
	}, s)
}
