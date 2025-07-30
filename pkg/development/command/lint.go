package command

import (
	"context"
	"k8s.io/klog/v2"
	"path/filepath"

	"github.com/beclab/oachecker"
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
	err := oachecker.LintWithDifferentOwnerAdmin(chartPath, "owner", "admin")
	if err != nil {
		klog.Errorf("failed to lint chart path=%s with different owner and admin %v", chartPath, err)
		return err
	}
	err = oachecker.LintWithSameOwnerAdmin(chartPath, "owner")
	if err != nil {
		klog.Errorf("failed to lint chart path=%s with same owner and admin %v", chartPath, err)
		return err
	}
	return err
}
