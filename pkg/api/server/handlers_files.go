package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/beclab/devbox/pkg/development/command"
	"github.com/beclab/devbox/pkg/files"
	"github.com/beclab/oachecker"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"k8s.io/klog/v2"
)

func (h *handlers) getFiles(ctx *fiber.Ctx) error {
	path := ctx.Params("*1")
	username := ctx.Locals("username").(string)

	file, err := files.NewFileInfo(files.FileOptions{
		Fs:         afero.NewBasePathFs(afero.NewOsFs(), BaseDir+"/"+username),
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

	pathParts := strings.SplitN(path, "/", 2)
	if len(pathParts) == 0 {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "Invalid path format",
		})
	}
	username := ctx.Locals("username").(string)
	appName := pathParts[0]
	file, err := WriteFileAndLint(ctx.Context(), username, path, appName, bytes.NewReader(content), command.Lint().WithDir(BaseDir).Run)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": file,
	})
}

func WriteFileAndLint(ctx context.Context, owner, originFilePath, name string, content io.Reader, lintFunc func(context.Context, string, string) error) (os.FileInfo, error) {
	exists := PathExists("/charts/tmp")
	if !exists {
		err := os.MkdirAll("/charts/tmp", 0755)
		if err != nil {
			return nil, err
		}
	}

	tempFile, err := os.CreateTemp("/charts/tmp", "bak-*"+filepath.Base(originFilePath))
	if err != nil {
		return nil, fmt.Errorf("create bak temp file failed %v", tempFile)
	}

	bakContent, err := os.ReadFile(filepath.Join(BaseDir, originFilePath))
	if err != nil {
		return nil, fmt.Errorf("read origin file %s failed %v", originFilePath, err)
	}

	_, err = tempFile.Write(bakContent)
	if err != nil {
		return nil, err
	}

	file, err := files.WriteFile(afero.NewBasePathFs(afero.NewOsFs(), BaseDir), originFilePath, content)
	if err != nil {
		return nil, err
	}

	if err = lintFunc(ctx, owner, name); err != nil {
		if restoreErr := os.Rename(tempFile.Name(), filepath.Join(BaseDir, owner, originFilePath)); restoreErr != nil {
			return nil, fmt.Errorf("lint failed: %v, and restore bak failed: %v", err, restoreErr)
		}
		return nil, fmt.Errorf("lint failed: %v", err)
	}
	if _, err = os.Stat(tempFile.Name()); err == nil {
		e := os.Remove(tempFile.Name())
		if e != nil {
			klog.Infof("remove temp file failed %v", e)
		}
	}

	return file, nil
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

func ManifestLint(content []byte) error {
	err := oachecker.CheckManifestFromContent(content)
	if err != nil {
		return err
	}
	err = oachecker.CheckManifestFromContent(content, oachecker.WithOwner("owner"),
		oachecker.WithAdmin("admin"))
	if err != nil {
		return err
	}
	return nil
}
