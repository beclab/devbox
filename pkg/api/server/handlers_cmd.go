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
	"strings"
	"time"

	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/development/command"
	"github.com/beclab/devbox/pkg/development/container"
	"github.com/beclab/devbox/pkg/development/helm"
	"github.com/beclab/devbox/pkg/store/db"
	"github.com/beclab/devbox/pkg/store/db/model"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func (h *handlers) createDevApp(ctx *fiber.Ctx) error {
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
	err = command.CreateApp().WithDir(BaseDir).Run(ctx.Context(), &config)
	if err != nil {
		klog.Error("create app chart error, ", err, ", ", config)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Create chart failed: %v", err),
		})
	}

	output, err := command.CheckCfg().WithDir(BaseDir).Run(ctx.Context(), config.Name)
	if err != nil {

		klog.Error("check OlaresManifest.yaml error, ", err, ", ", config)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("OlaresManifest.yaml has error: %v", err),
		})
	}
	if len(output) > 0 {
		err = os.RemoveAll(filepath.Join(BaseDir, config.Name))
		if err != nil {
			klog.Errorf("remove dir %s error", config.Name)
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": output,
			})
		}
	}

	// create app in db
	appData := model.DevApp{
		AppName: config.Name,
		DevEnv:  config.DevEnv,
		AppType: db.CommunityApp,
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
	list := make([]*model.DevApp, 0)
	err := h.db.DB.Order("update_time desc").Find(&list).Error
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
	appName := ctx.Params("name")
	if len(appName) == 0 {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusConflict,
			"message": "Not a valid application name",
		})
	}

	var app *model.DevApp
	err := h.db.DB.First(&app).Error
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

	err = command.UpdateRepo().WithDir(BaseDir).Run(ctx.Context(), name, false)
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
	app := make(map[string]string)
	err := ctx.BodyParser(&app)
	if err != nil {
		klog.Error("parse install app data error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Parse install application data failed: %v", err),
		})
	}

	token := ctx.Cookies("auth_token")
	if token == "" {
		klog.Error("token is empty")
		return ctx.JSON(fiber.Map{
			"code":    http.StatusUnauthorized,
			"message": fmt.Sprintf("Auth token not found"),
		})
	}

	name, ok := app["name"]
	if !ok {
		klog.Error("app name is empty, ", app)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("App name is empty"),
		})
	}
	source := app["source"]
	devName := name + "-dev"
	devNamespace := devName + "-" + constants.Owner

	output, err := command.Lint().WithDir(BaseDir).Run(context.TODO(), name)
	if err != nil {
		return err
	}
	if len(output) > 0 {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": output,
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
		err = WaitForUninstall(devName, token, h.kubeConfig)
		if err != nil {
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Wait for uninstall failed: %v", err),
			})
		}
	}

	klog.Info("preinstall, create a labeled namespace for webhook")
	_, err = container.CreateOrUpdateDevNamespace(ctx.Context(), h.kubeConfig, devName)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Check namespace failed: %v", err),
		})
	}

	if source != "cli" {
		klog.Info("auto update repo")
		err = command.UpdateRepo().WithDir(BaseDir).Run(ctx.Context(), name, releaseNotExist)
		if err != nil {
			klog.Error("command upgraderepo error, ", err, ", ", name)
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Update repo failed: %v", err),
			})
		}
	}

	output, err = command.Install().Run(ctx.Context(), devName, token)
	if err != nil {
		klog.Error("command install error, ", err, ", ", name)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Install failed: %v", err),
		})
	}

	if len(output) > 0 {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": output,
		})
	}
	return ctx.JSON(fiber.Map{
		"code":    http.StatusOK,
		"message": "Install success",
	})
}

func (h *handlers) downloadDevAppChart(ctx *fiber.Ctx) error {
	app := ctx.Query("app")
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Application Not Found"),
		})
	}

	buf, err := command.PackageChart().WithDir(BaseDir).Run(app)
	if err != nil {
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

	var devApp *model.DevApp
	err = h.db.DB.Where("app_name = ?", name).First(&devApp).Error
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
	if err != nil {
		klog.Error("exec sql error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql Failed: %v", err),
		})
	}

	err = h.db.DB.Where("id = ?", devApp.ID).Delete(&model.DevApp{}).Error
	if err != nil {
		klog.Error("exec sql error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Exec sql Failed: %v", err),
		})
	}

	err = command.DeleteChart().WithDir(BaseDir).Run(name)
	if err != nil {
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

func (h *handlers) uploadDevAppChart(ctx *fiber.Ctx) error {
	app := ctx.FormValue("app")
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("Applcation Not Found"),
		})
	}

	// parse incomming chart tgz/zip file
	fileHeader, err := ctx.FormFile("chart")
	if err != nil {
		klog.Error("read file from request error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Read file frome request Failed: %v", err),
		})
	}

	uniqueId := strings.ReplaceAll(uuid.NewString(), "-", "")

	path := "/tmp/" + uniqueId + filepath.Ext(fileHeader.Filename)
	klog.Infof("path is: %s\n", path)
	err = ctx.SaveFile(fileHeader, path)
	if err != nil {
		klog.Error("save tmp file error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Save tmp file Failed: %v", err),
		})
	}

	// uncompress tgz/zip
	untarPath := filepath.Join("/tmp", uniqueId)
	err = UnArchive(path, untarPath)
	if err != nil {
		klog.Error("unpackage chart error, ", err, ", ", untarPath)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("UnArchive Failed: %v", err),
		})
	}

	output, err := command.Lint().WithDir(untarPath).Run(context.TODO(), app)
	if err != nil {
		klog.Error("check chart error, ", err)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Lint Failed: %v", err),
		})
	}
	if len(output) > 0 {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": output,
		})
	}

	// copy uploaded chart
	klog.Infof("upload dev Chart untarPath: %s, baseDir: %s, app: %s", untarPath, BaseDir, app)
	err = command.CopyApp().WithDir(BaseDir).Run(filepath.Join(untarPath, app), app)
	if err != nil {
		klog.Error("copy chart error, ", err, ", ")
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Copy Application failed: %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code":    http.StatusOK,
		"message": "Upload chart to Application success",
	})
}

