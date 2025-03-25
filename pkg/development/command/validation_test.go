package command

import "testing"

func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name    string
		config  CreateWithOneDockerConfig
		isValid bool
	}{
		{
			name: "valid config",
			config: CreateWithOneDockerConfig{
				ID: "a", Name: "a",
				Container: CreateWithOneDockerContainer{
					Image: "nginx",
					Port:  80,
				},
				RequiredCpu:    "1Gi",
				LimitedCpu:     "",
				RequiredMemory: "1Gi",
				LimitedMemory:  "",
				RequiredGpu:    false,
				NeedPg:         false,
				NeedRedis:      false,
				Env:            map[string]string{},
				Mounts:         map[string]string{}},
			isValid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateStruct(tt.config)
			if len(errs) > 0 {
				t.Errorf("validate err %v", errs)
			}
		})
	}
}
