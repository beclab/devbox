package command

import (
	"context"
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

func (l *lint) Run(ctx context.Context, chart string) error {
	chartPath := filepath.Join(l.baseCommand.dir, chart)
	err := oachecker.LintWithDifferentOwnerAdmin(chartPath, "owner", "admin")
	if err != nil {
		return err
	}
	err = oachecker.LintWithSameOwnerAdmin(chartPath, "owner")
	if err != nil {
		return err
	}
	return err
}
