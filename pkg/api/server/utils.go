package server

import (
	"errors"
	"io"
	"k8s.io/klog/v2"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/development/application"

	"github.com/mholt/archiver/v3"
	"gopkg.in/yaml.v3"
)

func getAppPath(app string) string {
	return filepath.Join(BaseDir, app)
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

func readCfgFromFile(chartDir string) (*application.AppConfiguration, error) {
	cfgFile := findAppCfgFile(chartDir)
	klog.Infof("readCfgFromFile: %s", cfgFile)
	if len(cfgFile) == 0 {
		return nil, errors.New("not found TerminusManifest.yaml file")
	}
	appcfg, err := readAppInfo(cfgFile)
	if err != nil {
		return nil, err
	}
	return appcfg, nil
}

func readAppInfo(cfgFile string) (*application.AppConfiguration, error) {
	var appcfg application.AppConfiguration
	f, err := os.Open(cfgFile)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, &appcfg); err != nil {
		return nil, err
	}
	return &appcfg, nil
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
