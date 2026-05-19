package command

import (
	"context"
	"github.com/beclab/Olares/framework/oac"

	"github.com/beclab/api/manifest"
	"k8s.io/klog/v2"

	"github.com/beclab/devbox/pkg/utils"
	"github.com/beclab/devbox/pkg/utils/version"
)

const (
	// newSchemaConfigVersion is the OlaresManifest schema version used when the
	// running system reports a version that satisfies version.IsSupported. In
	// this schema the resource fields move from spec.requiredX/limitedX into
	// spec.resources[Mode=cpu].
	newSchemaConfigVersion = "0.12.0"

	// legacySchemaConfigVersion is the OlaresManifest schema version used when
	// running on older systems.
	legacySchemaConfigVersion = "0.8.0"

	// resourceModeCPU is the mode key used inside spec.resources for the CPU
	// resource profile.
	resourceModeCPU = "cpu"
)

// detectNewSchema reports whether the running system supports the new
// OlaresManifest schema (>= newSchemaConfigVersion). It falls back to false
// (legacy schema) when the system version cannot be obtained or parsed.
func detectNewSchema(ctx context.Context) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	sysVer, err := utils.GetSystemVersion(ctx)
	if err != nil {
		klog.Warningf("detectNewSchema: cannot read system version, fallback to legacy schema: %v", err)
		return false
	}
	if !version.IsSupported(sysVer) {
		klog.V(2).Infof("detectNewSchema: system version %q < supported, using legacy schema", sysVer)
		return false
	}
	klog.V(2).Infof("detectNewSchema: system version %q is supported, using new schema", sysVer)
	return true
}

// applyAppResources writes r into spec depending on the active schema.
//
// New schema: clear the legacy fields and project r into
// spec.Resources[Mode=cpu]; replace any existing cpu mode entry. Disk fields
// stay on the legacy spec field because spec.requiredDisk has no analogue in
// the resources mode (it's a global app constraint).
//
// Legacy schema: write r directly to spec.RequiredX/LimitedX.
func applyAppResources(spec *manifest.AppSpec, useNewSchema bool, r oac.ManifestResourceLimits) {
	if !useNewSchema {
		spec.RequiredCPU = r.RequiredCPU
		spec.LimitedCPU = r.LimitedCPU
		spec.RequiredMemory = r.RequiredMemory
		spec.LimitedMemory = r.LimitedMemory
		spec.RequiredDisk = r.RequiredDisk
		spec.LimitedDisk = r.LimitedDisk
		return
	}

	spec.RequiredCPU = ""
	spec.RequiredMemory = ""
	spec.RequiredDisk = ""
	spec.LimitedDisk = ""
	spec.LimitedCPU = ""
	spec.LimitedMemory = ""

	mode := manifest.ResourceMode{
		Mode: resourceModeCPU,
		ResourceRequirement: manifest.ResourceRequirement{
			RequiredCPU:    r.RequiredCPU,
			RequiredMemory: r.RequiredMemory,
			RequiredDisk:   r.RequiredDisk,
			LimitedDisk:    r.LimitedDisk,
			LimitedCPU:     r.LimitedCPU,
			LimitedMemory:  r.LimitedMemory,
		},
	}
	for i := range spec.Resources {
		if spec.Resources[i].Mode == resourceModeCPU {
			spec.Resources[i] = mode
			return
		}
	}
	spec.Resources = append(spec.Resources, mode)
}

