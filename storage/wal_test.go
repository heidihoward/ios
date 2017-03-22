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

	//verfiy that write ahead logging works
  expectedBytes := make([]byte, 1000)
	rand.Read(expectedBytes)
  wal.writeAhead(expectedBytes)
  actualBytes, err := ioutil.ReadFile(testFile)
  assert.Nil(err)
  assert.Equal(append(expectedBytes,0xa),actualBytes, "Bytes read are not same as written")

}
