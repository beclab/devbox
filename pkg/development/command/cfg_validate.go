package command

import (
	"context"
	"path/filepath"

	"github.com/beclab/oachecker"
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

func (c *checkCfg) Run(ctx context.Context, chart string) error {
	err := oachecker.CheckChart(filepath.Join(c.baseCommand.dir, chart))
	return err
}
