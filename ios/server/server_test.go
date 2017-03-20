// +build !linux

package server

import (
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"testing"

	"time"

	"os/exec"

	"github.com/golang/glog"
	"github.com/heidi-ann/ios/client/kvclient"
	"github.com/heidi-ann/ios/config"
	"github.com/stretchr/testify/assert"
)

//check recovery
func TestIosServerRestart(t *testing.T) {
	//Create temp directory
	dir, err := ioutil.TempDir("", "IosRecoveryTests")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	//parse configuration
	configPath := os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/config/testConfigs/example.conf"
	config := config.ParseServerConfig(configPath)

	// start server
	var filepath = os.Getenv("GOPATH") + "/bin/ios"
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		filepath := os.Getenv("GOPATH") + "/bin/ios.exe"
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			t.Fatal("Cannot find ios executable")
		}
	}
	cmd := exec.Command(filepath, "-id", "0", "-config", configPath, "-disk", dir)
	err = cmd.Start()
	if err != nil {
		t.Fatal("Error starting server process. Error: ", err.Error())
	}

	time.Sleep(1 * time.Second)
	//verify files were created in directory
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	filenames := make([]string, len(files))
	for i, file := range files {
		filenames[i] = file.Name()
	}
	assert.EqualValues(t, []string{"persistent_data_0.temp", "persistent_log_0.temp", "persistent_snapshot_0.temp"}, filenames)

	tmpfile, err := ioutil.TempFile(dir, "latency")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	//make some basic requests
	client := kvclient.StartKvClient(-1, tmpfile.Name(), config.Clients.Address, 1*time.Second)
	client.Update("A", "1")
	assert.Equal(t, "1", client.Get("A"))
	assert.Equal(t, "A, 1\n", client.Print())

	//kill server
	err = cmd.Process.Signal(syscall.SIGKILL)
	if err != nil {
		t.Fatal("Error killing server process. Error: ", err.Error())
	}
	time.Sleep(1 * time.Second)

	//we expect any commands from the client to now not complete
	getChan := make(chan string, 1)
	go func() { getChan <- client.Get("A") }()
	select {
	case <-getChan:
		t.Fatal("Server did not shut down")
	case <-time.After(1 * time.Second):
		glog.Info("Server shut down succesfully")
	}

	//restart server
	cmd = exec.Command(filepath, "-id", "0", "-config", configPath, "-disk", dir)
	err = cmd.Start()
	if err != nil {
		t.Fatal("Error starting server process. Error: ", err.Error())
	}
	time.Sleep(1 * time.Second)

	//verify log was persisted
	assert.Equal(t, "1", client.Get("A"))
	assert.Equal(t, "A, 1\n", client.Print())

	//close server
	err = cmd.Process.Signal(syscall.SIGKILL)
	if err != nil {
		t.Fatal("Error killing server process. Error: ", err.Error())
	}
	cmd.Wait()

}
