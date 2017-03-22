package storage

import (
	"bufio"
	"github.com/golang/glog"
	"os"
)

type fileReader struct {
	filename string
	rd       *bufio.Reader
	fd       *os.File
}

func openReader(filename string) (exists bool, reader fileReader) {
	// check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false, fileReader{}
	}

	// open file
	glog.Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		glog.Fatal(err)
	}

	// create reader
	r := bufio.NewReader(file)
	return true, fileReader{filename, r, file}
}

func (r fileReader) read() ([]byte, error) {
	return r.rd.ReadBytes(byte('\n'))
}

func (r fileReader) closeReader() {
	err := r.fd.Close()
	if err != nil {
		glog.Fatal(err)
	}
}
