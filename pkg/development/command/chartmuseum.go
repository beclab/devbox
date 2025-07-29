package command

import (
	"errors"
	"fmt"
	"k8s.io/klog/v2"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	helm_repo "helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

func getChartVersions(owner, name string) (helm_repo.ChartVersions, error) {
	chartVersions := make(helm_repo.ChartVersions, 0)
	client := resty.New().SetTimeout(5 * time.Second)
	url := fmt.Sprintf("http://127.0.0.1:8888/%s/api/charts/%s", owner, name)
	resp, err := client.R().Get(url)
	if err != nil {
		klog.Errorf("failed to send request to url=%s,err=%v", url, err)
		return chartVersions, err
	}
	if resp.StatusCode() != http.StatusOK {
		klog.Errorf("get chart versions from chartmuseum return unexpected status code %d,err=%v", resp.StatusCode(), resp.String())
		return chartVersions, fmt.Errorf("get chart versions from chartmuseum return unexpected status code, %d", resp.StatusCode())
	}
	if err = yaml.Unmarshal(resp.Body(), &chartVersions); err != nil {
		klog.Errorf("failed to unmarshal body to chartVersions %v", err)
		return chartVersions, err
	}
	return chartVersions, nil
}

func deleteChartVersion(owner, name, version string) error {
	client := resty.New().SetTimeout(5 * time.Second)
	url := fmt.Sprintf("http://127.0.0.1:8888/%s/api/charts/%s/%s", owner, name, version)
	resp, err := client.R().Delete(url)
	if err != nil {
		klog.Errorf("failed to send request to url=%s,err=%v", url, err)
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		msg := fmt.Sprintf("failed to delete chart %s, version %s from chart repo return unexpected status code %v,err=%v", name, version, resp.StatusCode(), resp.String())
		klog.Error(msg)
		return errors.New(msg)
	}
	return nil
}
