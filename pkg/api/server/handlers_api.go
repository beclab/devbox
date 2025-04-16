package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/development/application"
	"github.com/beclab/devbox/pkg/development/command"
	"github.com/beclab/devbox/pkg/development/container"
	"github.com/beclab/devbox/pkg/development/helm"
	"github.com/beclab/devbox/pkg/store/db"
	"github.com/beclab/devbox/pkg/store/db/model"
	"github.com/beclab/devbox/pkg/utils"
	"github.com/beclab/oachecker"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (h *handlers) getAppConfig(ctx *fiber.Ctx) error {
	app := ctx.Query("app")
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("Application not found"),
		})
	}

	path := getAppPath(app)
	appCfgPath := filepath.Join(path, constants.AppCfgFileName)
	data, err := os.ReadFile(appCfgPath)
	if err != nil {
		klog.Error("read app cfg error, ", err, ", ", app, ", ", appCfgPath)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Read OlaresManifest.yaml failed: %v", err),
		})
	}

	//appcfg, err := oachecker.GetAppConfiguration()
	//
	//var appcfg application.AppConfiguration
	//err = yaml.Unmarshal(data, &appcfg)

	appcfg, err := utils.GetAppConfig(data)
	if err != nil {
		klog.Error("parse app cfg error, ", err)
		klog.Error(string(data))
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Parse OlaresManifest.yaml failed: %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": appcfg,
	})
}

func (h *handlers) updateAppConfig(ctx *fiber.Ctx) error {
	app := ctx.Query("app")
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("Application not found"),
		})
	}

	path := getAppPath(app)
	appCfgPath := filepath.Join(path, constants.AppCfgFileName)

	var appcfg oachecker.AppConfiguration
	err := ctx.BodyParser(&appcfg)
	if err != nil {
		klog.Error("read app cfg post data error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("OlaresManifest.yaml has errors: %v", err),
		})
	}
	data, err := yaml.Marshal(&appcfg)
	if err != nil {
		klog.Error("parse post app cfg error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("OlaresManifest.yaml has errors: %v", err),
		})
	}
	appCfg := filepath.Join(path, constants.AppCfgFileName)
	uniquePath := uuid.NewString()
	appCfgBak := filepath.Join("/tmp", uniquePath, "OlaresManifest.yaml.bak")
	chartDeferFunc, err := command.BackupAndRestoreFile(appCfg, appCfgBak)
	if err != nil {
		chartDeferFunc()
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Backup and restore file error: %v", err),
		})
	}

	err = os.WriteFile(appCfgPath, data, 0644)
	if err != nil {
		klog.Error("save app cfg to file error, ", err, ", ", appCfgPath)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Save OlaresManifest.yaml error: %v", err),
		})
	}
	err = command.CheckCfg().WithDir(BaseDir).Run(ctx.Context(), app)
	if err != nil {
		klog.Error("check app cfg error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("OlaresManifest.yaml has errors: %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code":    http.StatusOK,
		"message": "Update Successes",
	})

}

func (h *handlers) listMyContainers(ctx *fiber.Ctx) error {
	sql := `select a.*, b.pod_selector, b.app_id, b.container_name, c.app_name
	from dev_containers a
	left join dev_app_containers b on a.id = b.container_id
	left join dev_apps c on b.app_id = c.id
	order by a.create_time desc`
	list := make([]*model.DevContainerInfo, 0)

	err := h.db.DB.Raw(sql).Scan(&list).Error
	if err != nil {
		klog.Error("exec sql error, ", err, ", ", sql)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql failed: %v", err),
		})
	}

	unbind := ctx.Query("unbind") == "true"
	ret := make([]*model.DevContainerInfo, 0, len(list))

	for i, c := range list {
		ret = append(ret, list[i])
		if c.AppID != nil {

			// filter binding dev container
			if unbind {
				ret = append(ret[:i], ret[i+1:]...)
				continue
			}
			// container is bind into an app, check the container running status
			// ignore error, cause state is not a critical message
			state, devPort, _ := container.GetContainerStatus(ctx.Context(), h.kubeConfig, c)
			c.State = &state

			port := "/proxy/" + devPort + "/"
			c.DevPath = &port
			appcfg, err := readAppInfo(filepath.Join(BaseDir, *c.AppName, constants.AppCfgFileName))
			if err != nil {
				klog.Error("readCfgFromFile error, ", err)
				continue
			}

			list[i].Icon = &appcfg.Metadata.Icon
		}
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": ret,
	})

}

