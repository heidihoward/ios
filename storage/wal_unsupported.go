// +build !darwin,!linux

package storage

import (
	"github.com/golang/glog"
	"os"
	"time"
)

type wal struct {
	file *os.File
	mode string
}

func openWriteAheadFile(filename string, mode string) wal {
	var file *os.File
	var err error
	switch mode {
	case "none", "fsync":
		file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	default:
		glog.Fatal("PersistenceMode not recognised, only fsync and none are avalible on darwin")
	}
	if err != nil {
		glog.Fatal(err)
	}
	if err != nil {
		glog.Fatal(err)
	}
	return wal{file, mode}
}

func (w wal) writeAhead(bytes []byte) {
	startTime := time.Now()
	_, err := w.file.Write(bytes)
	_,_ = w.file.Write([]byte("\n"))
	if err != nil {
		glog.Fatal(err)
	}
	if w.mode == "fsync" || w.mode == "direct" {
		err = w.file.Sync()
		if err != nil {
			glog.Fatal(err)
		}
	}
	if time.Since(startTime) > time.Millisecond {
		glog.Info(" bytes written & synced to persistent log in ", time.Since(startTime).String())
	}
	glog.V(1).Info(" bytes written & synced to persistent log in ", time.Since(startTime).String())
}
