package middlewares

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

var lldapBaseURL = "http://lldap-service.os-platform:17170"

type UserInfo struct {
	Username string `json:"username"`
}

type JWTClaims struct {
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
	Username string `json:"username"`
}

func TokenAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			klog.Warningf("token not found in request")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": "token not found in request",
			})
		}
		userInfo, err := validateToken(token, token)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": "token not found in request",
			})
		}
		klog.Infof("userInfo.Usernaeme: %s", userInfo.Username)

		c.Locals("username", userInfo.Username)
		c.Locals("auth_token", token)
		return c.Next()
	}
}

func extractToken(c *fiber.Ctx) string {
	authToken := c.Get("X-Authorization")
	if authToken != "" {
		return authToken
	}
	authToken = c.Cookies("auth_token")
	return authToken
}

func validateToken(accessToken, validToken string) (*UserInfo, error) {
	url := fmt.Sprintf("%s/auth/token/verify", lldapBaseURL)
	client := resty.New()
	resp, err := client.SetTimeout(30*time.Second).R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(accessToken).
		SetBody(map[string]string{
			"access_token": validToken,
		}).Post(url)
	if err != nil {
		klog.Infof("send request o lldap failed %v", err)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		klog.Infof("request lldap /auth/token/verify not 200, %v, body: %v", resp.StatusCode(), string(resp.Body()))
		return nil, errors.New(resp.String())
	}
	var response JWTClaims
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		klog.Errorf("unmarshal jwt claims failed: %v", err)
		return nil, err
	}
	userInfo := UserInfo{
		Username: response.Username,
	}
	return &userInfo, nil
}
