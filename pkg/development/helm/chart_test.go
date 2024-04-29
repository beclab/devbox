package helm

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	c "helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"
)

func TestChartMetaToYaml(t *testing.T) {
	meta := c.Metadata{
		Name:       "name",
		APIVersion: "v2",
		AppVersion: "v1.2",
	}

	b, err := json.Marshal(&meta)
	if err != nil {
		t.Fatal(err)
	}

	expectedYaml := "apiVersion: v2\nappVersion: v1.2\nname: name\n"

	yml, err := yaml.JSONToYAML(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedYaml, string(yml))
}
