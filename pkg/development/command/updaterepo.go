package command

import (
	"context"
	"errors"
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

func (c *updateRepo) Run(ctx context.Context, app string, notExist bool) (string, error) {
	if app == "" {
		return "", errors.New("repo path must be specified")
	}
	realPath := filepath.Join(c.baseCommand.dir, app)

	chart, err := helm.LoadChart(realPath)
	if err != nil {
		klog.Error("load chart to upgrade repo error, ", err, ", ", realPath)
		return "", err
	}

	klog.Info("upgrade chart version, ", app)
	version, err := helm.GetChartVersion(chart)
	if err != nil {
		return "", err
	}
	newVersion := version.IncPatch()
	uploadChartVersion := version.String()
	if !notExist {
		uploadChartVersion = newVersion.String()
		klog.Infof("uploadChartVersion: %s", uploadChartVersion)
		err = helm.UpgradeChartVersion(chart, app, realPath, &newVersion)
		if err != nil {
			klog.Error("upgrade chart version error, ", err)
			return "", err
		}
	}

	backupAndRestoreFile := func(orig, bak string) (func(), error) {
		klog.Info("backup ", orig)
		data, err := os.ReadFile(orig)
		if err != nil {
			klog.Error("read origin file error, ", err, ", ", orig)
			return nil, err
		}

		err = os.WriteFile(bak, data, 0644)
		if err != nil {
			klog.Error("backup origin file error, ", err, ", ", bak)
			return nil, err
		}

		return func() {
			klog.Info("restore ", orig)
			err = os.Remove(orig)
			if err != nil {
				klog.Error(err)
				return
			}

			err = os.Rename(bak, orig)
			if err != nil {
				klog.Error(err)
			}

		}, nil
	}

	chartYaml := filepath.Join(realPath, "Chart.yaml")
	chartYamlBak := filepath.Join(realPath, "Chart.bak")
	chartDeferFunc, err := backupAndRestoreFile(chartYaml, chartYamlBak)
	if err != nil {
		return "", err
	}
	defer chartDeferFunc()

	err = helm.UpdateChartName(chart, app, realPath)
	if err != nil {
		klog.Error("update chart name error, ", err)
		return "", err
	}

	if !notExist {
		err = helm.UpdateAppCfgVersion(realPath, &newVersion)
		if err != nil {
			klog.Error("update OlaresManifest.yaml metadata.version error, ", err)
			return "", err
		}
	}

	appcfg := filepath.Join(realPath, constants.AppCfgFileName)
	appcfgBak := filepath.Join(realPath, "OlaresManifest.yaml.bak")
	appcfgDeferFunc, err := backupAndRestoreFile(appcfg, appcfgBak)
	if err != nil {
		return "", err
	}
	defer appcfgDeferFunc()

	err = helm.UpdateAppCfgName(app, realPath)
	if err != nil {
		return "", err
	}

	output, err := c.baseCommand.run(ctx, "helm", "cm-push", "-f", app, "http://localhost:8888", "--debug")
	if err != nil {
		if len(output) > 0 {
			return "", errors.New(output)
		}
		return "", err
	}
	result := strings.Split(output, "\n")
	if len(result) > 0 && result[len(result)-2] == "Done." {
		if !notExist {
			err = deleteOldTgz(app+"-dev", newVersion.String())
			if err != nil {
				klog.Info("delete chartmuseum old tgz error, ", err)
			}
		}

	}
	klog.Infof("update repo app %s, newVersion: %s", app, uploadChartVersion)
	return uploadChartVersion, nil

}

func deleteOldTgz(name, notDeleteVersion string) error {
	chartVersions, err := getChartVersions(name)
	if err != nil {
		return err
	}
	errs := make([]error, 0, len(chartVersions)-1)
	for _, cv := range chartVersions {
		if cv.Version == notDeleteVersion {
			continue
		}
		err = deleteChartVersion(name, cv.Version)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return AggregateErrs(errs)
}
