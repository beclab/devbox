package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/beclab/devbox/pkg/constants"
	"github.com/go-resty/resty/v2"
	"github.com/thoas/go-funk"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

const (
	chartSourceStudio = "studio"
	chartRepoHost     = "http://chart-repo-service.os-framework:82"
	deleteApiPath     = "/chart-repo/api/v2/local-apps/delete"
	uploadApiPath     = "/chart-repo/api/v2/apps/upload"
	versionsApiPath   = "/api/v1/charts/%s/versions"
)

type ChartVersions struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    VersionData `json:"data"`
}
type VersionData struct {
	ChartName string   `json:"chart_name"`
	Versions  []string `json:"versions"`
	Total     int      `json:"total_count"`
}

type ChartOp interface {
	Upload(ctx context.Context, owner, devAppName, token, version string) error
	Delete(ctx context.Context, owner, devAppName, token, version string) error
	CheckVersion(ctx context.Context, owner, devAppName, version string) (bool, error)
}

type chartOp struct {
}

func NewChartOp() ChartOp {
	return &chartOp{}
}

func (c *chartOp) Delete(ctx context.Context, owner, devAppName, token, version string) error {
	url := fmt.Sprintf("%s%s", chartRepoHost, deleteApiPath)
	client := resty.New().SetTimeout(time.Minute)
	resp, err := client.R().SetContext(ctx).
		SetHeader("X-Authorization", token).
		SetHeader("X-Bfl-User", owner).
		SetBody(map[string]interface{}{
			"app_name":    devAppName,
			"app_version": version,
			"source_id":   chartSourceStudio,
		}).Delete(url)
	if err != nil {
		klog.Errorf("failed to send request delete chart in market %v", err)
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		klog.Errorf("/chart-repo/api/v2/local-apps/delete status code not = 200, err=%v", string(resp.Body()))
		return errors.New(string(resp.Body()))
	}
	return nil
}

func (c *chartOp) Upload(ctx context.Context, owner, devAppName, token, version string) error {
	url := fmt.Sprintf("%s%s", chartRepoHost, uploadApiPath)
	chartFilePath := fmt.Sprintf("/storage/%s/%s-%s.tgz", owner, devAppName, version)
	client := resty.New().SetTimeout(5 * time.Minute)
	resp, err := client.R().SetContext(ctx).
		SetHeader(constants.XAuthorization, token).
		SetHeader(constants.XBflUser, owner).
		SetFile("chart", chartFilePath).
		SetFormData(map[string]string{
			"source": chartSourceStudio,
		}).Post(url)
	if err != nil {
		klog.Errorf("upload app %s chart to market failed %v", devAppName, err)
		return fmt.Errorf("upload app %s chart to market failed %w", devAppName, err)
	}
	if resp.StatusCode() != http.StatusOK {
		klog.Errorf("/chart-repo/api/v2/apps/upload status code not = 200, err=%v", string(resp.Body()))
		return errors.New(string(resp.Body()))
	}
	klog.Infof("update app %s chart to market success", devAppName)
	return nil
}

func (c *chartOp) CheckVersion(ctx context.Context, owner, devAppName, version string) (bool, error) {
	url := fmt.Sprintf("%s%s", chartRepoHost, fmt.Sprintf(versionsApiPath, devAppName))
	client := resty.New()
	resp, err := client.R().
		SetHeader("X-Market-Source", chartSourceStudio).
		SetHeader("X-Market-User", owner).
		Get(url)
	if err != nil {
		klog.Errorf("check chart %s version failed %v", devAppName, err)
		return false, err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode() != http.StatusOK {
		klog.Errorf("/api/v1/charts/%s/versions status code not = 200, err=%v", string(resp.Body()))
		return false, errors.New(string(resp.Body()))
	}
	var ret ChartVersions
	err = json.Unmarshal(resp.Body(), &ret)
	if err != nil {
		klog.Errorf("unmarshal data to chartVersion failed %v", err)
		return false, err
	}
	t := funk.Contains(ret.Data.Versions, version)
	return t, nil
}
