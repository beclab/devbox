package application

type AppMetaData struct {
	Name        string   `yaml:"name" json:"name"`
	Icon        string   `yaml:"icon" json:"icon"`
	Description string   `yaml:"description" json:"description"`
	AppID       string   `yaml:"appid,omitempty" json:"appid,omitempty"`
	Title       string   `yaml:"title" json:"title"`
	Version     string   `yaml:"version" json:"version"`
	Categories  []string `yaml:"categories,omitempty" json:"categories,omitempty"`
	//Rating      float32  `yaml:"rating,omitempty" json:"rating,omitempty"`
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
}

type AppConfiguration struct {
	ConfigVersion string      `yaml:"olaresManifest.version" json:"olaresManifest.version"`
	ConfigType    string      `yaml:"olaresManifest.type" json:"olaresManifest.type"`
	Metadata      AppMetaData `yaml:"metadata" json:"metadata"`
	Entrances     []Entrance  `yaml:"entrances" json:"entrances"`
	Spec          AppSpec     `yaml:"spec" json:"spec"`
	Permission    Permission  `yaml:"permission,omitempty" json:"permission,omitempty" description:"app permission request"`
	Middleware    *Middleware `yaml:"middleware,omitempty" json:"middleware,omitempty" description:"app middleware request"`
	Options       Options     `yaml:"options,omitempty" json:"options,omitempty" description:"app options"`
}

type AppSpec struct {
	VersionName        string         `yaml:"versionName,omitempty" json:"versionName,omitempty"`
	FullDescription    string         `yaml:"fullDescription,omitempty" json:"fullDescription,omitempty"`
	UpgradeDescription string         `yaml:"upgradeDescription,omitempty" json:"upgradeDescription,omitempty"`
	PromoteImage       []string       `yaml:"promoteImage,omitempty" json:"promoteImage,omitempty"`
	PromoteVideo       string         `yaml:"promoteVideo,omitempty" json:"promoteVideo,omitempty"`
	SubCategory        string         `yaml:"subCategory,omitempty" json:"subCategory,omitempty"`
	Developer          string         `yaml:"developer,omitempty" json:"developer,omitempty"`
	RequiredMemory     string         `yaml:"requiredMemory,omitempty" json:"requiredMemory,omitempty"`
	RequiredDisk       string         `yaml:"requiredDisk,omitempty" json:"requiredDisk,omitempty"`
	SupportClient      *SupportClient `yaml:"supportClient,omitempty" json:"supportClient,omitempty"`
	SupportArch        []string       `yaml:"supportArch" json:"supportArch"`
	RequiredGPU        string         `yaml:"requiredGpu,omitempty" json:"requiredGpu,omitempty"`
	RequiredCPU        string         `yaml:"requiredCpu,omitempty" json:"requiredCpu,omitempty"`
	LimitedMemory      string         `yaml:"limitedMemory,omitempty" json:"limitedMemory"`
	LimitedCPU         string         `yaml:"limitedCpu,omitempty" json:"limitedCpu,omitempty"`

	Language     []string     `yaml:"language,omitempty" json:"language,omitempty"`
	Submitter    string       `yaml:"submitter,omitempty" json:"submitter,omitempty"`
	Doc          string       `yaml:"doc,omitempty" json:"doc,omitempty"`
	Website      string       `yaml:"website,omitempty" json:"website,omitempty"`
	FeatureImage string       `yaml:"featuredImage,omitempty" json:"featuredImage,omitempty"`
	SourceCode   string       `yaml:"sourceCode,omitempty" json:"sourceCode,omitempty"`
	License      []TextAndURL `yaml:"license,omitempty" json:"license,omitempty"`
	Legal        []TextAndURL `yaml:"legal,omitempty" json:"legal,omitempty"`
}

type TextAndURL struct {
	Text string `yaml:"text" json:"text" bson:"text"`
	URL  string `yaml:"url" json:"url" bson:"url"`
}

type SupportClient struct {
	Edge    string `yaml:"edge" json:"edge"`
	Android string `yaml:"android" json:"android"`
	Ios     string `yaml:"ios" json:"ios"`
	Windows string `yaml:"windows" json:"windows"`
	Mac     string `yaml:"mac" json:"mac"`
	Linux   string `yaml:"linux" json:"linux"`
}

type Permission struct {
	AppData  bool         `yaml:"appData,omitempty" json:"appData,omitempty"  description:"app data permission for writing"`
	AppCache bool         `yaml:"appCache" json:"appCache"`
	UserData []string     `yaml:"userData" json:"userData"`
	SysData  []SysDataCfg `yaml:"sysData,omitempty" json:"sysData,omitempty"  description:"system shared data permission for accessing"`
}

type SysDataCfg struct {
	Group    string   `yaml:"group" json:"group"`
	DataType string   `yaml:"dataType" json:"dataType"`
	Version  string   `yaml:"version" json:"version"`
	Ops      []string `yaml:"ops" json:"ops"`
}

type Policy struct {
	EntranceName string `yaml:"entranceName" json:"entranceName"`
	Description  string `yaml:"description" json:"description" description:"the description of the policy"`
	URIRegex     string `yaml:"uriRegex" json:"uriRegex" description:"uri regular expression"`
	Level        string `yaml:"level" json:"level"`
	OneTime      bool   `yaml:"oneTime" json:"oneTime"`
	Duration     string `yaml:"validDuration" json:"validDuration"`
}

type Analytics struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type Dependency struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
	// dependency type: system, application.
	Type string `yaml:"type" json:"type"`
}

type ResetCookie struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type Options struct {
	Policies     []Policy     `yaml:"policies,omitempty" json:"policies,omitempty"`
	Analytics    Analytics    `yaml:"analytics" json:"analytics"`
	ResetCookie  ResetCookie  `yaml:"resetCookie" json:"resetCookie"`
	Dependencies []Dependency `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	AppScope     AppScope     `yaml:"appScope" json:"appScope"`
	WsConfig     *WsConfig    `yaml:"websocket,omitempty" json:"websocket,omitempty"`
	Upload       *Upload      `yaml:"upload,omitempty" json:"upload,omitempty"`
}

type AppScope struct {
	ClusterScoped bool     `yaml:"clusterScoped" json:"clusterScoped"`
	AppRef        []string `yaml:"appRef" json:"appRef"`
}

type Entrance struct {
	Name       string `yaml:"name" json:"name"`
	Host       string `yaml:"host" json:"host"`
	Port       int32  `yaml:"port" json:"port"`
	Icon       string `yaml:"icon,omitempty" json:"icon,omitempty"`
	Title      string `yaml:"title" json:"title"`
	AuthLevel  string `yaml:"authLevel,omitempty" json:"authLevel,omitempty"`
	Invisible  bool   `yaml:"invisible,omitempty" json:"invisible,omitempty"`
	OpenMethod string `yaml:"openMethod" json:"openMethod"`
}

type WsConfig struct {
	Port int    `yaml:"port" json:"port"`
	URL  string `yaml:"url" json:"url"`
}

type Upload struct {
	FileType    []string `yaml:"fileType" json:"fileType"`
	Dest        string   `yaml:"dest" json:"dest"`
	LimitedSize int      `yaml:"limitedSize" json:"limitedSize"`
}
