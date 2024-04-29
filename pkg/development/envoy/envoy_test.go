package envoy

import (
	"os"
	"testing"

	"k8s.io/klog/v2"
)

func TestConfig(t *testing.T) {
	config, err := (&ConfigBuilder{}).WithDevcontainers([]*DevcontainerEndpoint{
		{Name: "dev1", Host: "localhost", Port: 4000, Path: "/proxy/5000"},
		{Name: "dev2", Host: "localhost", Port: 4000, Path: "/proxy/5001"},
	}).Build()
	if err != nil {
		klog.Error(err)
		t.Fail()
	} else {
		os.WriteFile("/tmp/c.yaml", []byte(config), 0644)
		t.Log(config)
	}
}
