package constants

import (
	"os"
)

const (
	DevOwnerLabel                      = "dev.bytetrade.io/dev-owner"
	AppCfgFileName                     = "OlaresManifest.yaml"
	OwnerLabel                         = "applications.app.bytetrade.io/owner"
	ExposePortsLabel                   = "applications.app.bytetrade.io/studio-expose-ports"
	XAuthorization                     = "X-Authorization"
	XBflUser                           = "X-Bfl-User"
	ApplicationGpuInjectKey            = "applications.app.bytetrade.io/gpu-inject"
	ApplicationDefaultThirdLevelDomain = "applications.app.bytetrade.io/default-thirdlevel-domains"
)

var (
	Namespace = ""
)

func init() {
	Namespace = os.Getenv("NAME_SPACE")

}
