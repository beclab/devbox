package envoy

import "os"

const (
	UUIDAnnotation             = "sidecar.bytetrade.io/proxy-uuid"
	SidecarConfigMapVolumeName = "devbox-sidecar-configs"
	SidecarInitContainerName   = "olares-sidecar-init"

	EnvoyUID                      int64 = 1000
	DefaultEnvoyLogLevel                = "debug"
	EnvoyImageVersion                   = "envoyproxy/envoy-distroless:v1.25.2"
	EnvoyContainerName                  = "olares-envoy-sidecar"
	EnvoyAdminPort                      = 15000
	EnvoyAdminPortName                  = "proxy-admin"
	EnvoyInboundListenerPort            = 15003
	EnvoyInboundListenerPortName        = "proxy-inbound"
	EnvoyOutboundListenerPort           = 15001
	EnvoyOutboundListenerPortName       = "proxy-outbound"
	EnvoyLivenessProbePort              = 15008
	EnvoyConfigFileName                 = "envoy.yaml"
	EnvoyConfigFilePath                 = "/etc/envoy"

	WsContainerName = "olares-ws-sidecar"
)

var (
	WsContainerImage = "beclab/ws-gateway:v1.0.3"
)

func init() {
	image := os.Getenv("WS_CONTAINER_IMAGE")
	if image != "" {
		WsContainerImage = image
	}
}
