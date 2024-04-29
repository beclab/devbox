package command

import (
	"context"
	"path/filepath"

	"k8s.io/klog/v2"
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

func (l *lint) Run(ctx context.Context, chart string) (string, error) {
	chartPath := l.baseCommand.dir
	chartPath = filepath.Join(chartPath, chart)

	output, err := l.run(ctx, "-c", chartPath)
	if err != nil {
		klog.Error("run check-chart error, ", err, ", ", chartPath)
		return "", err
	}

	if len(output) == 0 {
		return "", nil
	}

	return output, nil
}
