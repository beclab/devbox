package command

import (
	"context"
	"path/filepath"

	oac "github.com/beclab/Olares/framework/oac"
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

func (c *checkCfg) Run(ctx context.Context, owner, chart string) error {
	chartPath := filepath.Join(c.baseCommand.dir, owner, chart)
	return oac.Lint(chartPath)
}
