package constants

import (
	"fmt"
	"os"
)

const (
	DevOwnerLabel  = "dev.bytetrade.io/dev-owner"
	AppCfgFileName = "OlaresManifest.yaml"
)

var (
	Namespace    = ""
	Owner        = ""
	RepoURL      = ""
	ApiKey       = ""
	ApiSecret    = ""
	SystemServer = ""
)

func init() {
	Namespace = os.Getenv("NAME_SPACE")
	Owner = os.Getenv("OWNER")
	RepoURL = fmt.Sprintf("http://chartmuseum-studio.%s.svc.cluster.local:8080/", Namespace)
	ApiKey = os.Getenv("OS_API_KEY")
	ApiSecret = os.Getenv("OS_API_SECRET")
	SystemServer = os.Getenv("OS_SYSTEM_SERVER")

}