// applyAppGpuMode is the GPU-aware counterpart of applyAppResources. It is
// the only resource writer that should be used when the app requires a GPU;
// callers MUST NOT additionally call applyAppResources, otherwise the cpu
// mode and the gpu mode would carry duplicated requirements.
//
// New schema with a known gpuType: clear the legacy spec.requiredX/limitedX
// fields, drop any pre-existing cpu mode (this app is selected by GPU mode,
// not CPU mode), and write the full resource bundle into a single
// ResourceMode keyed by gpuType. An existing entry for the same gpuType is
// replaced in place; other modes are preserved.
//
// New schema without a gpuType (cluster reports no GPU-labelled nodes), or
// legacy schema: write the full resource bundle - including
// RequiredGPU/LimitedGPU - directly into spec.RequiredX/LimitedX so the
// requirement is preserved. A warning is logged in the new-schema fallback
// case so the missing cluster signal is visible.
func applyAppGpuMode(spec *manifest.AppSpec, useNewSchema bool, gpuType string, r oac.ManifestResourceLimits) {
	// Whenever applyAppGpuMode is invoked the app is selected by GPU mode
	// instead of CPU mode, so any pre-existing cpu entry in spec.Resources
	// (typically seeded by an earlier applyAppResources call) must go - even
	// on the legacy/fallback paths below where the actual write target is
	// the spec.RequiredX/LimitedX fields. Otherwise the manifest would
	// advertise both mode:cpu and the GPU requirement at the same time.
	if len(spec.Resources) > 0 {
		filtered := spec.Resources[:0]
		for _, m := range spec.Resources {
			if m.Mode == resourceModeCPU {
				continue
			}
			filtered = append(filtered, m)
		}
		spec.Resources = filtered
	}

	writeLegacy := func() {
		spec.RequiredCPU = r.RequiredCPU
		spec.LimitedCPU = r.LimitedCPU
		spec.RequiredMemory = r.RequiredMemory
		spec.LimitedMemory = r.LimitedMemory
		spec.RequiredDisk = r.RequiredDisk
		spec.LimitedDisk = r.LimitedDisk
		spec.RequiredGPU = r.RequiredGPU
		spec.LimitedGPU = r.LimitedGPU
	}

	if !useNewSchema {
		writeLegacy()
		return
	}
	if gpuType == "" {
		klog.Warningf("applyAppGpuMode: no cluster GPU type discovered, recording resources on legacy spec fields")
		writeLegacy()
		return
	}

	spec.RequiredCPU = ""
	spec.LimitedCPU = ""
	spec.RequiredMemory = ""
	spec.LimitedMemory = ""
	spec.RequiredDisk = ""
	spec.LimitedDisk = ""
	spec.RequiredGPU = ""
	spec.LimitedGPU = ""

	mode := manifest.ResourceMode{
		Mode: gpuType,
		ResourceRequirement: manifest.ResourceRequirement{
			RequiredCPU:    r.RequiredCPU,
			LimitedCPU:     r.LimitedCPU,
			RequiredMemory: r.RequiredMemory,
			LimitedMemory:  r.LimitedMemory,
			RequiredDisk:   r.RequiredDisk,
			LimitedDisk:    r.LimitedDisk,
			RequiredGPU:    r.RequiredGPU,
			LimitedGPU: func() string {
				if r.LimitedGPU != "" {
					return r.LimitedGPU
				}
				return r.RequiredGPU
			}(),
		},
	}

	for i := range spec.Resources {
		if spec.Resources[i].Mode == gpuType {
			spec.Resources[i] = mode
			return
		}
	}
	spec.Resources = append(spec.Resources, mode)
}

// extractAppResources reads the active resource bundle out of spec. It
// inverts what applyAppGpuMode / applyAppResources wrote, looking up entries
// in this preference order:
//
//  1. If a non-empty gpuType is supplied, the spec.Resources entry whose
//     Mode equals gpuType (the layout written by applyAppGpuMode under the
//     new schema).
//  2. The spec.Resources cpu-mode entry (the layout written by
//     applyAppResources under the new schema).
//  3. The legacy spec.RequiredX/LimitedX fields (legacy schema, or the
//     fallback that applyAppGpuMode takes when the cluster reports no GPU
//     type).
//
// The deployment builders pass the cluster's first GPU type so that an app
// requiring a GPU sizes its container off the gpu-mode bundle; non-GPU
// callers pass "" and naturally fall through to the cpu-mode bundle.
func extractAppResources(spec manifest.AppSpec, gpuType string) oac.ManifestResourceLimits {
	if gpuType != "" {
		for _, res := range spec.Resources {
			if res.Mode == gpuType {
				return resourceModeToLimits(res)
			}
		}
	}
	for _, res := range spec.Resources {
		if res.Mode == resourceModeCPU {
			return resourceModeToLimits(res)
		}
	}
	return oac.ManifestResourceLimits{
		RequiredCPU:    spec.RequiredCPU,
		LimitedCPU:     spec.LimitedCPU,
		RequiredMemory: spec.RequiredMemory,
		LimitedMemory:  spec.LimitedMemory,
		RequiredDisk:   spec.RequiredDisk,
		LimitedDisk:    spec.LimitedDisk,
		RequiredGPU:    spec.RequiredGPU,
		LimitedGPU:     spec.LimitedGPU,
	}
}

func resourceModeToLimits(res manifest.ResourceMode) oac.ManifestResourceLimits {
	return oac.ManifestResourceLimits{
		RequiredCPU:    res.RequiredCPU,
		LimitedCPU:     res.LimitedCPU,
		RequiredMemory: res.RequiredMemory,
		LimitedMemory:  res.LimitedMemory,
		RequiredDisk:   res.RequiredDisk,
		LimitedDisk:    res.LimitedDisk,
		RequiredGPU:    res.RequiredGPU,
		LimitedGPU:     res.LimitedGPU,
	}
}

// configVersionFor returns the manifest config version string that matches the
// active schema.
func configVersionFor(useNewSchema bool) string {
	if useNewSchema {
		return newSchemaConfigVersion
	}
	return legacySchemaConfigVersion
}
