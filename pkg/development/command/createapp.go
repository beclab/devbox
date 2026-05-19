package command

import (
	"context"
	"encoding/json"
	oac "github.com/beclab/Olares/framework/oac"
	"helm.sh/helm/v3/pkg/chart"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

const (
	defaultIcon = "https://app.cdn.olares.com/appstore/default/defaulticon.webp"
)

type AppTemplate struct {
	appCfg        *oac.AppConfiguration
	deployment    *appsv1.Deployment
	services      []*corev1.Service
	chartMetadata *chart.Metadata
	owner         *Owner

	// useNewSchema is set once at the entry of Run-style commands by
	// inspecting the cluster's Terminus CR. When true, the builder writes the
	// new-schema OlaresManifest layout (ConfigVersion 0.12.0,
	// spec.resources[Mode=cpu], no spec.requiredX/limitedX); when false the
	// legacy layout is used.
	useNewSchema bool

	// ctx is captured at construction so later builder methods can issue
	// cluster lookups (e.g. GPU type discovery) that respect the caller's
	// cancellation. It is never nil; NewAppTemplate substitutes
	// context.Background() when given a nil ctx.
	ctx context.Context
}

// NewAppTemplate returns an AppTemplate with useNewSchema initialised by
// detecting the cluster's Terminus CR version. Use this at every entry point
// that builds an AppTemplate so the schema choice stays consistent.
func NewAppTemplate(ctx context.Context) *AppTemplate {
	if ctx == nil {
		ctx = context.Background()
	}
	return &AppTemplate{useNewSchema: detectNewSchema(ctx), ctx: ctx}
}

type Owner struct {
}

func ToYaml(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	yml, err := yaml.JSONToYAML(b)
	if err != nil {
		return []byte{}, err
	}
	return yml, err
}
