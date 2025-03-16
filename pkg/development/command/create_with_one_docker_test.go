package command

import (
	"testing"
)

func TestWithDockerAppCfg(t *testing.T) {
	cfg := &CreateWithOneDockerConfig{
		ID:   "id",
		Name: "cc",
		Container: CreateWithOneDockerContainer{
			Name:         "beclab/app",
			StartCmd:     "",
			StartCmdArgs: "",
			Port:         8090,
		},
		RequiredCpu:    "1",
		LimitedCpu:     "2",
		RequiredMemory: "10Mi",
		LimitedMemory:  "20Mi",
		RequiredGpu:    false,
		NeedPg:         false,
		NeedRedis:      false,
		Env:            map[string]string{},
		Mounts:         map[string]string{},
	}
	at := AppTemplate{}
	at.WithDockerCfg(cfg).WithDockerDeployment(cfg).WithDockerDeployment(cfg).WithDockerService(cfg).WithDockerChartMetadata(cfg).WithDockerOwner(cfg)
	//b, _ := json.Marshal(at.appcfg)
	//yml, _ := yaml.JSONToYAML(b)
	//fmt.Println(string(yml))
	at.WriteDockerFile(cfg, "./tmp")
}
