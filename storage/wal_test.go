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
  assert.Equal(1001,len(actualBytes), "Number of bytes read is not same as bytes written")
  assert.Equal(expectedBytes,actualBytes[:1000], "Bytes read are not same as written")

}