func (h *handlers) bindContainer(ctx *fiber.Ctx) error {
	var postData struct {
		ContainerId      *int    `json:"containerId,omitempty"`
		AppId            int     `json:"appId"`
		PodSelector      string  `json:"podSelector"`
		ContainerName    string  `json:"containerName"`
		Image            string  `json:"image"`
		DevEnv           *string `json:"devEnv,omitempty"`
		DevContainerName string  `json:"devContainerName"`
	}
	err := ctx.BodyParser(&postData)
	if err != nil {
		klog.Error("request body error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Request body error: %v", err),
		})
	}
	var containerId int
	if postData.ContainerId == nil {
		// create a new dev container
		if postData.DevEnv == nil {
			err = errors.New("unknown dev-env to create a dev container")
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Unkown dev-env to create a dev container"),
			})
		}

		devContainer := model.DevContainers{
			DevEnv: *postData.DevEnv,
			Name:   postData.DevContainerName,
		}
		err = h.db.DB.Where("name = ?", devContainer.Name).First(&model.DevContainers{}).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err == nil {
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("devcontainer %s already exists", devContainer.Name),
			})
		}

		err = h.db.DB.Create(&devContainer).Error
		if err != nil {
			klog.Error("exec sql error, ", err)
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Exec sql failed: %v", err),
			})
		}

		containerId = int(devContainer.ID)
		if err != nil {
			klog.Error("get last insert id error, ", err)
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Get last insert id failed: %v", err),
			})
		}

	} else {
		containerId = *postData.ContainerId

		// container can be bind to just one app
		var existsContainers *model.DevAppContainers
		err = h.db.DB.Where("container_id = ?", containerId).First(&existsContainers).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			klog.Error("exec sql error, ", err)
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Exec sql failed: %v", err),
			})
		}

		if err == nil {
			klog.Error("container is binding to another app")
			//return errors.New("container is binding to another app")
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Container is binding to an another app"),
			})
		}
	}

	appContainer := model.DevAppContainers{
		AppID:         uint(postData.AppId),
		ContainerID:   uint(containerId),
		PodSelector:   postData.PodSelector,
		ContainerName: postData.ContainerName,
		Image:         postData.Image,
	}

	err = h.db.DB.Create(&appContainer).Error
	if err != nil {
		klog.Error("exec sql error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql failed: %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code":    http.StatusOK,
		"message": "Bind Container Successes",
	})

}

func (h *handlers) unbindContainer(ctx *fiber.Ctx) error {
	var postData struct {
		ContainerId   *int    `json:"containerId"`
		AppId         *int    `json:"appId"`
		PodSelector   *string `json:"podSelector"`
		ContainerName *string `json:"containerName"`
	}

	err := ctx.BodyParser(&postData)
	if err != nil {
		klog.Error("request body error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Requst body error: %v", err),
		})
	}

	appContainer := model.DevAppContainers{
		AppID:         uint(*postData.AppId),
		ContainerID:   uint(*postData.ContainerId),
		PodSelector:   *postData.PodSelector,
		ContainerName: *postData.ContainerName,
	}

	err = h.db.DB.Where("container_id = ?", appContainer.ContainerID).
		Where("pod_selector = ?", appContainer.PodSelector).
		Where("container_name = ?", appContainer.ContainerName).
		Where("app_id = ?", appContainer.AppID).
		Delete(&appContainer).Error
	if err != nil {
		klog.Error("exec sql error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql failed: %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code":    http.StatusOK,
		"message": "Unbind container successes",
	})

}

