package command

import (
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

type deleteChart struct {
	baseDir string
}

func DeleteChart() *deleteChart {
	return &deleteChart{baseDir: "/"}
}

func (c *deleteChart) WithDir(dir string) *deleteChart {
	c.baseDir = dir
	return c
}

func (c *deleteChart) Run(pathToPackage string) error {
	realPath := filepath.Join(c.baseDir, pathToPackage)

	err := os.RemoveAll(realPath)
	if err != nil {
		klog.Error("remove chart dir error, ", err, ", ", realPath)
	}

	return err
}
