package command

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

func CreateAppWithHelloWorldConfig(baseDir string, owner, name, title string) error {
	createConfigExample.Name = name
	createConfigExample.Title = title
	at := AppTemplate{}
	at.WithDockerCfg(createConfigExample).WithDockerDeployment(createConfigExample).
		WithDockerService(createConfigExample).WithDockerChartMetadata(createConfigExample).WithDockerOwner(createConfigExample)
	err := at.WriteDockerFile(createConfigExample, owner, baseDir)
	if err != nil {
		return err
	}
	return nil
}
