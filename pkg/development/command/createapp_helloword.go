package command

var createConfigExample = &CreateWithOneDockerConfig{
	Container: CreateWithOneDockerContainer{
		Image: "bytetrade/devbox-app:0.0.1",
		Port:  8080,
	},
	RequiredCpu:    "50m",
	RequiredMemory: "100Mi",
	RequiredGpu:    false,
	NeedPg:         false,
	NeedRedis:      false,
}

func CreateAppWithHelloWorldConfig(baseDir string, name string) error {
	createConfigExample.Name = name
	at := AppTemplate{}
	at.WithDockerCfg(createConfigExample).WithDockerDeployment(createConfigExample).
		WithDockerService(createConfigExample).WithDockerChartMetadata(createConfigExample).WithDockerOwner(createConfigExample)
	err := at.WriteDockerFile(createConfigExample, baseDir)
	if err != nil {
		return err
	}
	return nil
}
