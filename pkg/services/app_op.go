package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/beclab/devbox/pkg/constants"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	"github.com/beclab/devbox/pkg/appcfg"
	"github.com/beclab/devbox/pkg/utils"
)

const (
	appServiceHost   = "http://app-service.os-framework:6755"
	installApiPath   = "/app-service/v1/apps/%s/install"
	canDeployApiPath = "/app-service/v1/apps/%s/can-deploy"
	appStatusApiPath = "/app-service/v1/apps/%s/status"
	uninstallApiPath = "/app-service/v1/apps/%s/uninstall"
)

type Response struct {
	Code int32 `json:"code"`
}

type CanDeployResponse struct {
	Response
	Data CanDeployResponseData `json:"data"`
}

type CanDeployResponseData struct {
	CanOp bool `json:"canOp"`
}

type AppOp interface {
	UpdateAppTitle(ctx context.Context, owner, name, title string) (int64, error)
	Install(ctx context.Context, owner, devAppName, token string) error
	IsAllowedDeploy(ctx context.Context, owner, devAppName, token string) (bool, error)
	Uninstall(ctx context.Context, owner, devAppName, token string) (map[string]interface{}, error)
	CheckIfAppIsUninstalled(owner, devAppName, token string) (bool, error)
}

type appOp struct{}

func NewAppOp() AppOp {
	return &appOp{}
}

func (a *appOp) UpdateAppTitle(ctx context.Context, owner, name, title string) (int64, error) {
	appId, err := utils.UpdateDevApp(owner, name, map[string]interface{}{"title": title})
	if err != nil {
		return 0, err
	}
	appDir := filepath.Join(utils.GetUserBaseDir(owner), name)
	_, err = os.Stat(filepath.Join(appDir, constants.AppCfgFileName))
	if os.IsExist(err) {
		if err := appcfg.UpdateMetadataField(appDir, "title", title); err != nil {
			return 0, err
		}
	}
	return appId, nil
}

func (a *appOp) Install(ctx context.Context, owner, devAppName, token string) error {
	url := fmt.Sprintf("%s%s", appServiceHost, fmt.Sprintf(installApiPath, devAppName))
	client := resty.New()
	body := map[string]interface{}{
		"source":  "devbox",
		"repoUrl": chartRepoHost,
	}
	klog.Infof("install request body: %v", body)
	resp, err := client.R().SetContext(ctx).
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader("X-Market-Source", chartSourceStudio).
		SetHeader(constants.XAuthorization, token).
		SetHeader(constants.XBflUser, owner).
		SetBody(body).Post(url)
	if err != nil {
		klog.Errorf("send install request failed : %v", err)
		return err
	}
	klog.Infof("install: statusCode: %d", resp.StatusCode())
	if resp.StatusCode() != http.StatusOK {
		klog.Errorf("get response from url=%s, with statusCode=%d,err=%v", url, resp.StatusCode(), resp.String())
		return errors.New(string(resp.Body()))
	}

	return nil
}
func (a *appOp) IsAllowedDeploy(ctx context.Context, owner, devAppName, token string) (bool, error) {
	url := fmt.Sprintf("%s%s", appServiceHost, fmt.Sprintf(canDeployApiPath, devAppName))
	client := resty.New().SetTimeout(5 * time.Second)
	resp, err := client.R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader(constants.XAuthorization, token).
		SetHeader(constants.XBflUser, owner).
		Get(url)
	if err != nil {
		klog.Errorf("failed to send request %v", err)
		return false, err
	}
	if resp.StatusCode() != http.StatusOK {
		dump, e := httputil.DumpRequest(resp.Request.RawRequest, true)
		if e == nil {
			klog.Error("request error, ", string(dump))
		}
		return false, errors.New(string(resp.Body()))
	}
	uninstalled, err := a.CheckIfAppIsUninstalled(owner, devAppName, token)
	if err != nil {
		return false, err
	}
	if uninstalled {
		return true, nil
	}

	klog.Infof("resp:xxx: %v", string(resp.Body()))
	data := CanDeployResponse{}
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		klog.Errorf("unmarshal can deploy response error %v", err)
		return false, err
	}
	return data.Data.CanOp, nil
}

func (a *appOp) Uninstall(ctx context.Context, owner, devAppName, token string) (map[string]interface{}, error) {
	uninstalled, err := a.CheckIfAppIsUninstalled(owner, devAppName, token)
	if err != nil {
		return nil, err
	}
	if uninstalled {
		return nil, nil
	}
	url := fmt.Sprintf("%s%s", appServiceHost, fmt.Sprintf(uninstallApiPath, devAppName))
	client := resty.New()
	data := make(map[string]interface{})
	resp, err := client.R().SetContext(ctx).
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader(constants.XAuthorization, token).
		SetHeader(constants.XBflUser, owner).
		Post(url)
	if err != nil {
		klog.Errorf("failed to send request to uninstall app %s, err=%v", devAppName, err)
		return data, err
	}
	klog.Info("request uninstall resp.StatusCode: ", resp.StatusCode())
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New(string(resp.Body()))
	}
	klog.Info("resp.Body: ", string(resp.Body()))
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (a *appOp) CheckIfAppIsUninstalled(owner, devAppName, token string) (bool, error) {
	url := fmt.Sprintf("%s%s", appServiceHost, fmt.Sprintf(appStatusApiPath, devAppName))
	data := make(map[string]interface{})

	client := resty.New()
	resp, err := client.R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader(constants.XAuthorization, token).
		SetHeader(constants.XBflUser, owner).
		Get(url)
	if err != nil {
		klog.Errorf("failed to send request to get app status %s, err=%v", devAppName, err)
		return false, err
	}
	klog.Infof("request app %s status resp.StatusCode: %d", devAppName, resp.StatusCode())
	if resp.StatusCode() == http.StatusNotFound {
		return true, nil
	}
	if resp.StatusCode() != http.StatusOK {
		return false, errors.New(string(resp.Body()))
	}
	klog.Info("resp.Body: ", string(resp.Body()))
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return false, err
	}
	appStatus, ok := data["status"]
	if !ok {
		return false, fmt.Errorf("status filed not found")
	}
	statusMap, ok := appStatus.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("status is not a map")
	}
	state, ok := statusMap["state"].(string)
	if !ok {
		return false, fmt.Errorf("state is not a string")
	}
	if state == "uninstalled" || state == "installFailed" ||
		state == "pendingCanceled" || state == "downloadingCanceled" ||
		state == "installingCanceled" || state == "downloadFailed" ||
		state == "pendingCancelFailed" || state == "downloadingCancelFailed" ||
		state == "installingCancelFailed" {
		return true, nil
	}

	return false, nil
}
