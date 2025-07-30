package command

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/beclab/devbox/pkg/development/helm"

	"k8s.io/klog/v2"
)

type copyApp struct {
	baseDir string
}

func CopyApp() *copyApp {
	return &copyApp{baseDir: "/"}
}

func (c *copyApp) WithDir(dir string) *copyApp {
	c.baseDir = dir
	return c
}

func (c *copyApp) Run(src, dstApp string) error {
	if !existDir(src) {
		klog.Error("copy app error, from ", src)
		return errors.New("copy app error")
	}

	files, err := os.ReadDir(src)
	if err != nil {
		klog.Errorf("failed to read dir %s, err=%v", src, err)
		return err
	}

	var realFiles []fs.DirEntry
	for _, f := range files {
		if f.Name() != "." && f.Name() != ".." {
			realFiles = append(realFiles, f)
		}
	}
	if len(realFiles) == 1 {
		if realFiles[0].IsDir() {
			src = filepath.Join(src, realFiles[0].Name())
		}
	}

	srcChart, err := helm.LoadChart(src)
	if err != nil {
		klog.Errorf("failed to load chart from source %s, err=%v", src, err)
		return err
	}

	realPath := filepath.Join(c.baseDir, dstApp)

	if existDir(realPath) {
		dstChart, err := helm.LoadChart(realPath)
		if err != nil {
			klog.Errorf("failed to load dest chart %s, err=%v", realPath, err)
		} else {
			version, err := helm.GetChartVersion(dstChart)
			if err != nil {
				klog.Errorf("failed to get dest chart %s version, err=%v", realPath, err)
			} else {
				err = helm.UpdateChartVersion(srcChart, dstApp, src, version)
				if err != nil {
					klog.Errorf("failed to update chart name=%s,path=%s, to version=%s, err=%v", dstApp, src, version.String(), err)
				}
			}
		}
		err = os.RemoveAll(realPath)
		if err != nil {
			msg := fmt.Sprintf("failed to remove app chart path %s,err=%v", realPath, err)
			return errors.New(msg)
		}
	}

	err = copyDir(src, realPath)
	if err != nil {
		klog.Errorf("failed to copy dir from %s to %s, err=%v", src, realPath, err)
	}
	return err
}
