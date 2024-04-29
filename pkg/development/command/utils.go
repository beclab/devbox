package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

func existDir(dirname string) bool {
	fi, err := os.Stat(dirname)
	return (err == nil || os.IsExist(err)) && fi.IsDir()
}

func copyDir(src string, dest string) error {
	destDirToken := filepath.SplitList(dest)
	srcDirToken := filepath.SplitList(src)

	if filepath.Join(destDirToken[:len(srcDirToken)]...) == src {
		return fmt.Errorf("cannot copy a folder into the folder itself")
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}

	file, err := f.Stat()
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("Source " + file.Name() + " is not a directory!")
	}

	err = os.Mkdir(dest, 0755)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, f := range files {

		if f.IsDir() {

			err = copyDir(src+"/"+f.Name(), dest+"/"+f.Name())
			if err != nil {
				return err
			}

		}

		if !f.IsDir() {

			content, err := os.ReadFile(src + "/" + f.Name())
			if err != nil {
				return err

			}

			err = os.WriteFile(dest+"/"+f.Name(), content, 0755)
			if err != nil {
				return err

			}

		}

	}

	return nil
}

func AggregateErrs(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		var errStr string
		for _, e := range errs {
			errStr += e.Error() + "\t"
		}
		return errors.New(errStr[:len(errStr)-1])
	}
}

func BackupAndRestoreFile(orig, bak string) (func(), error) {
	klog.Info("backup ", orig)
	data, err := os.ReadFile(orig)
	if err != nil {
		klog.Error("read origin file error, ", err, ", ", orig)
		return nil, err
	}
	err = os.MkdirAll(filepath.Dir(bak), 0755)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(bak, data, 0644)
	if err != nil {
		klog.Error("backup origin file error, ", err, ", ", bak)
		return nil, err
	}

	return func() {
		klog.Info("restore ", orig)
		err = os.Remove(orig)
		if err != nil {
			klog.Error(err)
			return
		}

		err = os.Rename(bak, orig)
		if err != nil {
			klog.Error(err)
		}

	}, nil
}
