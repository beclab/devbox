package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/development/helm"

	"k8s.io/klog/v2"
)

type updateRepo struct {
	baseCommand
}

func UpdateRepo() *updateRepo {
	return &updateRepo{*newBaseCommand()}
}

func (c *updateRepo) WithDir(dir string) *updateRepo {
	c.baseCommand.withDir(dir)
	return c
}

func (c *updateRepo) Run(ctx context.Context, owner, app string, notExist bool) (string, error) {
	if app == "" {
		return "", errors.New("repo path must be specified")
	}
	realPath := filepath.Join(c.baseCommand.dir, owner, app)

	chart, err := helm.LoadChart(realPath)
	if err != nil {
		klog.Errorf("failed to load chart path=%s to repo,err=%v", realPath, err)
		return "", err
	}

	klog.Infof("start to upgrade chart version app=%s", app)
	version, err := helm.GetChartVersion(chart)
	if err != nil {
		klog.Errorf("failed to get app=%s chart version %v", app, err)
		return "", err
	}
	newVersion := version.IncPatch()
	uploadChartVersion := version.String()
	if !notExist {
		uploadChartVersion = newVersion.String()
		klog.Infof("uploadChartVersion to %s", uploadChartVersion)
		err = helm.UpgradeChartVersion(chart, app, realPath, &newVersion)
		if err != nil {
			klog.Errorf("failed to upgrade chart version,app=%s,version=%s,err=%v", app, uploadChartVersion, err)
			return "", err
		}
	}

	backupAndRestoreFile := func(orig, bak string) (func(), error) {
		klog.Infof("backup origin path=%s", orig)
		data, err := os.ReadFile(orig)
		if err != nil {
			klog.Errorf("failed to read origin file path=%s,err=%v", orig, err)
			return nil, err
		}

		err = os.WriteFile(bak, data, 0644)
		if err != nil {
			klog.Errorf("failed to backup origin file %s,err=%v", bak, err)
			return nil, err
		}

		return func() {
			klog.Infof("restore file path=%s", orig)
			err = os.Remove(orig)
			if err != nil {
				klog.Errorf("failed to remove file path=%s", orig)
				return
			}

			err = os.Rename(bak, orig)
			if err != nil {
				klog.Errorf("failed to rename from path=%s to path=%s", bak, orig)
			}

		}, nil
	}

	chartYaml := filepath.Join(realPath, "Chart.yaml")
	chartYamlBak := filepath.Join(realPath, "Chart.bak")
	chartDeferFunc, err := backupAndRestoreFile(chartYaml, chartYamlBak)
	if err != nil {
		klog.Errorf("failed to get chartDeferFunc %v", err)
		return "", err
	}
	defer chartDeferFunc()

	err = helm.UpdateChartName(chart, app, realPath)
	if err != nil {
		klog.Errorf("failed to update chart app=%s,path=%s,err=%v", app, realPath, err)
		return "", err
	}

	if !notExist {
		err = helm.UpdateAppCfgVersion(owner, realPath, &newVersion)
		if err != nil {
			klog.Errorf("failed to update OlaresManifest.yaml metadata.version %v", err)
			return "", err
		}
	}

	appcfg := filepath.Join(realPath, constants.AppCfgFileName)
	appcfgBak := filepath.Join(realPath, "OlaresManifest.yaml.bak")
	appcfgDeferFunc, err := backupAndRestoreFile(appcfg, appcfgBak)
	if err != nil {
		klog.Errorf("failed to get appcfg defer func %v", err)
		return "", err
	}
	defer appcfgDeferFunc()

	err = helm.UpdateAppCfgName(owner, app, realPath)
	if err != nil {
		klog.Errorf("failed to update app cfg name app=%s,path=%s,err=%v", app, realPath, err)
		return "", err
	}

	output, err := c.baseCommand.run(ctx, "helm", "cm-push", "-f", fmt.Sprintf("--context-path=%s", owner), owner+"/"+app, "http://localhost:8888", "--debug")
	if err != nil {
		if len(output) > 0 {
			return "", errors.New(output)
		}
		return "", err
	}
	result := strings.Split(output, "\n")
	if len(result) > 0 && result[len(result)-2] == "Done." {
		if !notExist {
			err = deleteOldTgz(owner, app+"-dev", newVersion.String())
			if err != nil {
				klog.Errorf("failed to delete chart repo old tgz %v", err)
			}
		}

	}
	klog.Infof("update repo app %s, newVersion: %s", app, uploadChartVersion)
	return uploadChartVersion, nil

}

func deleteOldTgz(owner, name, notDeleteVersion string) error {
	chartVersions, err := getChartVersions(owner, name)
	if err != nil {
		return err
	}
	errs := make([]error, 0, len(chartVersions)-1)
	for _, cv := range chartVersions {
		if cv.Version == notDeleteVersion {
			continue
		}
		err = deleteChartVersion(owner, name, cv.Version)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return AggregateErrs(errs)
}
