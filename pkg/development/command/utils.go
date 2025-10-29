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
		klog.Errorf("failed to open file path=%s, err=%v", src, err)
		return err
	}

	file, err := f.Stat()
	if err != nil {
		klog.Errorf("failed to stat file path=%s, err=%v", src, err)
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("Source " + file.Name() + " is not a directory!")
	}

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		klog.Errorf("failed to mkdir path=%v,err=%v", dest, err)
		return err
	}

	files, err := os.ReadDir(src)
	if err != nil {
		klog.Errorf("failed to read dir path=%v,err=%v", src, err)
		return err
	}

	for _, f := range files {

		if f.IsDir() {

			err = copyDir(src+"/"+f.Name(), dest+"/"+f.Name())
			if err != nil {
				klog.Errorf("failed to copy dir from %s to %s, err=%v", src+"/"+f.Name(), dest+"/"+f.Name())
				return err
			}

		}

		if !f.IsDir() {

			content, err := os.ReadFile(src + "/" + f.Name())
			if err != nil {
				klog.Errorf("failed to read file path=%v, err=%v", src+"/"+f.Name(), err)
				return err

			}

			err = os.WriteFile(dest+"/"+f.Name(), content, 0755)
			if err != nil {
				klog.Errorf("failed to write file path=%s,err=%v", dest+"/"+f.Name(), err)
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
	klog.Infof("backup path=%s", orig)
	data, err := os.ReadFile(orig)
	if err != nil {
		klog.Errorf("failed to read origin file path=%s,err=%v", orig, err)
		return nil, err
	}
	err = os.MkdirAll(filepath.Dir(bak), 0755)
	if err != nil {
		klog.Errorf("failed to mkdir dir=%s, err=%v", filepath.Dir(bak), err)
		return nil, err
	}
	err = os.WriteFile(bak, data, 0644)
	if err != nil {
		klog.Errorf("failed to backup origin file %s, err=%v", bak, err)
		return nil, err
	}

	return func() {
		klog.Infof("restore path=%s", orig)
		err = os.Remove(orig)
		if err != nil {
			klog.Errorf("failed to remove path=%s", orig)
			return
		}

		err = os.Rename(bak, orig)
		if err != nil {
			klog.Errorf("failed to rename from %s to %s,err=%v", bak, orig, err)
		}

	}, nil
}
