package client

import (
	"github.com/stretchr/testify/assert"
  "github.com/heidi-ann/ios/ios/server"
  "github.com/heidi-ann/ios/config"
  "github.com/golang/glog"
	"testing"
  "os"
  "io/ioutil"
)

func TestStartClient(t *testing.T) {
	assert := assert.New(t)
  //Create temp directories
  dirServer, err := ioutil.TempDir("", "IosStartClientTests")
  if err != nil {
    glog.Fatal(err)
  }
  defer os.RemoveAll(dirServer)
  dirClient, err := ioutil.TempDir("", "IosStartClientTests")
  if err != nil {
    glog.Fatal(err)
  }
  defer os.RemoveAll(dirClient)

  //start 1 node Ios cluster
  serverConfigFile := os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/ios/example.conf"
  go server.RunIos(0, config.ParseServerConfig(serverConfigFile), dirServer)

  //start client
  clientConfigFile := os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/client/example.conf"
  client := StartClientFromConfigFile(1, dirClient+"/latency.csv", clientConfigFile)
  success, reply := client.SubmitRequest("update A 1")
  assert.True(success,"Request not successful")
  assert.Equal("OK",reply,"Response not as expected")

}
