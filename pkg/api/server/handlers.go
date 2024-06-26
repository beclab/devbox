package server

import (
	"github.com/beclab/devbox/pkg/store/db"
	"github.com/beclab/devbox/pkg/webhook"

	"k8s.io/client-go/rest"
)

type handlers struct {
	db         *db.DbOperator
	kubeConfig *rest.Config
}

type webhooks struct {
	webhook *webhook.Webhook
}
