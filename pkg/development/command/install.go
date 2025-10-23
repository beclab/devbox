package command

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/beclab/devbox/pkg/services"
	"github.com/beclab/devbox/pkg/utils"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"k8s.io/klog/v2"
)

type install struct {
}

func Install() *install {
	return &install{}
}

func (c *install) Run(ctx context.Context, owner, name string, token string, version string) error {
	// update chartVersion in dev-app
	_, err := utils.UpdateDevApp(owner, name, map[string]interface{}{
		"chart_version": version,
	})
	if err != nil {
		klog.Errorf("failed to update chart version %v", err)
		return err
	}
	appOp := services.NewAppOp()
	err = appOp.Install(ctx, owner, utils.DevName(name), token)
	if err != nil {
		return err
	}
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
