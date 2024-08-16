package server

import (
	"os"
	"strconv"

	"github.com/beclab/devbox/pkg/store/db"
	"github.com/beclab/devbox/pkg/webhook"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type server struct {
	handlers *handlers
	webhooks *webhooks
}

var (
	BaseDir         string = "/tmp"
	defaultCertPath        = "/etc/certs/server.crt"
	defaultKeyPath         = "/etc/certs/server.key"
	tlsCertEnv             = "WEBHOOK_TLS_CERT"
	tlsKeyEnv              = "WEBHOOK_TLS_KEY"
)

func NewServer(db *db.DbOperator) *server {
	config := ctrl.GetConfigOrDie()
	webhook := &webhook.Webhook{
		KubeClient: kubernetes.NewForConfigOrDie(config),
		DB:         db,
	}
	utilruntime.Must(webhook.CreateOrUpdateDevContainerMutatingWebhook())
	utilruntime.Must(webhook.CreateOrUpdateImageManagerMutatingWebhook())

	return &server{
		handlers: &handlers{db: db, kubeConfig: config},
		webhooks: &webhooks{webhook: webhook},
	}
}

func (s *server) Start() {
	dir := os.Getenv("BASE_DIR")
	if dir != "" {
		BaseDir = dir
	}

	tlsCert, tlsKey := defaultCertPath, defaultKeyPath
	if os.Getenv(tlsCertEnv) != "" && os.Getenv(tlsKeyEnv) != "" {
		tlsCert, tlsKey = os.Getenv(tlsCertEnv), os.Getenv(tlsKeyEnv)
	}

	app := fiber.New()
	webhookServer := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	api := app.Group("api")

	// commands /api/command
	command := api.Group("command")
	command.Post("/create-app", s.handlers.createDevApp)
	command.Get("/list-app", s.handlers.listDevApps)
	command.Get("/apps/:name", s.handlers.getDevApp)
	command.Post("/update-app-repo", s.handlers.updateDevAppRepo)
	command.Post("/install-app", s.handlers.installDevApp)
	command.Get("/download-app-chart", s.handlers.downloadDevAppChart)
	command.Post("/open-application", s.handlers.openApplication)
	command.Post("/delete-app", s.handlers.deleteDevApp)
	command.Post("/upload-app-chart", s.handlers.uploadDevAppChart)
	command.Get("/lint-app-chart", s.handlers.lintDevAppChart)
	command.Post("/uninstall/:name", s.handlers.uninstall)
	command.Post("/upload-app-archive", s.handlers.createAppByArchive)

	// files /api/files
	files := api.Group("files")
	files.Get("/*", s.handlers.getFiles)
	files.Put("/*", s.handlers.saveFile)
	files.Post("/*", s.handlers.resourcePostHandler)
	files.Delete("/*", s.handlers.resourceDeleteHandler)
	files.Patch("/*", s.handlers.resourcePatchHandler)

	// front end api  /api
	api.Post("/bind-container", s.handlers.bindContainer)
	api.Post("/unbind-container", s.handlers.unbindContainer)
	api.Get("/list-app-containers", s.handlers.listAppContainersInChart)
	api.Get("/list-my-containers", s.handlers.listMyContainers)
	api.Get("/app-cfg", s.handlers.getAppConfig)
	api.Post("/app-cfg", s.handlers.updateAppConfig)

	api.Get("/app-state", s.handlers.getAppState)
	api.Get("/app-status", s.handlers.getAppStatus)
	api.Post("/apps/:name/cancel", s.handlers.cancel)
	api.Get("/dev-container/:name", s.handlers.getDevContainer)
	api.Delete("/dev-container/:name", s.handlers.delDevContainer)
	api.Patch("/dev-container/:name", s.handlers.updateDevContainer)

	api.Get("/dev-containers/:id", s.handlers.getDevContainer)
	// webhooks /webhook
	wh := webhookServer.Group("webhook")
	wh.Post("/devcontainer", s.webhooks.devcontainer)
	wh.Post("/imagemanager", s.webhooks.imageManager)

	klog.Info("dev box api server listening on 8088 ")

	go func() {
		err := webhookServer.ListenTLS(":"+strconv.Itoa(int(webhook.WebhookPort)), tlsCert, tlsKey)
		app.Shutdown()
		klog.Fatal(err)
	}()

	err := app.Listen(":8088")
	webhookServer.Shutdown()

	klog.Fatal(err)
}