func (h *handlers) listAppContainersInChart(ctx *fiber.Ctx) error {
	app := ctx.Query("app")
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("Application Not Found"),
		})
	}

	appName := fmt.Sprintf("%s-dev", app)
	testNamespace := fmt.Sprintf("%s-%s", appName, constants.Owner)

	// mock vals
	values := make(map[string]interface{})
	values["bfl"] = map[string]interface{}{
		"username": "bfl-username",
	}
	values["user"] = map[string]interface{}{
		"zone": "user-zone",
	}
	values["schedule"] = map[string]interface{}{
		"nodeName": "node",
	}
	values["userspace"] = map[string]interface{}{
		"appCache": "appcache",
		"userData": "userspace/Home",
	}
	values["os"] = map[string]interface{}{
		"appKey":    "appKey",
		"appSecret": "appSecret",
	}
	values["domain"] = map[string]interface{}{}
	values["dep"] = map[string]interface{}{}
	values["postgres"] = map[string]interface{}{
		"username":  "username",
		"databases": map[string]interface{}{},
		"password":  "password",
	}
	values["redis"] = map[string]interface{}{
		"username":  "username",
		"databases": map[string]interface{}{},
		"password":  "password",
	}
	values["mongodb"] = map[string]interface{}{
		"username":  "username",
		"databases": map[string]interface{}{},
		"password":  "password",
	}
	values["zinc"] = map[string]interface{}{
		"username": "username",
		"indexes":  map[string]interface{}{},
		"password": "password",
	}
	values["svcs"] = map[string]interface{}{}
	values["cluster"] = map[string]interface{}{}
	values["GPU"] = map[string]interface{}{
		"Type": "nvidia",
		"Cuda": os.Getenv("CUDA_VERSION"),
	}

	values["gpu"] = "nvidia"

	path := getAppPath(app)
	appCfgPath := filepath.Join(path, constants.AppCfgFileName)
	data, err := os.ReadFile(appCfgPath)
	if err != nil {
		klog.Error("read app cfg error, ", err, ", ", app, ", ", appCfgPath)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Read OlaresManifest.yaml failed: %v", err),
		})
	}

	appcfg, err := utils.GetAppConfig(data)

	if err != nil {
		klog.Error("parse app cfg error, ", err)
		klog.Error(string(data))
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Parse OlaresManifest.yaml failed: %v", err),
		})
	}

	entries := make(map[string]interface{})
	for _, e := range appcfg.Entrances {
		entries[e.Name] = "dryrun"
	}
	values["domain"] = entries

	manifest, err := helm.DryRun(ctx.Context(), h.kubeConfig, testNamespace, appName, getAppPath(app), values)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Dry run failed: %v", err),
		})
	}

	resources, err := helm.DecodeManifest(manifest)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Decode manifest failed: %v", err),
		})
	}

	var da *model.DevApp
	err = h.db.DB.Where("app_name = ?", app).First(&da).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("app %s can not found", app),
		})
	}

	containers := helm.FindContainers(resources)
	client, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		klog.Error("get kubernetes client error, ", err)
		return err
	}

	for i := range containers {
		containers[i].AppID = pointer.Int(int(da.ID))
		if container.IsSysAppDevImage(containers[i].Image) {
			containers[i].DevPath = pointer.String("/proxy/3000/")
			userspace := "user-space-" + constants.Owner
			pods, err := client.CoreV1().Pods(userspace).List(ctx.Context(), metav1.ListOptions{
				LabelSelector: containers[i].PodSelector,
			})

			if err != nil {
				klog.Error("get pods status error, ", err, ", ", containers[i].PodSelector)
			} else {
				if len(pods.Items) > 0 {
					if pods.Items[0].Status.Phase == "Running" {
						containers[i].State = pointer.String(string(pods.Items[0].Status.Phase))
					}
				} else {
					klog.Warning("pods not found, ", containers[i].PodSelector)
				}
			}
		}

		var dac *model.DevAppContainers
		err = h.db.DB.Where("app_id = ?", da.ID).Where("container_name", containers[i].ContainerName).First(&dac).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		var dc *model.DevContainers
		err = h.db.DB.Where("id = ?", dac.ContainerID).First(&dc).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			containers[i].DevContainerName = dc.Name
		}
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": containers,
	})
}

