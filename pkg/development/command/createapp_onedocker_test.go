package command

import (
	"encoding/json"
	"fmt"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestWithDockerAppCfg(t *testing.T) {
	cfg := &CreateWithOneDockerConfig{
		ID:   "id",
		Name: "cc",
		Container: CreateWithOneDockerContainer{
			Image:        "beclab/app",
			StartCmd:     "",
			StartCmdArgs: "",
			Port:         8090,
		},
		RequiredCpu:    "1",
		LimitedCpu:     "2",
		RequiredMemory: "10Mi",
		LimitedMemory:  "20Mi",
		GpuVendor:      "nvidia",
		RequiredGpu:    false,
		NeedPg:         true,
		NeedRedis:      true,
		Env:            map[string]string{},
		Mounts:         map[string]string{"/app/data/aaa": "/aaa", "/app/cache/bbb": "/bbb", "/Home/ccc": "/ccc", "/app/data/aaa2": "/aaa2"},
	}
	at := AppTemplate{}
	at.WithDockerCfg(cfg).WithDockerDeployment(cfg).WithDockerDeployment(cfg).WithDockerService(cfg).WithDockerChartMetadata(cfg).WithDockerOwner(cfg)
	b, _ := json.Marshal(at.appCfg)
	yml, _ := yaml.JSONToYAML(b)
	fmt.Println(string(yml))
}
