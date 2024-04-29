package helm

type ContainerInfo struct {
	Image            string  `json:"image"`
	PodSelector      string  `json:"podSelector"`
	ContainerName    string  `json:"containerName"`
	DevContainerName string  `json:"devContainerName"`
	DevPath          *string `json:"devPath,omitempty"`
	State            *string `json:"state,omitempty"`
	AppID            *int    `json:"appId,omitempty"`
}
