//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"

	"github.com/cpssd-students/paranoid/pkg/libpfs"
)

func createTestDir(t *testing.T, name string) {
	os.RemoveAll(path.Join(os.TempDir(), name))
	err := os.Mkdir(path.Join(os.TempDir(), name), 0777)
	if err != nil {
		if os.IsExist(err) {
			cmd := exec.Command("fusermount", "-z", "-u", path.Join(os.TempDir(), name))
			_ = cmd.Run()
			os.RemoveAll(path.Join(os.TempDir(), name))
			err = os.Mkdir(path.Join(os.TempDir(), name), 0777)
			if err != nil {
				t.Fatal("Error creating directory", err)
			}
		} else {
			t.Fatal("Error creating directory", err)
		}
	}
}

func removeTestDir(name string) {
	time.Sleep(1 * time.Second)
	os.RemoveAll(path.Join(os.TempDir(), name))
}

func TestKillSignal(t *testing.T) {
	createTestDir(t, "testksMountpoint")
	defer removeTestDir("testksMountpoint")
	createTestDir(t, "testksDirectory")
	defer removeTestDir("testksDirectory")

	discovery := exec.Command("discovery-server", "--port=10102", "-state=false")
	err := discovery.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := discovery.Process.Kill()
		if err != nil {
			t.Error("Failed to kill discovery server,", err)
		}
	}()

	_, err = libpfs.InitCommand(path.Join(os.TempDir(), "testksDirectory"))
	if err != nil {
		t.Fatal(err)
	}

	attributes := &globals.FileSystemAttributes{
		Encrypted:    true,
		KeyGenerated: false,
		NetworkOff:   false,
	}

	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		t.Fatal("unable to save file system attributes to file:", err)
	}

	if err = ioutil.WriteFile(
		path.Join(os.TempDir(), "testksDirectory", "meta", "attributes"),
		attributesJSON,
		0600,
	); err != nil {
		t.Fatal("unable to save file system attributes to file:", err)
	}

	pfsd := exec.Command(
		"pfsd",
		path.Join(os.TempDir(), "testksDirectory"),
		path.Join(os.TempDir(), "testksMountpoint"),
		"localhost",
		"10102",
		"testPool",
		"",
	)
	pfsd.Stderr = os.Stderr

	err = pfsd.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pfsd.Process.Kill() }()
	defer func() {
		cmd := exec.Command("fusermount", "-z", "-u", path.Join(os.TempDir(), "testksMountPoint"))
		_ = cmd.Run()
	}()

	time.Sleep(10 * time.Second)

	pidPath := path.Join(os.TempDir(), "testksDirectory", "meta", "pfsd.pid")
	if _, err := os.Stat(pidPath); err == nil {
		pidByte, err := ioutil.ReadFile(pidPath)
		if err != nil {
			t.Fatal("Can't read pid file", err)
		}
		pid, err := strconv.Atoi(string(pidByte))
		if err != nil {
			t.Fatal("Incorrect pid information", err)
		}
		err = syscall.Kill(pid, syscall.SIGTERM)
		if err != nil {
			t.Fatal("Can not kill PFSD,", err)
		}

		done := make(chan bool, 1)
		go func() {
			_ = pfsd.Wait()
			done <- true
		}()

		select {
		case <-time.After(10 * time.Second):
			t.Fatal("pfsd did not finish within 10 seconds")
		case <-done:
			break
		}
	} else {
		t.Fatal("Could not read pid file:", err)
	}
}
