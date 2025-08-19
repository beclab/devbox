package command

import (
	"github.com/beclab/devbox/pkg/utils"
)

var createConfigExample = &CreateWithOneDockerConfig{
	Container: CreateWithOneDockerContainer{
		Image: "beclab/studio-app:0.0.1",
		Port:  80,
	},
	RequiredCpu:    "50m",
	RequiredMemory: "100Mi",
	RequiredGpu:    false,
	NeedPg:         false,
	NeedRedis:      false,
}

func CreateAppWithHelloWorldConfig(owner, name, title string) error {
	at := &AppTemplate{}
	createConfigExample.Name = name
	createConfigExample.Title = title

	at.WithDockerCfg(createConfigExample).WithDockerDeployment(createConfigExample).
		WithDockerService(createConfigExample).WithDockerChartMetadata(createConfigExample).WithDockerOwner(createConfigExample)

	err := at.WriteDockerFile(createConfigExample, utils.GetAppPath(owner, name))
	if err != nil {
		return err
	}
	return nil
}