func (h *handlers) getAppState(ctx *fiber.Ctx) error {
	app := ctx.Query("app")
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("Application Not Found"),
		})
	}

	appName := fmt.Sprintf("%s-dev", app)

	token := ctx.Cookies("auth_token", "")
	if token == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusUnauthorized,
			"message": fmt.Sprintf("Auth token not found"),
		})
	}

	url := "http://app-service.os-system:6755/app-service/v1/apps/" + appName + "/operate"
	client := resty.New().SetTimeout(2 * time.Second)

	resp, err := client.R().SetDebug(true).
		SetHeader("X-Authorization", token).
		SetResult(&application.Operate{}).
		Get(url)

	if err != nil {
		klog.Error("get app operate status error, ", err, ", ", appName)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Get operator status failed: %v", err),
		})
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("get app operate status response error, ", resp.StatusCode(), ", ", string(resp.Body()), ", ", appName)
		if resp.StatusCode() == http.StatusNotFound {
			return ctx.JSON(fiber.Map{
				"code": http.StatusOK,
				"data": resp.Result(),
			})
		}

		return ctx.JSON(fiber.Map{
			"code":    resp.StatusCode(),
			"message": string(resp.Body()),
		})
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": resp.Result(),
	})
}

func (h *handlers) getAppStatus(ctx *fiber.Ctx) error {
	app := ctx.Query("app")
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("Application Not Found"),
		})
	}

	appName := fmt.Sprintf("%s-dev", app)

	token := ctx.Cookies("auth_token", "")
	if token == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusUnauthorized,
			"message": fmt.Sprintf("Auth token not found"),
		})
	}

	url := "http://app-service.os-system:6755/app-service/v1/apps/" + appName + "/status"
	client := resty.New().SetTimeout(2 * time.Second)

	resp, err := client.R().SetDebug(true).
		SetHeader("X-Authorization", token).
		SetResult(&application.Status{}).
		Get(url)

	if err != nil {
		klog.Error("get app operate status error, ", err, ", ", appName)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Get app status failed: %v", err),
		})
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("get app status response error, ", resp.StatusCode(), ", ", string(resp.Body()), ", ", appName)
		if resp.StatusCode() == http.StatusNotFound {
			return ctx.JSON(fiber.Map{
				"code": http.StatusOK,
				"data": resp.Result(),
			})
		}

		return ctx.JSON(fiber.Map{
			"code":    resp.StatusCode(),
			"message": string(resp.Body()),
		})
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": resp.Result(),
	})
}

func (h *handlers) cancel(ctx *fiber.Ctx) error {
	app := ctx.Params("name")
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("Application Not Found"),
		})
	}

	appName := fmt.Sprintf("%s-dev", app)

	token := ctx.Cookies("auth_token", "")
	if token == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusUnauthorized,
			"message": fmt.Sprintf("Auth token not found"),
		})
	}

	url := "http://app-service.os-system:6755/app-service/v1/apps/" + appName + "/cancel"
	client := resty.New().SetTimeout(2 * time.Second)

	resp, err := client.R().SetDebug(true).
		SetHeader("X-Authorization", token).
		SetHeader("Accept", "*/*").
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetResult(&InstallationResponse{}).
		Post(url)

	if err != nil {
		klog.Error("cancel app error, ", err, ", ", appName)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Cancel app failed: %v", err),
		})
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("cancel app response error, ", resp.StatusCode(), ", ", string(resp.Body()), ", ", appName)

		return ctx.JSON(fiber.Map{
			"code":    resp.StatusCode(),
			"message": string(resp.Body()),
		})
	}
	ret := resp.Result().(*InstallationResponse)
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]string{
			"uid": ret.Data.UID,
		},
	})
}

