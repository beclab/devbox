package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/beclab/devbox/pkg/constants"

	"github.com/go-resty/resty/v2"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/klog/v2"
)

type AccessTokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiredAt   time.Time `json:"expired_at"`
}

func GetAccessToken() (string, error) {
	timestamp := time.Now().UnixMilli() / 1000
	text := constants.ApiKey + strconv.Itoa(int(timestamp)) + constants.ApiSecret
	token, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	client := resty.New().SetTimeout(5 * time.Second)
	resp, err := client.R().SetBody(
		map[string]interface{}{
			"app_key":   constants.ApiKey,
			"timestamp": timestamp,
			"token":     string(token),
			"perm": map[string]interface{}{
				"group":    "service.appstore",
				"dataType": "app",
				"version":  "v1",
				"ops":      []string{"InstallDevApp", "UninstallDevApp"},
			},
		}).Post(fmt.Sprintf("http://%s/permission/v1alpha1/access", constants.SystemServer))
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		dump, e := httputil.DumpRequest(resp.Request.RawRequest, true)
		if e == nil {
			klog.Error("request system-server.permission/v1alpha1/access", string(dump))
		}
		return "", errors.New(string(resp.Body()))
	}

	at := struct {
		Data AccessTokenResponse `json:"data"`
	}{}
	err = json.Unmarshal(resp.Body(), &at)
	if err != nil {
		return "", err
	}
	return at.Data.AccessToken, nil
}
