package application

import (
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

type ApplicationManagerState string

var (
	Pending      ApplicationManagerState = "pending"
	Installing   ApplicationManagerState = "installing"
	Upgrading    ApplicationManagerState = "upgrading"
	Uninstalling ApplicationManagerState = "uninstalling"
	Canceled     ApplicationManagerState = "canceled"
	Failed       ApplicationManagerState = "failed"
	Completed    ApplicationManagerState = "completed"
	Suspend      ApplicationManagerState = "suspend"

	Processing ApplicationManagerState = "processing"
)

func (a ApplicationManagerState) String() string {
	return string(a)
}

type OpType string

var (
	Install    OpType = "install"
	Uninstall  OpType = "uninstall"
	Upgrade    OpType = "upgrade"
	SuspendApp OpType = "suspend"
	ResumeApp  OpType = "resume"
	Cancel     OpType = "cancel"
)
