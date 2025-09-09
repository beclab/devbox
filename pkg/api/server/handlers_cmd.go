package server

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/beclab/devbox/pkg/development/command"
	"github.com/beclab/devbox/pkg/development/container"
	"github.com/beclab/devbox/pkg/development/helm"
	"github.com/beclab/devbox/pkg/store/db"
	"github.com/beclab/devbox/pkg/store/db/model"
	"github.com/beclab/devbox/pkg/utils"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var regxPattern = "^[a-zA-Z][a-zA-Z0-9 ._-]{0,30}$"

func (h *handlers) createDevApp(ctx *fiber.Ctx) error {

	username := ctx.Locals("username").(string)

	var config command.CreateConfig
	err := ctx.BodyParser(&config)
	if err != nil {
		klog.Error("parse create config error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Bad Request: %v", err),
		})
	}

	klog.Info("create app in db with config, ", config)
	if config.DevEnv == "" {
		config.DevEnv = "default"
	}

	// create app via command cli
	err = command.CreateApp().WithDir(BaseDir).Run(ctx.Context(), &config, username)
	if err != nil {
		klog.Error("create app chart error, ", err, ", ", config)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Create chart failed: %v", err),
		})
	}

	err = command.CheckCfg().WithDir(BaseDir).Run(ctx.Context(), username, config.Name)
	if err != nil {
		e := os.RemoveAll(filepath.Join(BaseDir, config.Name))
		if e != nil {
			klog.Errorf("remove dir %s error", config.Name)
		}
		klog.Error("check OlaresManifest.yaml error, ", err, ", ", config)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("OlaresManifest.yaml has error: %v", err),
		})
	}

	// create app in db
	appData := model.DevApp{
		AppName: config.Name,
		DevEnv:  config.DevEnv,
		AppType: db.CommunityApp,
		Owner:   username,
	}
	appId, err := InsertDevApp(&appData)
	if err != nil {
		if errors.Is(err, ErrAppIsExist) {
			klog.Error("app already exists, ", appData.AppName)
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": "Application already exists",
			})
		}
		return err
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]interface{}{
			"appId": appId,
		},
	})
}

func (h *handlers) listDevApps(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)
	list := make([]*model.DevApp, 0)
	err := h.db.DB.Where("owner=?", username).Order("update_time desc").Find(&list).Error
	if err != nil {
		klog.Error("exec sql error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql failed: %v", err),
		})
	}

	appid := func(name string) string {
		hash := md5.Sum([]byte(name + "-dev"))
		hashString := hex.EncodeToString(hash[:])
		return hashString[:8]
	}

	host := ctx.Request().Header.PeekBytes([]byte("Host"))
	zone := strings.Join(strings.Split(string(host), ".")[1:], ".")
	for i, l := range list {
		appId := appid(l.AppName)
		list[i].Chart = "/" + l.AppName       // TODO: save app chart dir into db
		list[i].Entrance = appId + "." + zone // TODO:
		list[i].AppID = appId
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": list,
	})
}

func (h *handlers) getDevApp(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)
	appName := ctx.Params("name")
	if len(appName) == 0 {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusConflict,
			"message": "Not a valid application name",
		})
	}

	var app *model.DevApp
	err := h.db.DB.Where("owner = ?", username).First(&app).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("exec sql error, ")
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql failed: %v", err),
		})
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.JSON(fiber.Map{
			"code": http.StatusOK,
			"data": map[string]string{},
		})
	}
	devApp := app
	appid := func(name string) string {
		hash := md5.Sum([]byte(name + "-dev"))
		hashString := hex.EncodeToString(hash[:])
		return hashString[:8]
	}

	host := ctx.Request().Header.PeekBytes([]byte("Host"))
	zone := strings.Join(strings.Split(string(host), ".")[1:], ".")

	devApp.AppID = appid(appName)
	devApp.Chart = "/" + appName
	devApp.Entrance = devApp.AppID + "." + zone

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": devApp,
	})
}

