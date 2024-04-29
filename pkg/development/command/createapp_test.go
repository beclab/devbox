package command

import (
	"testing"
)

func TestWithAppCfg(t *testing.T) {
	createConfig := CreateConfig{
		Name:           "cc",
		OSVersion:      ">=0.1.0",
		Img:            "busybox",
		Ports:          []int{8080},
		WebsitePort:    "8080",
		SystemDB:       false,
		Redis:          false,
		MongoDB:        false,
		PostgreSQL:     false,
		SystemCall:     false,
		IngressRouter:  false,
		Traefik:        true,
		AppData:        true,
		UserData:       []string{},
		NeedGPU:        false,
		RequiredGPU:    "",
		RequiredMemory: "1Mi",
	}
	at := AppTemplate{}
	at.WithAppCfg(&createConfig).
		WithDeployment(&createConfig).
		WithService(&createConfig).
		WithChartMetadata(&createConfig).
		WithTraefik(&createConfig)
	//b, _ := json.Marshal(at.appcfg)
	//yml, _ := yaml.JSONToYAML(b)
	//fmt.Println(string(yml))
	at.WriteFile(&createConfig, "/tmp")
}
