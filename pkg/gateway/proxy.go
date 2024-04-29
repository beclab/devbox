package gateway

import (
	"context"
	"net/url"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"k8s.io/klog/v2"
)

type DevContainerProxy struct {
	proxy *echo.Echo
	ctx   context.Context
}

func NewDevContainerProxy(ctx context.Context) *DevContainerProxy {
	p := &DevContainerProxy{
		proxy: echo.New(),
		ctx:   ctx,
	}

	config := middleware.DefaultProxyConfig
	config.Balancer = p
	p.proxy.Use(middleware.ProxyWithConfig(config))

	return p
}

func (p *DevContainerProxy) Start(addr string) error {
	klog.Info("gateway start on ", addr)
	return p.proxy.Start(addr)
}

func (p *DevContainerProxy) Shutdown() {
	klog.Info("gateway shutdown")
	if err := p.proxy.Shutdown(p.ctx); err != nil {
		klog.Error("shutdown error, ", err)
	}
}

func (p *DevContainerProxy) Next(c echo.Context) *middleware.ProxyTarget {
	c.Cookies()
	u, err := url.Parse("http://192.168.50.32:3000/")
	if err != nil {
		klog.Error("url error, ", err)
		return nil
	}
	return &middleware.ProxyTarget{URL: u}
}

func (p *DevContainerProxy) AddTarget(*middleware.ProxyTarget) bool { return true }
func (p *DevContainerProxy) RemoveTarget(string) bool               { return true }