func (h *handlers) updateDevAppRepo(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)
	app := make(map[string]string)
	err := ctx.BodyParser(&app)
	if err != nil {
		klog.Error("parse update app data error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Parse update app data failed: %v", err),
		})
	}

	name, ok := app["name"]
	if !ok {
		klog.Error("app name is empty, ", app)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Application name is empty"),
		})
	}

	_, err = command.UpdateRepo().WithDir(BaseDir).Run(ctx.Context(), username, name, false)
	if err != nil {
		klog.Error("command upgraderepo error, ", err, ", ", name)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Update repo failed: %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code":    http.StatusOK,
		"message": "update success",
	})
}

func (h *handlers) installDevApp(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)
	var err error
	app := make(map[string]string)
	err = ctx.BodyParser(&app)
	if err != nil {
		klog.Error("parse install app data error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Parse install application data failed: %v", err),
		})
	}

	token := ctx.Locals("auth_token").(string)

	name, ok := app["name"]
	if !ok {
		klog.Error("app name is empty, ", app)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("App name is empty"),
		})
	}
	defer func() {
		if err != nil {
			err = UpdateDevAppState(username, name, abnormal)
			if err != nil {
				klog.Errorf("update app state to abnormal err %v", err)
			}
		}
	}()
	err = UpdateDevAppState(username, name, deploying)
	if err != nil {
		klog.Errorf("failed to update dev app state name=%s,err=%v", name, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("update app state err %v", err),
		})
	}

	source := app["source"]
	devName := fmt.Sprintf("%s-%s", name, "dev")
	devNamespace := fmt.Sprintf("%s-%s", devName, username)

	err = command.Lint().WithDir(BaseDir).Run(context.TODO(), username, name)
	if err != nil {
		klog.Errorf("failed to lint app=%s, err=%v", name, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
	}

	klog.Info("uninstall prev app or not")
	var releaseNotExist bool

	err = helm.GetRelease(h.kubeConfig, devNamespace, devName)
	if err != nil {
		if !errors.Is(err, driver.ErrReleaseNotFound) {
			return err
		}
		releaseNotExist = true
	}
	if !releaseNotExist {
		err = WaitForUninstall(username, devName, token, h.kubeConfig)
		if err != nil {
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Wait for uninstall failed: %v", err),
			})
		}
	}

	klog.Info("preinstall, create a labeled namespace for webhook")
	_, err = container.CreateOrUpdateDevNamespace(ctx.Context(), h.kubeConfig, username, devName)
	if err != nil {
		klog.Errorf("failed to check namespace %v", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Check namespace failed: %v", err),
		})
	}
	version := "0.0.1"
	if source != "cli" {
		klog.Info("auto update repo")
		version, err = command.UpdateRepo().WithDir(BaseDir).Run(ctx.Context(), username, name, releaseNotExist)
		if err != nil {
			klog.Error("command upgraderepo error, ", err, ", ", name)
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Update repo failed: %v", err),
			})
		}
	}

	_, err = command.Install().Run(ctx.Context(), username, devName, token, version)

	if err != nil {
		klog.Error("command install error, ", err, ", ", name)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Install failed: %v", err),
		})
	}

	err = UpdateDevAppState(username, name, deployed)
	if err != nil {
		klog.Errorf("failed to update app=%s state to deployed %v", name, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("update app state to deployed err %v", err),
		})
	}
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]string{
			"namespace": devNamespace,
		},
		"message": "Install success",
	})
}

func (h *handlers) downloadDevAppChart(ctx *fiber.Ctx) error {
	app := ctx.Query("app")
	username := ctx.Locals("username").(string)
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Application Not Found"),
		})
	}

	buf, err := command.PackageChart().WithDir(BaseDir).WithUser(username).Run(app)
	if err != nil {
		klog.Errorf("failed to package app=%s chart %v", app, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Package chart Failed: %v", err),
		})
	}

	ctx.Response().Header.SetCanonical([]byte(fiber.HeaderContentDisposition), []byte(`attachment; filename="`+app+`.tgz"`))
	return ctx.SendStream(buf)
}

