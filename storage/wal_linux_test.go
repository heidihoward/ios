// +build linux

package storage

import (
	"io/ioutil"
	"os"
	"testing"
	"math/rand"
	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
)

func TestWAL(t *testing.T) {
	assert := assert.New(t)

	//Create temp directory
	dir, err := ioutil.TempDir("", "IosWALTests")
	if err != nil {
		glog.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	//create file
  testFile := dir + "/test.temp"
	wal := openWriteAheadFile(testFile, "fsync")
	actualBytes, err := ioutil.ReadFile(testFile)
	assert.Equal(64*1000*1000,len(actualBytes), "File is expected size")

	//verfiy that write ahead logging works
  expectedBytes := make([]byte, 100)
	rand.Read(expectedBytes)
  wal.writeAhead(expectedBytes)
  actualBytes, err = ioutil.ReadFile(testFile)
  assert.Nil(err)
  //assert.Equal(1001,len(actualBytes), "Number of bytes read is not same as bytes written")
  assert.Equal(append(expectedBytes,0xa),actualBytes[len(actualBytes)-100-1:], "Bytes read are not same as written")

}