func (h *handlers) getDevContainer(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	if name == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "Not a valid dev container name",
		})
	}
	var dc *model.DevContainers
	err := h.db.DB.Where("name = ?", name).First(&dc).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.JSON(fiber.Map{
			"code": http.StatusOK,
			"data": map[string]string{},
		})
	}
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": dc,
	})
}

func (h *handlers) delDevContainer(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	if name == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "Not a valid dev container id",
		})
	}

	// checkout is under binding

	var dc *model.DevContainers
	err := h.db.DB.Where("name = ?", name).First(&dc).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	klog.Infof("get devcontainer %v", err)
	if err == nil {
		var dac *model.DevAppContainers
		e := h.db.DB.Where("container_id = ?", dc.ID).First(&dac).Error
		if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Can not delete devcontainer %s since it under binding", name),
			})
		}
		if e == nil {
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Can not delete devcontainer %s since it under binding", name),
			})
		} else {
			e := h.db.DB.Where("name = ?", name).Delete(&dc).Error
			if e != nil {
				klog.Error("delete error, ", e)
			}
		}
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]string{},
	})
}

func (h *handlers) updateDevContainer(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	if name == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "Not a valid dev container name",
		})
	}
	app := make(map[string]string)
	err := ctx.BodyParser(&app)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Parse body failed: %v", err),
		})
	}

	newName := app["devContainerName"]
	if newName == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "Not a valid dev container name",
		})
	}

	err = h.db.DB.Model(&model.DevContainers{}).Where("name = ?", name).Update("name", newName).Error
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Update dev conainter failed: %v", err),
		})
	}
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]string{},
	})
}