func (h *handlers) openApplication(ctx *fiber.Ctx) error {
	app := make(map[string]string)
	err := ctx.BodyParser(&app)
	if err != nil {
		klog.Error("parse update app data error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Parse app data failed: %v", err),
		})
	}

	appid, ok := app["appid"]
	if !ok {
		klog.Error("app id is empty, ", app)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Appid is empty"),
		})
	}

	path, ok := app["path"]
	if !ok {
		klog.Error("path is empty, ", app)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Path is empty"),
		})
	}

	//httpposturl := fmt.Sprintf("http://%s/legacy/v1alpha1/api.intent/v1/server/intent/send", os.Getenv("OS_SYSTEM_SERVER"))
	httpposturl := fmt.Sprintf("http://edge-desktop.user-space-%s/server/intent/send", os.Getenv("OWNER"))

	fmt.Println("HTTP JSON POST URL:", httpposturl)

	var jsonData = []byte(`{
			"action": "view",
			"category": "launcher",
			"data": {
				"appid": "` + appid + `",
				"path": "` + path + `"
			}
		}`)

	request, error := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
	if error != nil {
		klog.Error("create intent request error, ", error)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Create intent request failed: %v", error),
		})
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		klog.Error("request intent error, ", error)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Request intent failed: %v", error),
		})
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return ctx.JSON(fiber.Map{
			"code":    response.StatusCode,
			"message": fmt.Sprintf("Request intent failed: %v", error),
		})
	}

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := io.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))

	return ctx.JSON(fiber.Map{
		"code":    http.StatusOK,
		"message": "Open Application success",
	})
}

func (h *handlers) deleteDevApp(ctx *fiber.Ctx) error {
	app := make(map[string]string)
	err := ctx.BodyParser(&app)
	if err != nil {
		klog.Error("parse delete app data error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Parse delete app data error %v", err),
		})
	}

	name, ok := app["name"]
	if !ok {
		klog.Error("app name is empty, ", app)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Application name is empty"),
		})
	}
	username := ctx.Locals("username").(string)

	var devApp *model.DevApp
	err = h.db.DB.Where("owner = ?", username).Where("app_name = ?", name).First(&devApp).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("exec sql error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec Sql Failed: %v", err),
		})
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("app not found in db, ", devApp.AppName)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("Application Not Found"),
		})
	}

	// unbind app's containers
	err = h.db.DB.Where("app_id = ?", devApp.ID).Delete(&model.DevAppContainers{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("exec sql error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql Failed: %v", err),
		})
	}

	klog.Infof("devApp.ID: %v", devApp.ID)
	err = h.db.DB.Where("id = ?", devApp.ID).Delete(&model.DevApp{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("exec sql error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql Failed: %v", err),
		})
	}
	err = h.db.DB.Where("name = ?", devApp.AppName).Delete(&model.DevContainers{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("exec sql error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql Failed: %v", err),
		})
	}

	err = command.DeleteChart().WithDir(BaseDir).WithUser(username).Run(name)
	if err != nil {
		klog.Errorf("failed to delete chart %s, err=%v", name, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Delete Chart Failed: %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code":    http.StatusOK,
		"message": "Delete Application success",
	})
}

