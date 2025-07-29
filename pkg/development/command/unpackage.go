package command

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

type unpackageChart struct {
	baseDir string
}

func UnpackageChart() *unpackageChart {
	return &unpackageChart{baseDir: "/"}
}

func (c *unpackageChart) WithDir(dir string) *unpackageChart {
	c.baseDir = dir
	return c
}

func (c *unpackageChart) Run(path string) error {
	buf, err := os.ReadFile(path)
	if err != nil {
		klog.Errorf("failed to read file path=%s, err=%v", path, err)
		return err
	}

	zr, err := gzip.NewReader(bytes.NewBuffer(buf))
	if err != nil {
		klog.Errorf("failed to gunzip path=%s,err=%v", path, err)
		return err
	}

	tgz := tar.NewReader(zr)
	if err := os.MkdirAll(c.baseDir, 0775); err != nil {
		klog.Errorf("failed to mkdir path=%s, err=%v", c.baseDir, err)
		return err
	}

	for {
		header, err := tgz.Next()

		switch {
		case err == io.EOF:
			klog.Info("untar success, ", path)
			return nil
		case err != nil:
			klog.Errorf("failed to untar path=%s, err=%v", path, err)
			return err
		case header == nil:
			continue
		}

		dstFileOrDir := filepath.Join(c.baseDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if !existDir(dstFileOrDir) {
				if err := os.MkdirAll(dstFileOrDir, 0775); err != nil {
					klog.Errorf("failed to mkdir path=%v, err=%v", dstFileOrDir, err)
					return err
				}
			}
		case tar.TypeReg:
			file, err := os.OpenFile(dstFileOrDir, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				klog.Errorf("failed to open file path=%s, err=%v", dstFileOrDir, err)
				return err
			}
			defer file.Close()

			n, err := io.Copy(file, tgz)
			if err != nil {
				klog.Errorf("failed to copy file %v", err)
				return err
			}

			klog.Info("Extract file ", dstFileOrDir, ", size: ", n)
		}
	}
}
