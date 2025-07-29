package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"github.com/nats-io/nats.go"
	"k8s.io/klog/v2"
)

type install struct {
}

func Install() *install {
	return &install{}
}

func (c *install) Run(ctx context.Context, owner, app string, token string, version string) (string, error) {
	klog.Infof("run appname: %s", app)

	err := c.UploadChartToMarket(ctx, owner, app, token, version)
	if err != nil {
		klog.Errorf("failed to upload app=%s chart to market chart repo %v", app, err)
		return "", err
	}
	err = c.waitForMarketUpdate(ctx, owner, app, version)
	if err != nil {
		klog.Errorf("wait market ready app: %s failed, err=%v", app, err)
		return "", err
	}

	// get chart tgz file from storage and push to market
	// if more than one user upload same name tgz file to market what would happen

	url := fmt.Sprintf("http://appstore-service.os-framework:81/app-store/api/v2/apps/%s/install", app)
	client := resty.New().SetTimeout(30 * time.Second)
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
		klog.Errorf("send install request failed : %v", err)
		return "", err
	}
	klog.Infof("install: statusCode: %d", resp.StatusCode())
	if resp.StatusCode() != http.StatusOK {
		klog.Errorf("get response from url=%s, with statusCode=%d,err=%v", url, resp.StatusCode(), resp.String())
		return "", errors.New(string(resp.Body()))
	}

	return "", nil

}

func (c *install) UploadChartToMarket(ctx context.Context, owner, app string, token string, version string) error {
	client := resty.New().SetTimeout(2 * time.Minute)

	chartFilePath := fmt.Sprintf("/storage/%s/%s-%s.tgz", owner, app, version)
	klog.Infof("chartFilePath: %s", chartFilePath)
	resp, err := client.R().
		SetHeader("X-Authorization", token).
		SetFile("chart", chartFilePath).
		SetFormData(map[string]string{
			"source": "local",
		}).Post("http://appstore-service.os-framework:81/app-store/api/v2/apps/upload")
	if err != nil {
		klog.Errorf("upload app %s chart to market failed %v", app, err)
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

func (c *install) waitForMarketUpdate(ctx context.Context, owner, app, version string) error {
	natsHost := os.Getenv("NATS_HOST")
	natsPort := os.Getenv("NATS_PORT")

	//subject = os.Getenv("NATS_SUBJECT")
	username := os.Getenv("NATS_USERNAME")
	password := os.Getenv("NATS_PASSWORD")

	natsURL := fmt.Sprintf("nats://%s:%s", natsHost, natsPort)
	nc, err := nats.Connect(natsURL, nats.UserInfo(username, password))
	if err != nil {
		klog.Errorf("failed to connect to nats: %s, err=%v", natsURL, err)
		return err
	}
	subject := fmt.Sprintf("os.market.%s", owner)
	klog.Infof("subscribe subject: %s", subject)

	msgChan := make(chan *nats.Msg, 1)
	timeoutChan := time.After(2 * time.Minute)

	sub, err := nc.ChanSubscribe(subject, msgChan)
	if err != nil {
		klog.Errorf("failed to subscribe subject: %s,err=%v", subject, err)
		return err
	}
	defer sub.Unsubscribe()
	klog.Infof("start to wait market update message, timeout is 2 minutes")
	for {
		select {
		case msg := <-msgChan:
			var updateInfo MarketSystemUpdate
			if err := json.Unmarshal(msg.Data, &updateInfo); err != nil {
				klog.Errorf("failed to unmarshal marketSystemUpdate %v", err)
				continue
			}
			klog.Infof("message: %#v", updateInfo)
			if updateInfo.NotifyType != "market_system_point" {
				continue
			}
			if updateInfo.User != owner {
				continue
			}
			if updateInfo.Extensions["app_name"] != app {
				continue
			}
			if updateInfo.Extensions["app_version"] != version {
				continue
			}
			if updateInfo.Point != "new_app_ready" {
				continue
			}

			return nil
		case <-timeoutChan:
			klog.Infof("wait for market app:%s ready timeout after 2 minutes", app)
			return fmt.Errorf("wait for market ready timeout")
		case <-ctx.Done():
			klog.Infof("wait for market ready context canceled")
			return ctx.Err()
		}

	}

}