//func (h *handlers) uploadDevAppChart(ctx *fiber.Ctx) error {
//	app := ctx.FormValue("app")
//	if app == "" {
//		return ctx.JSON(fiber.Map{
//			"code":    http.StatusNotFound,
//			"message": fmt.Sprintf("Applcation Not Found"),
//		})
//	}
//
//	username := ctx.Locals("username").(string)
//
//	// parse incomming chart tgz/zip file
//	fileHeader, err := ctx.FormFile("chart")
//	if err != nil {
//		klog.Error("read file from request error, ", err)
//		return ctx.JSON(fiber.Map{
//			"code":    http.StatusBadRequest,
//			"message": fmt.Sprintf("Read file frome request Failed: %v", err),
//		})
//	}
//
//	uniqueId := strings.ReplaceAll(uuid.NewString(), "-", "")
//
//	path := "/tmp/" + uniqueId + filepath.Ext(fileHeader.Filename)
//	klog.Infof("path is: %s\n", path)
//	err = ctx.SaveFile(fileHeader, path)
//	if err != nil {
//		klog.Error("save tmp file error, ", err)
//		return ctx.JSON(fiber.Map{
//			"code":    http.StatusBadRequest,
//			"message": fmt.Sprintf("Save tmp file Failed: %v", err),
//		})
//	}
//
//	// uncompress tgz/zip
//	untarPath := filepath.Join("/tmp", uniqueId)
//	err = UnArchive(path, untarPath)
//	if err != nil {
//		klog.Error("unpackage chart error, ", err, ", ", untarPath)
//		return ctx.JSON(fiber.Map{
//			"code":    http.StatusBadRequest,
//			"message": fmt.Sprintf("UnArchive Failed: %v", err),
//		})
//	}
//
//	err = command.Lint().WithDir(untarPath).Run(context.TODO(), username, app)
//	if err != nil {
//		klog.Error("check chart error, ", err)
//		return ctx.JSON(fiber.Map{
//			"code":    http.StatusBadRequest,
//			"message": fmt.Sprintf("Lint Failed: %v", err),
//		})
//	}
//
//	// copy uploaded chart
//	klog.Infof("upload dev Chart untarPath: %s, baseDir: %s, app: %s", untarPath, BaseDir, app)
//	err = command.CopyApp().WithDir(BaseDir).Run(filepath.Join(untarPath, app), app)
//	if err != nil {
//		klog.Error("copy chart error, ", err, ", ")
//		return ctx.JSON(fiber.Map{
//			"code":    http.StatusBadRequest,
//			"message": fmt.Sprintf("Copy Application failed: %v", err),
//		})
//	}
//
//	return ctx.JSON(fiber.Map{
//		"code":    http.StatusOK,
//		"message": "Upload chart to Application success",
//	})
//}

func (h *handlers) lintDevAppChart(ctx *fiber.Ctx) error {
	app := ctx.Query("app")
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("Application Not Found"),
		})
	}
	username := ctx.Locals("username").(string)

	err := command.Lint().WithDir(BaseDir).Run(ctx.Context(), username, app)
	if err != nil {
		klog.Errorf("failed to lint app %s, err=%v", app, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Lint Failed: %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]string{"result": ""},
	})
}

func (h *handlers) uninstall(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)
	name := ctx.Params("name")
	if name == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": "Application Not Found",
		})
	}
	klog.Info("uninstall name: ", name)
	token := ctx.Cookies("auth_token")
	if token == "" {
		klog.Error("token is empty")
		return ctx.JSON(fiber.Map{
			"code":    http.StatusUnauthorized,
			"message": "Auth token not found",
		})
	}
	devName := fmt.Sprintf("%s-%s", name, "dev")
	res, err := uninstall(devName, token, username)
	if err != nil {
		klog.Errorf("failed to uninstall %s, err=%v", devName, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Uninstall Failed: %v", err),
		})
	}
	err = UpdateDevAppState(username, name, undeploy)
	if err != nil {
		klog.Errorf("update dev app state to undeploy err %v", err)
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": res,
	})
}

