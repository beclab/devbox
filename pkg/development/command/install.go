package command

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
)

type install struct {
}

func Install() *install {
	return &install{}
}

func (c *install) Run(ctx context.Context, app string, token string, version string) (string, error) {
	klog.Infof("run appname: %s", app)

	err := c.UploadChartToMarket(ctx, app, token, version)
	if err != nil {
		return "", err
	}
	for i := 0; i < 45; i++ {
		klog.Infof("wait for chart %d", i)
		time.Sleep(time.Second)
	}

	// get chart tgz file from storage and push to market
	// if more than one user upload same name tgz file to market what would happen

	url := fmt.Sprintf("http://appstore-service.os-framework:81/app-store/api/v2/apps/%s/install", app)
	client := resty.New().SetTimeout(5 * time.Second)
	body := map[string]interface{}{
		"source":   "local",
		"app_name": app,
		"version":  version,
	}
	klog.Infof("install request body: %v", body)
	resp, err := client.R().SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader("X-Authorization", token).
		SetBody(body).Post(url)
	if err != nil {
		klog.Errorf("send install  request failed : %v", err)
		return "", err
	}
	klog.Infof("install: statusCode: %d", resp.StatusCode())
	if resp.StatusCode() != http.StatusOK {
		dump, e := httputil.DumpRequest(resp.Request.RawRequest, true)
		if e == nil {
			klog.Error("request bfl.InstallDevApp", string(dump))
		}
		return "", errors.New(string(resp.Body()))
	}
	klog.Infof("body: %s\n", string(resp.Body()))

	return "", nil

}

func (c *install) UploadChartToMarket(ctx context.Context, app string, token string, version string) error {
	client := resty.New().SetTimeout(30 * time.Second)

	chartFilePath := fmt.Sprintf("/storage/%s-%s.tgz", app, version)
	klog.Infof("chartFilePath: %s", chartFilePath)
	resp, err := client.R().
		SetHeader("X-Authorization", token).
		SetFile("chart", chartFilePath).
		SetFormData(map[string]string{
			"source": "local",
		}).Post("http://appstore-service.os-framework:81/app-store/api/v2/apps/upload")
	if err != nil {
		klog.Errorf("upload app %s chart to market failed %w", app, err)
		return fmt.Errorf("upload app %s chart to market failed %w", app, err)
	}
	if resp.StatusCode() != http.StatusOK {
		dump, e := httputil.DumpRequest(resp.Request.RawRequest, true)
		if e != nil {
			klog.Error("request /app-store/api/v2/apps/upload", string(dump))
		}
		klog.Errorf("status code not = 200, err=%v", string(resp.Body()))
		return errors.New(string(resp.Body()))
	}
	klog.Infof("update app %s chart to market success", app)
	return nil
}
