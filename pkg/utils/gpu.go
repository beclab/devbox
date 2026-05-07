package utils

import (
	"context"
	"errors"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NodeGPUTypeLabel is the label that app-service writes onto cluster nodes
// to advertise the GPU model present on each node. Mirrors the constant used
// by the Olares app-service so devbox sees the same values.
const NodeGPUTypeLabel = "gpu.bytetrade.io/type"

// GetAllGpuTypesFromNodes extracts the set of unique GPU type strings from
// the NodeGPUTypeLabel of each node. Empty label values are ignored. It
// returns an error only when nodes is nil so callers can distinguish a
// missing list from a cluster that genuinely has no GPU-labelled nodes.
func GetAllGpuTypesFromNodes(nodes *corev1.NodeList) (map[string]struct{}, error) {
	gpuTypes := make(map[string]struct{})
	if nodes == nil {
		return gpuTypes, errors.New("empty node list")
	}
	for _, n := range nodes.Items {
		if typeLabel, ok := n.Labels[NodeGPUTypeLabel]; ok && typeLabel != "" {
			gpuTypes[typeLabel] = struct{}{}
		}
	}
	return gpuTypes, nil
}

// GetClusterGpuTypes lists nodes from the running cluster and returns the
// unique GPU types reported via NodeGPUTypeLabel. Errors talking to the
// cluster surface to the caller; an empty (but non-nil) map with a nil error
// means the cluster has no GPU-labelled nodes.
func GetClusterGpuTypes(ctx context.Context) (map[string]struct{}, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	nodes, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return GetAllGpuTypesFromNodes(nodes)
}

// GetFirstClusterGpuType returns the lexicographically first GPU type
// reported by cluster nodes, or "" with a nil error when the cluster has no
// GPU types. Cluster-API errors are logged and surface to the caller so the
// caller can decide whether to proceed.
func GetFirstClusterGpuType(ctx context.Context) (string, error) {
	types, err := GetClusterGpuTypes(ctx)
	if err != nil {
		klog.Warningf("GetFirstClusterGpuType: cannot read cluster gpu types: %v", err)
		return "", err
	}
	if len(types) == 0 {
		return "", nil
	}
	keys := make([]string, 0, len(types))
	for k := range types {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys[0], nil
}
