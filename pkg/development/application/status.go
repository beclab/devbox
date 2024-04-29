package application

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Status struct {
	Name              string      `json:"name"`
	AppID             string      `json:"appID"`
	Namespace         string      `json:"namespace"`
	CreationTimestamp metav1.Time `json:"creationTimestamp"`
	Source            string      `json:"source"`
	AppStatus         AppStatus   `json:"status"`
}

type AppStatus struct {
	// the state of the application: draft, submitted, passed, rejected, suspended, active
	State      string       `json:"state,omitempty"`
	UpdateTime *metav1.Time `json:"updateTime"`
	StatusTime *metav1.Time `json:"statusTime"`
}
