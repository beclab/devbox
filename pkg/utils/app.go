package utils

import (
	"path/filepath"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/loader"
	"github.com/kubernetes/kompose/pkg/transformer"
	"github.com/kubernetes/kompose/pkg/transformer/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
)

func Convert(opt kobject.ConvertOptions) ([]runtime.Object, error) {
	l, err := loader.GetLoader("compose")
	if err != nil {
		return nil, err
	}
	komposeObject := kobject.KomposeObject{
		ServiceConfigs: make(map[string]kobject.ServiceConfig),
	}
	komposeObject, err = l.LoadFile(opt.InputFiles, opt.Profiles, opt.NoInterpolate)
	if err != nil {
		return nil, err
	}
	komposeObject.Namespace = opt.Namespace
	workDir, err := transformer.GetComposeFileDir(opt.InputFiles)
	if err != nil {
		return nil, err
	}
	// convert env_file from absolute to relative path
	for _, service := range komposeObject.ServiceConfigs {
		if len(service.EnvFile) <= 0 {
			continue
		}
		for i, envFile := range service.EnvFile {
			if !filepath.IsAbs(envFile) {
				continue
			}

			relPath, err := filepath.Rel(workDir, envFile)
			if err != nil {
				return nil, err
			}

			service.EnvFile[i] = filepath.ToSlash(relPath)
		}
	}
	t := &kubernetes.Kubernetes{Opt: opt}
	objects, err := t.Transform(komposeObject, opt)
	if err != nil {
		return nil, err
	}
	err = kubernetes.PrintList(objects, opt)
	if err != nil {
		return nil, err
	}
	return objects, nil
}