func (h *handlers) createAppByArchive(ctx *fiber.Ctx) error {
	override := ctx.Query("override") == "true"
	username := ctx.Locals("username").(string)

	file, err := ctx.FormFile("chart")
	if err != nil {
		klog.Error("read file from request error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Read file from request error: %v", err),
		})
	}
	os.RemoveAll(filepath.Join("/tmp", file.Filename))
	err = ctx.SaveFile(file, filepath.Join("/tmp", file.Filename))
	if err != nil {
		klog.Error("save tmp file error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Save tmp file error: %v", err),
		})
	}

	//uniqueId := strings.ReplaceAll(uuid.NewString(), "-", "")
	err = UnArchive(filepath.Join("/tmp", file.Filename), filepath.Join("/tmp", username))
	if err != nil {
		klog.Errorf("failed to unarchive file %s, err=%v", filepath.Join("/tmp", file.Filename), err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("UnArchive failed: %v", err),
		})
	}

	cfg, err := readCfgFromFile(username, filepath.Join("/tmp", username))
	if err != nil {
		klog.Errorf("failed to read cfg from file %v", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Read cfg frome file failed: %v", err),
		})
	}
	klog.Infof("readCfgFromFile cfg: %#v\n", cfg)
	chartDir := filepath.Dir(findAppCfgFile(filepath.Join("/tmp", username)))
	klog.Infof("chartDir: %s\n", chartDir)

	klog.Infof("WithDir: %s\n", filepath.Dir(chartDir))
	klog.Infof("chart Base : %s\n", filepath.Base(chartDir))

	err = command.Lint().WithDir("/tmp").Run(context.TODO(), username, filepath.Base(chartDir))
	if err != nil {
		klog.Errorf("lint failed %v", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Lint failed: %v", err),
		})
	}
	//klog.Infof("output: %s\n", output)
	//if len(output) > 0 {
	//	return ctx.JSON(fiber.Map{
	//		"code":    http.StatusBadRequest,
	//		"message": output,
	//	})
	//}
	var exists *model.DevApp
	err = h.db.DB.Where("app_name = ?", cfg.Metadata.Name).First(&exists).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("exec sql error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql error: %v", err),
		})
	}

	var appID int64
	if err == nil {
		if !override {
			return ctx.JSON(fiber.Map{
				"code":    http.StatusConflict,
				"message": fmt.Sprintf("app %s already exists", cfg.Metadata.Name),
			})
		}
		return ctx.JSON(fiber.Map{
			"code": http.StatusOK,
			"data": map[string]interface{}{
				"appId": appID,
			},
		})
	} else {
		// insert db
		klog.Info("insert into db devapp")
		appData := model.DevApp{
			AppName: cfg.Metadata.Name,
			DevEnv:  "default",
			AppType: db.CommunityApp,
			State:   undeploy,
			Owner:   username,
		}
		appID, err = InsertDevApp(&appData)
		if err != nil {
			klog.Errorf("failed to insert app %s,err=%v", cfg.Metadata.Name, err)
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Insert app failed: %v", err),
			})
		}
	}
	// copy chart to /charts
	//chartDir := filepath.Dir(findAppCfgFile(filepath.Join("/tmp", uniqueId)))
	err = command.CopyApp().WithDir(BaseDir).WithUser(username).Run(filepath.Join("/tmp", username, cfg.Metadata.Name), cfg.Metadata.Name)
	if err != nil {
		e := h.db.DB.Where("id = ?", appID).Delete(&model.DevApp{}).Error
		if err != nil {
			klog.Error(e)
		}
		klog.Errorf("failed to copy app dir %v", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Copy app withdir failed: %v", err),
		})
	}
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]interface{}{
			"appId": appID,
		},
	})
}

