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

	"k8s.io/klog/v2"
)

type packageChart struct {
	baseDir string
}

func PackageChart() *packageChart {
	return &packageChart{baseDir: "/"}
}

func (c *packageChart) WithDir(dir string) *packageChart {
	c.baseDir = dir
	return c
}

func (c *packageChart) Run(pathToPackage string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	realPath := filepath.Join(c.baseDir, pathToPackage)

	err := c.compress(realPath, &buf)
	if err != nil {
		klog.Error("compress chart error, ", err, ", ", pathToPackage, ", ", realPath)
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
		return err
	}
	mode := fi.Mode()
	if mode.IsRegular() {
		// get header
		header, err := tar.FileInfoHeader(fi, src)
		if err != nil {
			return err
		}
		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// get content
		data, err := os.Open(src)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tw, data); err != nil {
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
				return err
			}

			// must provide real name
			// (see https://golang.org/src/archive/tar/common.go?#L626)
			// strip base_dir
			header.Name = strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(file), c.baseDir), "/")

			// write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			// if not a dir, write file content
			if !fi.IsDir() {
				data, err := os.Open(file)
				if err != nil {
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
