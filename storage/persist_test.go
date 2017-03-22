package storage

import (
	"io/ioutil"
	"os"
	"testing"


	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
  "github.com/heidi-ann/ios/msgs"

)

func TestPersistentStorage(t *testing.T) {
  assert := assert.New(t)

	//Create temp directory
	dir, err := ioutil.TempDir("", "IosPersistentStorageTests")
	if err != nil {
		glog.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	//check file creation
	fs := MakeFileStorage(dir,"fsync")

	//verify files were created in directory
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		glog.Fatal(err)
	}

	filenames := make([]string, len(files))
	for i, file := range files {
		filenames[i] = file.Name()
	}
	assert.EqualValues([]string{"log.temp", "snapshot.temp", "view.temp"}, filenames)

  //verfiy that view storage works
  viewFile := dir+"/view.temp"
  found, view := restoreView(viewFile)
  assert.False(found,"Unexpected view found")
  for v := 0; v <5; v++ {
    fs.PersistView(v)
    found, view = restoreView(viewFile)
    assert.True(found,"Missing view in ",viewFile)
    assert.Equal(v,view,"Incorrect view")
  }

  //verfiy that log storage works
  logFile := dir+"/log.temp"
  found, log := restoreLog(logFile,100,-1)
  assert.False(found,"Unexpected log found")

  req1 :=  msgs.ClientRequest{
  	ClientID:1,
  	RequestID: 1,
  	ForceViewChange: false,
  	Request:"update A 1"}

  entry1 := msgs.Entry{
    View: 0,
    Committed: false,
    Requests:  []msgs.ClientRequest{req1}}

  up1 := msgs.LogUpdate{
  	StartIndex: 0,
  	EndIndex: 1,
  	Entries:    []msgs.Entry{entry1}}

  fs.PersistLogUpdate(up1)
  found, log = restoreLog(logFile,100,-1)
  assert.True(found,"Log expected but is missing")
  assert.Equal(entry1,log.GetEntry(0),"Log entry not as expected")

}
