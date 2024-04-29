package command

import (
	"context"
	"path/filepath"

	"k8s.io/klog/v2"
)

type checkCfg struct {
	checkChart
}

func CheckCfg() *checkCfg {
	return &checkCfg{*newCheckChart()}
}

func (c *checkCfg) WithDir(dir string) *checkCfg {
	c.baseCommand.withDir(dir)
	return c
}

func (c *checkCfg) Run(ctx context.Context, chart string) (string, error) {
	chartPath := c.baseCommand.dir
	chartPath = filepath.Join(chartPath, chart)

	output, err := c.run(ctx, "cfg", "-c", chartPath)
	if err != nil {
		klog.Error("run check-chart cfg error, ", err, ", ", chartPath)
		return "", err
	}

	if len(output) == 0 {
		return "", nil
	}

	return output, nil
}
