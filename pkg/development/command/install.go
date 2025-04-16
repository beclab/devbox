package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/beclab/devbox/pkg/constants"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
)

type install struct {
}

func Install() *install {
	return &install{}
}

func (c *install) Run(ctx context.Context, app string, token string) (string, error) {
	accessToken, err := GetAccessToken()
	if err != nil {
		return "", err
	}
	klog.Infof("run appname: %s", app)

	url := fmt.Sprintf("http://%s/system-server/v1alpha1/app/service.appstore/v1/InstallDevApp", constants.SystemServer)
	client := resty.New().SetTimeout(5 * time.Second)
	resp, err := client.R().SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader("X-Authorization", token).
		SetHeader("X-Access-Token", accessToken).
		SetBody(
			map[string]interface{}{
				"appName": app,
				"repoUrl": constants.RepoURL,
				"source":  "devbox",
			}).Post(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		dump, e := httputil.DumpRequest(resp.Request.RawRequest, true)
		if e == nil {
			klog.Error("reauest bfl.InstallDevApp", string(dump))
		}
		return "", errors.New(string(resp.Body()))
	}
	klog.Infof("body: %s\n", string(resp.Body()))
	ret := make(map[string]interface{})
	err = json.Unmarshal(resp.Body(), &ret)
	if err != nil {
		return "", err
	}

	code, ok := ret["code"]
	if int(code.(float64)) != 0 {
		return "", fmt.Errorf("%s", ret["message"])
	}
	if ok && int(code.(float64)) == 0 {
		data := ret["data"].(map[string]interface{})
		code, ok := data["code"]
		if ok && int(code.(float64)) != http.StatusOK {
			return "", fmt.Errorf("message: %s", data["message"])
		}

	}

	return "", nil

}