func InsertDevApp(app *model.DevApp) (appId int64, err error) {
	op := db.NewDbOperator()
	// if err rollback db
	defer func() {
		if err != nil {
			e := op.DB.Where("owner = ?", app.Owner).Where("app_name = ?", app.AppName).Delete(&model.DevApp{}).Error
			if e != nil {
				klog.Warning("delete to rollback db error, ", err)
			}
		}
	}()
	var exists *model.DevApp
	err = op.DB.Where("owner = ?", app.Owner).Where("app_name = ?", app.AppName).First(&exists).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("exec sql error, ", err)
		return appId, err
	}

	if err == nil {
		return appId, ErrAppIsExist
	}

	err = op.DB.Create(app).Error
	if err != nil {
		klog.Error("exec sql error, ", err)
		return appId, err
	}

	appId = int64(app.ID)
	if err != nil {
		klog.Error("get last insert id error, ", err)
		return appId, err
	}
	return appId, nil
}

func UpdateDevApp(owner, name string, updates map[string]interface{}) (appId int64, err error) {
	op := db.NewDbOperator()
	var exists *model.DevApp
	err = op.DB.Where("owner = ?", owner).Where("app_name = ?", name).First(&exists).Error
	if err != nil {
		return 0, err
	}

	err = op.DB.Model(&exists).Updates(updates).Error
	if err != nil {
		klog.Errorf("update dev_app err %v", err)
		return 0, err
	}
	appId = int64(exists.ID)
	return appId, nil
}

func UpdateDevAppState(owner, name string, state string) error {
	updates := map[string]interface{}{
		"state": state,
	}
	_, err := UpdateDevApp(owner, name, updates)
	if err != nil {
		return err
	}
	return nil
}

type InstallationResponseData struct {
	UID string `json:"uid"`
}
type Response struct {
	Code int32 `json:"code"`
}

type InstallationResponse struct {
	Response
	Data InstallationResponseData `json:"data"`
}

type SystemServerWrap struct {
	Code    int32                `json:"code"`
	Message string               `json:"message"`
	Data    InstallationResponse `json:"data"`
}

func uninstall(name, token, owner string) (data map[string]interface{}, err error) {
	uninstalled, err := checkIfAppIsUninstalled(name, token, owner)
	if err != nil {
		return data, err
	}
	if uninstalled {
		return data, nil
	}
	url := fmt.Sprintf("http://app-service.os-framework:6755/app-service/v1/apps/%s/uninstall", name)

	client := resty.New().SetTimeout(5 * time.Second)
	resp, err := client.R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader("X-Authorization", token).
		SetHeader("X-Bfl-User", owner).
		Post(url)
	if err != nil {
		klog.Errorf("failed to send request to uninstall app %s, err=%v", name, err)
		return data, err
	}
	klog.Info("request uninstall resp.StatusCode: ", resp.StatusCode())
	if resp.StatusCode() != http.StatusOK {
		dump, e := httputil.DumpRequest(resp.Request.RawRequest, true)
		if e == nil {
			klog.Error("request error, ", string(dump))
		}
		return nil, errors.New(string(resp.Body()))
	}
	klog.Info("resp.Body: ", string(resp.Body()))
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func WaitForUninstall(owner, name, token string, kubeConfig *rest.Config) error {
	_, err := uninstall(name, token, owner)
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Error(err)
		return err
	}

	devNamespace := fmt.Sprintf("%s-%s", name, owner)
	klog.Infof("wait for uninstall: %s", devNamespace)
	return wait.PollUntilContextTimeout(context.TODO(), time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		if err != nil {
			return false, err
		}
		_, err = client.CoreV1().Namespaces().Get(ctx, devNamespace, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})
}

type App struct {
	Title string `json:"title"`
}

