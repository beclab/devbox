package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/beclab/devbox/pkg/files"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"k8s.io/klog/v2"
)

func (h *handlers) getFiles(ctx *fiber.Ctx) error {
	path := ctx.Params("*1")

	file, err := files.NewFileInfo(files.FileOptions{
		Fs:         afero.NewBasePathFs(afero.NewOsFs(), BaseDir),
		Path:       path,
		Modify:     true,
		Expand:     true,
		ReadHeader: true,
		Checker:    &noCheck{},
		Content:    true,
	})
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Get files failed: %v", err),
		})
	}

	if file.IsDir {
		file.Listing.ApplySort()
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": file,
	})
}

func (h *handlers) saveFile(ctx *fiber.Ctx) error {
	path := ctx.Params("*1")
	content := ctx.Body()

	file, err := files.WriteFile(afero.NewBasePathFs(afero.NewOsFs(), BaseDir), path, bytes.NewReader(content))
	if err != nil {
		klog.Error("write file error, ", err, ", ", path)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Write file error : %v path: %s", err, path),
		})
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": file,
	})
}

func (h *handlers) resourcePostHandler(ctx *fiber.Ctx) error {
	path := ctx.Params("*1")
	klog.Infof("resourcePostHandler: %s", path)
	isDir := ctx.Query("file_type") == "dir"
	if isDir {
		klog.Infof("resourcePostHandler mkdir: %s", filepath.Join(BaseDir, path))
		err := os.MkdirAll(filepath.Join(BaseDir, path), 0755)
		if err != nil {
			return ctx.JSON(fiber.Map{
				"code":    errToStatus(err),
				"message": err.Error(),
			})
		}
		return ctx.JSON(fiber.Map{
			"code": http.StatusOK,
			"data": map[string]string{},
		})
	}
	klog.Infof("resource post override: %s", ctx.Query("override"))
	_, err := os.Stat(filepath.Join(BaseDir, path))
	if err == nil {
		if ctx.Query("override") != "true" {
			return ctx.JSON(fiber.Map{
				"code":    http.StatusConflict,
				"message": "File already exists",
			})
		}
	}
	file, err := files.WriteFile(afero.NewBasePathFs(afero.NewOsFs(), BaseDir), path, bytes.NewReader(ctx.Body()))
	if err != nil {
		klog.Error("write file error, ", err, ", ", path)
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Write file failed: %v path: %s", err, path),
		})
	}
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": file,
	})
}

func (h *handlers) resourceDeleteHandler(ctx *fiber.Ctx) error {
	path := ctx.Params("*1")
	if len(strings.Split(path, "/")) < 2 {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusForbidden,
			"message": "Permission denied",
		})
	}
	_, err := os.Stat(filepath.Join(BaseDir, path))
	if os.IsNotExist(err) {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": "No such file or directory",
		})
	}
	err = os.RemoveAll(filepath.Join(BaseDir, path))
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("Delete file failed: %v", err),
		})
	}
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]string{},
	})
}

func (h *handlers) resourcePatchHandler(ctx *fiber.Ctx) error {
	path := ctx.Params("*1")
	dst := ctx.Query("destination")
	action := ctx.Query("action")
	override := ctx.Query("override") == "true"

	if !override {
		if _, err := os.Stat(dst); err == nil {
			return ctx.JSON(fiber.Map{
				"code":    errToStatus(err),
				"message": err.Error(),
			})
		}
	}
	src := filepath.Join(BaseDir, path)
	dst = filepath.Join(BaseDir, dst)
	klog.Infof("src: %s", src)
	klog.Infof("dst: %s", dst)
	err := patchAction(action, src, dst)
	if err != nil {
		klog.Infof("patchAction error: %v", err)
		return ctx.JSON(fiber.Map{
			"code":    errToStatus(err),
			"message": err.Error(),
		})
	}
	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": map[string]string{},
	})
}

func patchAction(action, src, dst string) error {
	switch action {
	case "rename":
		return files.MoveFile(afero.NewOsFs(), src, dst)
	default:
		return fmt.Errorf("unsupported action %s: %w", action, ErrInvalidRequestParams)
	}
}

type noCheck struct {
}

func (*noCheck) Check(path string) bool { return true }
