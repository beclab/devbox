package command

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/beclab/devbox/pkg/utils"

	"k8s.io/klog/v2"
)

type packageChart struct {
	baseDir string
	uername string
}

func PackageChart() *packageChart {
	return &packageChart{baseDir: "/"}
}

func (c *packageChart) WithDir(dir string) *packageChart {
	c.baseDir = dir
	return c
}
func (c *packageChart) WithUser(username string) *packageChart {
	c.uername = username
	return c
}

func (c *packageChart) Run(pathToPackage string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	realPath := filepath.Join(utils.GetUserBaseDir(c.uername), pathToPackage)

	err := c.compress(realPath, &buf)
	if err != nil {
		klog.Errorf("failed to compress chart path=%s, err=%v", realPath, err)
		return nil, err
	}

	return &buf, nil
}

func (c *packageChart) compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	// is file a folder?
	fi, err := os.Stat(src)
	if err != nil {
		klog.Errorf("failed to stat path=%s, err=%v", src, err)
		return err
	}
	mode := fi.Mode()
	if mode.IsRegular() {
		// get header
		header, err := tar.FileInfoHeader(fi, src)
		if err != nil {
			klog.Errorf("failed to get file info header %v", err)
			return err
		}
		// write header
		if err := tw.WriteHeader(header); err != nil {
			klog.Errorf("failed to write header %v", err)
			return err
		}
		// get content
		data, err := os.Open(src)
		if err != nil {
			klog.Errorf("failed to open path=%s, err=%v", src, err)
			return err
		}
		if _, err := io.Copy(tw, data); err != nil {
			klog.Errorf("failed to copy data %v", err)
			return err
		}
	} else if mode.IsDir() { // folder

		// walk through every file in the folder
		filepath.Walk(src, func(file string, fi os.FileInfo, e error) error {
			if e != nil {
				return e
			}
			// generate tar header
			header, err := tar.FileInfoHeader(fi, file)
			if err != nil {
				klog.Errorf("failed to generate tar header file=%s,err=%v", file, err)
				return err
			}

			// must provide real name
			// (see https://golang.org/src/archive/tar/common.go?#L626)
			// strip base_dir and username prefix to get relative path starting from app name
			relativePath := strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(file), c.baseDir), "/")
			// Remove username prefix (e.g., "olaresid/bbb/" -> "bbb/")
			pathParts := strings.Split(relativePath, "/")
			if len(pathParts) > 1 {
				header.Name = strings.Join(pathParts[1:], "/")
			} else {
				header.Name = relativePath
			}

			// write header
			if err := tw.WriteHeader(header); err != nil {
				klog.Errorf("failed to write header %v", err)
				return err
			}
			// if not a dir, write file content
			if !fi.IsDir() {
				data, err := os.Open(file)
				if err != nil {
					klog.Errorf("failed to open file=%s, err=%v", file, err)
					return err
				}
				if _, err := io.Copy(tw, data); err != nil {
					return err
				}
			}
			return nil
		})
	} else {
		return fmt.Errorf("error: file type not supported")
	}

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	//
	return nil
}
