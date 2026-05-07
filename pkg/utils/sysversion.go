package utils

import (
	"context"

	"github.com/beclab/api/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

const terminusCRName = "terminus"

// GetSystemVersion returns the version reported by the cluster's Terminus CR
// (sys.bytetrade.io/v1alpha1, name "terminus"). It returns an empty string and
// the underlying error when the CR cannot be fetched.
func GetSystemVersion(ctx context.Context) (string, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("GetSystemVersion: get kube config %v", err)
		return "", err
	}
	c, err := versioned.NewForConfig(config)
	if err != nil {
		klog.Errorf("GetSystemVersion: new versioned client %v", err)
		return "", err
	}
	t, err := c.SysV1alpha1().Terminus().Get(ctx, terminusCRName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("GetSystemVersion: get terminus CR %v", err)
		return "", err
	}
	return t.Spec.Version, nil
}