func (h *handlers) createApp(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	var app App
	err := ctx.BodyParser(&app)
	if err != nil {
		klog.Error("parse app info error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Bad Request: %v", err),
		})
	}
	appName := removeSpecialCharsMap(strings.ToLower(app.Title))

	if app.Title == "tmp" || appName == "tmp" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("name %s is reserved word", app.Title),
		})
	}

	regex := regexp.MustCompile(regxPattern)
	if !regex.MatchString(app.Title) {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Bad Request: this field must conform to the pattern ^[a-zA-Z][a-zA-Z0-9 ._-]{0,29}$"),
		})
	}
	err = h.db.DB.Where("owner = ?", username).Where("title = ?", app.Title).First(&model.DevApp{}).Error
	if err == nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("create app failed, app ID %s already exists", appName),
		})
	}
	err = h.db.DB.Where("owner = ?", username).Where("app_name = ?", appName).First(&model.DevApp{}).Error
	if err == nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("create app failed, app ID %s already exists", appName),
		})
	}

	appData := model.DevApp{
		Title:   app.Title,
		AppName: appName,
		AppType: db.CommunityApp,
		State:   empty,
		Owner:   username,
	}
	appId, err := InsertDevApp(&appData)
	if err != nil {
		klog.Errorf("create app err %v", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("create app err %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]interface{}{
			"appId": appId,
		},
	})
}

func (h *handlers) fillApp(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)
	name := ctx.Params("name")
	var cfg command.CreateWithOneDockerConfig

	err := ctx.BodyParser(&cfg)
	if err != nil {
		klog.Errorf("parse create config err %v", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Bad Request: %v", err),
		})
	}

	if errs := command.ValidateStruct(cfg); len(errs) > 0 {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Bad Request: %v", errs),
		})
	}
	at := command.AppTemplate{}
	at.WithDockerCfg(&cfg).WithDockerDeployment(&cfg).
		WithDockerService(&cfg).WithDockerChartMetadata(&cfg).WithDockerOwner(&cfg)
	err = at.WriteDockerFile(&cfg, utils.GetAppPath(username, cfg.Name))
	if err != nil {
		klog.Errorf("write docker file err %v", err)
		e := os.RemoveAll(filepath.Join(BaseDir, name))
		if e != nil {
			klog.Errorf("remove dir %s err %v", name, e)
		}
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("create app err %v", err),
		})
	}

	updates := map[string]interface{}{
		"app_type": db.CommunityApp,
		"dev_env":  "default",
		"state":    undeploy,
	}
	appId, err := UpdateDevApp(username, cfg.Name, updates)
	if err != nil {
		klog.Errorf("failed to update dev app %s, err=%v", cfg.Name, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("update app err %v", err),
		})
	}
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]interface{}{
			"appId": appId,
		},
	})
}

func (h *handlers) appState(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	var app *model.DevApp

	username := ctx.Locals("username").(string)
	op := db.NewDbOperator()
	err := op.DB.Where("owner = ?", username).Where("app_name = ?", name).First(&app).Error
	if err != nil {
		klog.Errorf("get app name=%s err %v", name, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("get state err %v", err),
		})
	}
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]interface{}{
			"state": app.State,
		},
	})
}

func (h *handlers) fillAppWithExample(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)
	name := ctx.Params("name")

	var app App
	err := ctx.BodyParser(&app)
	if err != nil {
		klog.Errorf("failed to parse body %v", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Bad Request: %v", err),
		})
	}

	err = command.CreateAppWithHelloWorldConfig(username, name, app.Title)
	if err != nil {
		klog.Errorf("write docker file err %v", err)
		e := os.RemoveAll(filepath.Join(BaseDir, username, name))
		if e != nil {
			klog.Errorf("remove dir %s err %v", name, e)
		}
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("create app err %v", err),
		})
	}

	updates := map[string]interface{}{
		"app_type": db.CommunityApp,
		"dev_env":  "default",
		"state":    undeploy,
	}

	appId, err := UpdateDevApp(username, name, updates)
	if err != nil {
		klog.Errorf("failed to update app %s, err=%v", name, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("update app err %v", err),
		})
	}
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]interface{}{
			"appId": appId,
		},
	})
}

type BindData struct {
	ContainerId      *int
	AppName          string
	AppId            int64
	PodSelector      string
	ContainerName    string
	DevEnv           *string
	DevContainerName string
	Image            string
}

