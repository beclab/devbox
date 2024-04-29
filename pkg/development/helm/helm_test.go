package helm

import (
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestUpgradeVersion(t *testing.T) {
	chart, err := LoadChart("/tmp/testapp")
	if err != nil {
		t.Fail()
		return
	}

	err = UpgradeChartVersion(chart, "testapp", "/tmp/testapp", &semver.Version{})
	if err != nil {
		t.Fail()
		return
	}

	t.Log("upgraded, ", chart.Metadata.Version)
}

func TestParseMani(t *testing.T) {
	manifest := `
---
apiVersion: iam.kubesphere.io/v1alpha2
kind: User
metadata:
  name: "liuyu"
  annotations:
    iam.kubesphere.io/uninitialized: "true"
    helm.sh/resource-policy: keep
    bytetrade.io/owner-role: platform-admin
    bytetrade.io/terminus-name: "asds"
spec:
  email: "asas"
  password: "asdas"
status:
  state: Active	
---
apiVersion: v1
kind: Service
metadata:
  name: authelia-svc
  namespace: user-space-liuyu
spec:
  selector:
    app: authelia
  type: ClusterIP
  ports:
  - protocol: TCP
    name: authelia
    port: 80
    targetPort: 80

`

	res, _ := DecodeManifest(manifest)

	for _, r := range res {
		if r != nil {
			t.Log("resource: ", r.GetObjectKind().GroupVersionKind().Kind)
		}
	}
}
