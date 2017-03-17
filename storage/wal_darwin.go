// +build darwin

package storage

import (
	"github.com/golang/glog"
	"os"
	"syscall"
	"time"
)

type WAL struct {
	fd   int
	mode string
}

func openWriteAheadFile(filename string, mode string) WAL {
	var file int
	var err error
	switch mode {
	case "none", "fsync":
		file, err = syscall.Open(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	default:
		glog.Fatal("PersistenceMode not recognised, only fsync and none are avalible on darwin")
	}
	if err != nil {
		glog.Fatal(err)
	}
	if err != nil {
		glog.Fatal(err)
	}
	return WAL{file, mode}
}

func (w WAL) writeAhead(bytes []byte) {
	startTime := time.Now()
	n, err := syscall.Write(w.fd, bytes)
	_,_ = syscall.Write(w.fd, []byte("\n"))
	if err != nil || n != len(bytes) {
		glog.Fatal(err)
	}
	if w.mode == "fsync" || w.mode == "direct" {
		err = syscall.Fsync(w.fd)
		if err != nil {
			glog.Fatal(err)
		}
	}
	if time.Since(startTime) > time.Millisecond {
		glog.Info("Slow disk warning - ", n, " bytes written & synced to persistent log in ", time.Since(startTime).String())
	}
	glog.V(1).Info(" bytes written & synced to persistent log in ", time.Since(startTime).String())
}
