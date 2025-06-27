package command

import (
	"context"
	"io"
	"os"
	"testing"

	"k8s.io/klog/v2"
)

func TestCreateApp(t *testing.T) {
	err := CreateApp().WithDir("/tmp").Run(context.Background(), &CreateConfig{Name: "testdev"})
	if err != nil {
		klog.Error(err)
		t.Fail()
	} else {
		t.Log("run CreateApp command success")
	}
}

func TestInstall(t *testing.T) {
	_, err := Install().Run(context.Background(), "newapp", "test", "0.0.1")
	if err != nil {
		klog.Error(err)
		t.Fail()
	} else {
		t.Log("run Install command success")
	}
}

func TestUpdateRepo(t *testing.T) {
	_, err := UpdateRepo().WithDir("/tmp").Run(context.Background(), "newapp", false)
	if err != nil {
		klog.Error(err)
		t.Fail()
	} else {
		t.Log("run UpdateRepo command success")
	}

}

func TestPackage(t *testing.T) {
	b, err := PackageChart().WithDir("/tmp").Run("newapp")
	if err != nil {
		klog.Error(err)
		t.Fail()
	} else {
		fileToWrite, err := os.OpenFile("/tmp/compress.tar.gzip", os.O_CREATE|os.O_RDWR, os.FileMode(0644))
		if err != nil {
			klog.Error(err)
			t.Fail()
			return
		}
		if _, err := io.Copy(fileToWrite, b); err != nil {
			klog.Error(err)
			t.Fail()
			return
		}

		t.Log("run packagechart command success")
	}
}

func TestUnpackage(t *testing.T) {
	err := UnpackageChart().WithDir("/tmp").Run("/tmp/dify.tgz")
	if err != nil {
		klog.Error(err)
		t.Fail()
	} else {
		t.Log("success untar")
	}
}

func TestCop(t *testing.T) {
	err := copyDir("/tmp/dify", "/tmp/dify1")
	if err != nil {
		klog.Error(err)
		t.Fail()
	} else {
		t.Log("success copy dir")
	}
}
