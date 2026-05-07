package appcfg

import (
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/manifest"
	"k8s.io/apimachinery/pkg/api/resource"
)

type AppPermission interface{}

type AppDataPermission string
type AppCachePermission string
type UserDataPermission string

type Middleware interface{}

type SysDataPermission struct {
	AppName   string   `yaml:"appName" json:"appName"`
	Port      string   `yaml:"port" json:"port"`
	Svc       string   `yaml:"svc,omitempty" json:"svc,omitempty"`
	Namespace string   `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Group     string   `yaml:"group" json:"group"`
	DataType  string   `yaml:"dataType" json:"dataType"`
	Version   string   `yaml:"version" json:"version"`
	Ops       []string `yaml:"ops" json:"ops"`
}

type AppRequirement struct {
	Memory *resource.Quantity
	Disk   *resource.Quantity
	GPU    *resource.Quantity
	CPU    *resource.Quantity
}

type AppPolicy struct {
	EntranceName string        `yaml:"entranceName" json:"entranceName"`
	URIRegex     string        `yaml:"uriRegex" json:"uriRegex" description:"uri regular expression"`
	Level        string        `yaml:"level" json:"level"`
	OneTime      bool          `yaml:"oneTime" json:"oneTime"`
	Duration     time.Duration `yaml:"validDuration" json:"validDuration"`
}

const (
	AppDataRW  AppDataPermission  = "appdata-perm"
	AppCacheRW AppCachePermission = "appcache-perm"
	UserDataRW UserDataPermission = "userdata-perm"
)

type ApplicationConfig struct {
	AppID                string
	CfgFileVersion       string
	Namespace            string
	ChartsName           string
	RepoURL              string
	Title                string
	Version              string
	Target               string
	AppName              string // name of application displayed on shortcut
	OwnerName            string // name of owner who installed application
	Entrances            []appv1.Entrance
	Ports                []appv1.ServicePort
	Icon                 string          // base64 icon data
	Permission           []AppPermission // app permission requests
	Requirement          AppRequirement
	Policies             []AppPolicy
	Middleware           *manifest.Middleware
	AnalyticsEnabled     bool
	ResetCookieEnabled   bool
	Dependencies         []manifest.Dependency
	AppScope             manifest.AppScope
	WsConfig             manifest.WsConfig
	Upload               manifest.Upload
	OnlyAdmin            bool
	MobileSupported      bool
	ApiTimeout           *int64
	RunAsUser            bool
	AllowedOutboundPorts []int
	RequiredGPU          string
}

type DefaultThirdLevelDomainConfig struct {
	AppName          string `json:"appName"`
	EntranceName     string `json:"entranceName"`
	ThirdLevelDomain string `json:"thirdLevelDomain,omitempty"`
}
