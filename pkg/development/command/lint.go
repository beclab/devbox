package command

import (
	"context"
	"path/filepath"

	"k8s.io/klog/v2"

	oac "github.com/beclab/Olares/framework/oac"
)

type lint struct {
	checkChart
}

func Lint() *lint {
	return &lint{*newCheckChart()}
}

func (l *lint) WithDir(dir string) *lint {
	l.baseCommand.withDir(dir)
	return l
}

func (l *lint) Run(ctx context.Context, owner, chart string) error {
	chartPath := filepath.Join(l.baseCommand.dir, owner, chart)

	if err := oac.LintBothOwnerScenarios(chartPath, oac.SkipSameVersionCheck()); err != nil {
		klog.Errorf("failed to lint chart path=%s: %v", chartPath, err)
		return err
	}
	return nil
}
