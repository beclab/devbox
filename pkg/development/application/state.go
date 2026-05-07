package application

import (
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Operate struct {
	AppName           string                  `json:"appName"`
	AppNamespace      string                  `json:"appNamespace"`
	AppOwner          string                  `json:"appOwner"`
	State             ApplicationManagerState `json:"state"`
	OpType            OpType                  `json:"opType"`
	Message           string                  `json:"message"`
	ResourceType      string                  `json:"resourceType"`
	CreationTimestamp metav1.Time             `json:"creationTimestamp"`
	Source            string                  `json:"source"`
}

type ApplicationManagerState = appv1.ApplicationManagerState

type OpType = appv1.OpType