func (h *handlers) lintDevAppChart(ctx *fiber.Ctx) error {
	app := ctx.Query("app")
	if app == "" {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("Application Not Found"),
		})
	}

	res, err := command.Lint().WithDir(BaseDir).Run(ctx.Context(), app)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Lint Failed: %v", err),
		})
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]string{"result": res},
	})
}

func (h *handlers) uninstall(ctx *fiber.Ctx) error {
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
	devName := name + "-dev"
	res, err := uninstall(devName, token)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Uninstall Failed: %v", err),
		})
	}

	klog.Infof("res: %#v", res.Data)
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]string{
			"uid": res.Data.Data.UID,
		},
	})
}

func (h *handlers) createAppByArchive(ctx *fiber.Ctx) error {
	override := ctx.Query("override") == "true"

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

	uniqueId := strings.ReplaceAll(uuid.NewString(), "-", "")
	err = UnArchive(filepath.Join("/tmp", file.Filename), filepath.Join("/tmp", uniqueId))
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("UnArchive failed: %v", err),
		})
	}

	cfg, err := readCfgFromFile(filepath.Join("/tmp", uniqueId))
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Read cfg frome file failed: %v", err),
		})
	}
	klog.Infof("readCfgFromFile cfg: %#v\n", cfg)
	chartDir := filepath.Dir(findAppCfgFile(filepath.Join("/tmp", uniqueId)))
	klog.Infof("chartDir: %s\n", chartDir)

	klog.Infof("WithDir: %s\n", filepath.Dir(chartDir))
	klog.Infof("chart Base : %s\n", filepath.Base(chartDir))

	output, err := command.Lint().WithDir(filepath.Dir(chartDir)).Run(context.TODO(), filepath.Base(chartDir))
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Lint failed: %v", err),
		})
	}
	klog.Infof("output: %s\n", output)
	if len(output) > 0 {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": output,
		})
	}
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
		}
		appID, err = InsertDevApp(&appData)
		if err != nil {
			return ctx.JSON(fiber.Map{
				"code":    http.StatusBadRequest,
				"message": fmt.Sprintf("Insert app failed: %v", err),
			})
		}
	}
	// copy chart to /charts
	//chartDir := filepath.Dir(findAppCfgFile(filepath.Join("/tmp", uniqueId)))
	err = command.CopyApp().WithDir(BaseDir).Run(filepath.Join("/tmp", uniqueId, cfg.Metadata.Name), cfg.Metadata.Name)
	if err != nil {
		e := h.db.DB.Where("id = ?", appID).Delete(&model.DevApp{}).Error
		if err != nil {
			klog.Error(e)
		}
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
			e := op.DB.Where("app_name = ?", app.AppName).Delete(&model.DevApp{}).Error
			if e != nil {
				klog.Warning("delete to rollback db error, ", err)
			}
		}
	}()
	var exists *model.DevApp
	err = op.DB.Where("app_name = ?", app.AppName).First(&exists).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("exec sql error, ", err)
		return appId, err
	}

	if err == nil {
		return appId, ErrAppIsExist
	}

	da := model.DevApp{
		AppName: app.AppName,
		AppType: app.AppType,
		DevEnv:  app.DevEnv,
	}
	err = op.DB.Create(&da).Error
	if err != nil {
		klog.Error("exec sql error, ", err)
		return appId, err
	}

	appId = int64(da.ID)
	if err != nil {
		klog.Error("get last insert id error, ", err)
		return appId, err
	}
	return appId, nil
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

func uninstall(name, token string) (data *SystemServerWrap, err error) {
	url := fmt.Sprintf("http://%s/system-server/v1alpha1/app/service.appstore/v1/UninstallDevApp", constants.SystemServer)
	accessToken, err := command.GetAccessToken()
	if err != nil {
		return data, err
	}

	client := resty.New().SetTimeout(5 * time.Second)
	resp, err := client.R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader("X-Authorization", token).
		SetHeader("X-Access-Token", accessToken).
		SetBody(map[string]interface{}{
			"name": name,
		}).Post(url)
	if err != nil {
		return data, err
	}
	klog.Info("resp.StatusCode: ", resp.StatusCode())
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

	code := data.Code
	if code != 0 {
		return nil, errors.New(data.Message)
	}

	return data, nil
}

func WaitForUninstall(name, token string, kubeConfig *rest.Config) error {
	_, err := uninstall(name, token)
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Error(err)
		return err
	}
	devNamespace := name + "-" + constants.Owner
	klog.Infof("wait for uninstall: %s", devNamespace)
	return wait.PollUntilContextTimeout(context.TODO(), time.Second, 3*time.Minute, true, func(ctx context.Context) (done bool, err error) {
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
