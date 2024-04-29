package command

import (
	"errors"
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
		klog.Error("load chart from source error, ", err, ", ", src)
		return err
	}

	realPath := filepath.Join(c.baseDir, dstApp)

	if existDir(realPath) {
		dstChart, err := helm.LoadChart(realPath)
		if err != nil {
			klog.Error("load dest chart error, ", err)
		} else {
			version, err := helm.GetChartVersion(dstChart)
			if err != nil {
				klog.Error("get dest chart version error, ", err)
			} else {
				err = helm.UpdateChartVersion(srcChart, dstApp, src, version)
				if err != nil {
					klog.Error("update source chart error, ", err)
				}
			}
		}
		err = os.RemoveAll(realPath)
		if err != nil {
			klog.Error("remove app chart path error, ", err, ", ", realPath)
			return err
		}
	}

	err = copyDir(src, realPath)
	if err != nil {
		klog.Error("copy dir error, ", err, ", from ", src, " to ", realPath)
	}
	return err
}