func (h *handlers) fillAppWithDevContainer(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)
	name := ctx.Params("name")
	var cfg command.CreateDevContainerConfig

	err := ctx.BodyParser(&cfg)
	if err != nil {
		klog.Errorf("parse create dev container config err %v", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Bad Request: %v", err),
		})
	}
	if errs := command.ValidateStruct(cfg); len(errs) > 0 {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Bad Request: %v", errs),
		})
	}
	if cfg.RequiredMemory != "" {
		memoryQuantity, _ := resource.ParseQuantity(cfg.RequiredMemory)

		minMemory, _ := resource.ParseQuantity("256Mi")
		if memoryQuantity.Cmp(minMemory) < 0 {
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": "RequiredMemory must be at least 256Mi",
			})
		}
	}

	err = command.CreateAppWithDevConfig(&cfg, username, name)
	if err != nil {
		klog.Errorf("write dev docker file err %v", err)
		e := os.RemoveAll(utils.GetAppPath(username, name))
		if e != nil {
			klog.Errorf("remove dir %s err %v", name, e)
		}
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("create app err %v", err),
		})
	}

	updates := map[string]interface{}{
		"app_type": db.CommunityApp,
		"dev_env":  cfg.DevEnv,
		"state":    undeploy,
	}

	appId, err := UpdateDevApp(username, name, updates)
	if err != nil {
		klog.Errorf("failed to update dev app %w,err=%v", name, err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("update app err %v", err),
		})
	}

	containers, err := GetAppContainersInChart(username, name)
	if err != nil || len(containers) == 0 {
		klog.Errorf("failed to get app containers in chart err=%v, len(containers)=%d", err, len(containers))
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("get bind containers err %v", err),
		})
	}

	bindData := &BindData{
		AppId:            appId,
		AppName:          name,
		PodSelector:      containers[0].PodSelector,
		ContainerName:    containers[0].ContainerName,
		DevEnv:           &cfg.DevEnv,
		DevContainerName: name,
		Image:            containers[0].Image,
	}
	err = BindContainer(bindData)
	if err != nil {
		klog.Errorf("failed to bind container app=%s,err=%v", name, err)
		e := h.db.DB.Where("app_id = ?", appId).Delete(&model.DevAppContainers{}).Error
		if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
			klog.Errorf("delete devAppContainer app_id=%d err %v", appId, e)
		}
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("bind container err %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]interface{}{
			"appId": appId,
		},
	})
}

func checkIfAppIsUninstalled(name, token, owner string) (bool, error) {
	url := fmt.Sprintf("http://app-service.os-framework:6755/app-service/v1/apps/%s/status", name)
	data := make(map[string]interface{})

	client := resty.New().SetTimeout(5 * time.Second)
	resp, err := client.R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader("X-Authorization", token).
		SetHeader("X-Bfl-User", owner).
		Get(url)
	if err != nil {
		klog.Errorf("failed to send request to get app status %s, err=%v", name, err)
		return false, err
	}
	klog.Info("request app %s status resp.StatusCode: %d", name, resp.StatusCode())
	if resp.StatusCode() == http.StatusNotFound {
		return true, nil
	}
	if resp.StatusCode() != http.StatusOK {
		dump, e := httputil.DumpRequest(resp.Request.RawRequest, true)
		if e == nil {
			klog.Error("request error, ", string(dump))
		}
		return false, errors.New(string(resp.Body()))
	}
	klog.Info("resp.Body: ", string(resp.Body()))
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return false, err
	}
	appStatus, ok := data["status"]
	if !ok {
		return false, fmt.Errorf("status filed not found")
	}
	statusMap, ok := appStatus.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("status is not a map")
	}
	state, ok := statusMap["state"].(string)
	if !ok {
		return false, fmt.Errorf("state is not a string")
	}
	if state != "uninstalled" {
		return false, nil
	}

	return true, nil
}
