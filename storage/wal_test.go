package storage

import (
	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

func TestOpenWriteAheadFile(t *testing.T) {
	assert := assert.New(t)

	//Create temp directory
	dir, err := ioutil.TempDir("", "IosWALTests")
	if err != nil {
		glog.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	//create file
	testFile := dir + "/test.temp"
	wal := openWriteAheadFile(testFile, "fsync", 2560)
	actualBytes, err := ioutil.ReadFile(testFile)
	assert.Nil(err)
	// TODO; check file is expected size (maybe system dependant)
	glog.Info("File created of size ", len(actualBytes))
	// TODO: check actualBytes are empty

	//verfiy that write ahead logging works
	start := 0
	history := make([]byte, 0, 1000)
	for size := 1; size < 100; size += 10 {
		expectedBytes := make([]byte, size)
		rand.Read(expectedBytes)
		wal.writeAhead(expectedBytes)
		actualBytes, err = ioutil.ReadFile(testFile)
		assert.Nil(err)
		assert.Equal(expectedBytes, actualBytes[start:start+size], "Bytes read are not same as written")
		assert.Equal([]byte{0xa}, actualBytes[start+size:start+size+1], "Delim missing from end of write")
		assert.Equal(history[:start], actualBytes[:start], "Past writes have been corrupted")
		// TODO: check that rest of file is empty
		// update state for next run
		history = append(history, expectedBytes...)
		history = append(history, 0xa)
		start = start + size + 1
	}

	//verfiy that write ahead logging continues after failures
	walf := openWriteAheadFile(testFile, "fsync", 2560)
	actualBytesf, err := ioutil.ReadFile(testFile)
	assert.Nil(err)
	glog.Info("File now of size ", len(actualBytes))
	assert.Equal(history, actualBytesf[:start], "File has not recovered")

	// continue logging
	// for size := 1; size < 100; size += 10 {
	// 	expectedBytes := make([]byte, size)
	// 	rand.Read(expectedBytes)
	// 	walf.writeAhead(expectedBytes)
	// 	actualBytes, err = ioutil.ReadFile(testFile)
	// 	assert.Nil(err)
	// 	assert.Equal(expectedBytes, actualBytes[start:start+size], "Bytes read are not same as written")
	// 	assert.Equal([]byte{0xa}, actualBytes[start+size:start+size+1], "Delim missing from end of write")
	// 	assert.Equal(history[:start], actualBytes[:start], "Past writes have been corrupted")
	// 	// update state for next run
	// 	history = append(history, expectedBytes...)
	// 	history = append(history, 0xa)
	// 	start = start + size + 1
	// }

}
