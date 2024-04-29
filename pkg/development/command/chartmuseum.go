package command

import (
	"fmt"
	"net/http"
	"time"

	"github.com/beclab/devbox/pkg/constants"

	"github.com/go-resty/resty/v2"
	helm_repo "helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

func getChartVersions(name string) (helm_repo.ChartVersions, error) {
	chartVersions := make(helm_repo.ChartVersions, 0)
	client := resty.New().SetTimeout(5 * time.Second)
	url := fmt.Sprintf("http://chartmuseum.user-space-%s:8080/api/charts/%s", constants.Owner, name)
	resp, err := client.R().Get(url)
	if err != nil {
		return chartVersions, err
	}
	if resp.StatusCode() != http.StatusOK {
		return chartVersions, fmt.Errorf("get chart versions from chartmuseum return unexpected status code, %d", resp.StatusCode())
	}
	if err = yaml.Unmarshal(resp.Body(), &chartVersions); err != nil {
		return chartVersions, err
	}
	return chartVersions, nil
}

func deleteChartVersion(name, version string) error {
	client := resty.New().SetTimeout(5 * time.Second)
	url := fmt.Sprintf("http://chartmuseum.user-space-%s:8080/api/charts/%s/%s", constants.Owner, name, version)
	resp, err := client.R().Delete(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("delete chart %s, version %s from chartmuseum return unexpected status code", name, version)
	}
	return nil
}
