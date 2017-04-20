package storage

import (
	"bufio"
	"os"

	"github.com/golang/glog"
)

type fileReader struct {
	filename string
	rd       *bufio.Reader
	fd       *os.File
}

func openReader(filename string) (bool, *fileReader, error) {
	// check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false, nil, nil
	}

	// open file
	glog.Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return false, nil, err
	}

	// create reader
	r := bufio.NewReader(file)
	return true, &fileReader{filename, r, file}, nil
}

func (r *fileReader) read() ([]byte, error) {
	return r.rd.ReadBytes(byte('\n'))
}

func (r *fileReader) closeReader() error {
	return r.fd.Close()
}
