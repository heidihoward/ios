package storage

import (
	"bufio"
	"github.com/golang/glog"
	"os"
)

type fileWriter struct {
	filename string
	wt       *bufio.Writer
	fd       *os.File
}

func openWriter(filename string) fileWriter {
	// open file
	glog.Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		glog.Fatal(err)
	}

	// create writer
	wt := bufio.NewWriter(file)
	return fileWriter{filename, wt, file}
}

func (w fileWriter) write(b []byte) {
	n, err := w.wt.Write(b)
	if err != nil {
		glog.Fatal(err)
	}
	if n != len(b) {
		glog.Fatal("Short write")
	}
	_, err = w.wt.Write([]byte("\n"))
	if err != nil {
		glog.Fatal(err)
	}
	if n != len(b) {
		glog.Fatal("Short write")
	}
	return
}

func (w fileWriter) closeWriter() {
	// first flush bufio
	err := w.wt.Flush()
	if err != nil {
		glog.Fatal(err)
	}
	// then close file
	err = w.fd.Close()
	if err != nil {
		glog.Fatal(err)
	}
}