func GetAppContainersInChart(app string) ([]*helm.ContainerInfo, error) {

	appName := fmt.Sprintf("%s-dev", app)
	testNamespace := fmt.Sprintf("%s-%s", appName, constants.Owner)

	// mock vals
	values := make(map[string]interface{})
	values["bfl"] = map[string]interface{}{
		"username": "bfl-username",
	}
	values["user"] = map[string]interface{}{
		"zone": "user-zone",
	}
	values["schedule"] = map[string]interface{}{
		"nodeName": "node",
	}
	values["userspace"] = map[string]interface{}{
		"appCache": "appcache",
		"userData": "userspace/Home",
	}
	values["os"] = map[string]interface{}{
		"appKey":    "appKey",
		"appSecret": "appSecret",
	}
	values["domain"] = map[string]interface{}{}
	values["dep"] = map[string]interface{}{}
	values["postgres"] = map[string]interface{}{
		"username":  "username",
		"databases": map[string]interface{}{},
		"password":  "password",
	}
	values["redis"] = map[string]interface{}{
		"username":  "username",
		"databases": map[string]interface{}{},
		"password":  "password",
	}
	values["mongodb"] = map[string]interface{}{
		"username":  "username",
		"databases": map[string]interface{}{},
		"password":  "password",
	}
	values["zinc"] = map[string]interface{}{
		"username": "username",
		"indexes":  map[string]interface{}{},
		"password": "password",
	}
	values["svcs"] = map[string]interface{}{}
	values["cluster"] = map[string]interface{}{}
	values["GPU"] = map[string]interface{}{
		"Type": "nvidia",
		"Cuda": os.Getenv("CUDA_VERSION"),
	}

	values["gpu"] = "nvidia"

	path := getAppPath(app)
	appCfgPath := filepath.Join(path, constants.AppCfgFileName)
	data, err := os.ReadFile(appCfgPath)
	if err != nil {
		klog.Error("read app cfg error, ", err, ", ", app, ", ", appCfgPath)
		return nil, err
	}

	appcfg, err := utils.GetAppConfig(data)

	if err != nil {
		klog.Error("parse app cfg error, ", err)
		klog.Error(string(data))
		return nil, err
	}

	entries := make(map[string]interface{})
	for _, e := range appcfg.Entrances {
		entries[e.Name] = "dryrun"
	}
	values["domain"] = entries
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	manifest, err := helm.DryRun(context.TODO(), kubeConfig, testNamespace, appName, getAppPath(app), values)
	if err != nil {
		return nil, err
	}

	resources, err := helm.DecodeManifest(manifest)
	if err != nil {
		return nil, err
	}
	op := db.NewDbOperator()

	var da *model.DevApp
	err = op.DB.Where("app_name = ?", app).First(&da).Error
	if err != nil {
		klog.Errorf("GetAppContainersInchar: app_name:%s,err:%v", app, err)
		return nil, err
	}

	containers := helm.FindContainers(resources)
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Error("get kubernetes client error, ", err)
		return nil, err
	}

	for i := range containers {
		containers[i].AppID = pointer.Int(int(da.ID))
		if container.IsSysAppDevImage(containers[i].Image) {
			containers[i].DevPath = pointer.String("/proxy/3000/")
			userspace := "user-space-" + constants.Owner
			pods, err := client.CoreV1().Pods(userspace).List(context.TODO(), metav1.ListOptions{
				LabelSelector: containers[i].PodSelector,
			})

			if err != nil {
				klog.Error("get pods status error, ", err, ", ", containers[i].PodSelector)
			} else {
				if len(pods.Items) > 0 {
					if pods.Items[0].Status.Phase == "Running" {
						containers[i].State = pointer.String(string(pods.Items[0].Status.Phase))
					}
				} else {
					klog.Warning("pods not found, ", containers[i].PodSelector)
				}
			}
		}

		var dac *model.DevAppContainers
		err = op.DB.Where("app_id = ?", da.ID).Where("container_name", containers[i].ContainerName).First(&dac).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		var dc *model.DevContainers
		err = op.DB.Where("id = ?", dac.ContainerID).First(&dc).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err == nil {
			containers[i].DevContainerName = dc.Name
		}
	}

	return containers, nil
}

func BindContainer(data *BindData) error {
	op := db.NewDbOperator()
	var containerId int
	if data.ContainerId == nil {
		// create a new dev container
		if data.DevEnv == nil {
			err := errors.New("unknown dev-env to create a dev container")
			return err
		}

		devContainer := model.DevContainers{
			DevEnv: *data.DevEnv,
			Name:   data.DevContainerName,
		}
		err := op.DB.Where("name = ?", devContainer.Name).First(&model.DevContainers{}).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err == nil {
			return fmt.Errorf("devcontainer %s already exists", devContainer.Name)
		}

		err = op.DB.Create(&devContainer).Error
		if err != nil {
			klog.Error("exec sql error, ", err)
			return err
		}

		containerId = int(devContainer.ID)
		if err != nil {
			klog.Error("get last insert id error, ", err)
			return err
		}

	} else {
		containerId = *data.ContainerId

		// container can be bind to just one app
		var existsContainers *model.DevAppContainers
		err := op.DB.Where("container_id = ?", containerId).First(&existsContainers).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			klog.Error("exec sql error, ", err)
			return err
		}

		if err == nil {
			return nil
		}
	}

	appContainer := model.DevAppContainers{
		AppID:         uint(data.AppId),
		ContainerID:   uint(containerId),
		PodSelector:   data.PodSelector,
		ContainerName: data.ContainerName,
		Image:         data.Image,
	}

	err := op.DB.Create(&appContainer).Error
	if err != nil {
		klog.Error("exec sql error, ", err)
		return err
	}

	return nil
}
