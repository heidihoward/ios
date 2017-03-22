// +build linux

package storage

import (
	"github.com/golang/glog"
	"os"
	"syscall"
	"time"
)

type wal struct {
	fd   int
	mode string
}

func openWriteAheadFile(filename string, mode string) wal {
	var file int
	var err error
	switch mode {
	case "osync":
		file, err = syscall.Open(filename, syscall.O_SYNC|os.O_WRONLY|os.O_CREATE, 0666)
	case "dsync":
		file, err = syscall.Open(filename, syscall.O_DSYNC|os.O_WRONLY|os.O_CREATE, 0666)
	case "direct":
		file, err = syscall.Open(filename, syscall.O_DIRECT|os.O_WRONLY|os.O_CREATE, 0666)
	case "none", "fsync":
		file, err = syscall.Open(filename, os.O_WRONLY|os.O_CREATE, 0666)
	default:
		glog.Fatal("PersistenceMode not reconised")
	}
	if err != nil {
		glog.Fatal(err)
	}
	glog.Info("Opened file: ",filename)
	// TOD0: remove hardcoded filesize
	SEEK_CUR := 1
	start, err := syscall.Seek(file,0,SEEK_CUR)
	if err != nil {
		glog.Fatal(err)
	}

	glog.Info("Starting file pointer ",start)
	err = syscall.Fallocate(file, 0, 0, int64(64*1000*1000)) // 64MB
	if err != nil {
		glog.Fatal(err)
	}
	curr, err := syscall.Seek(file,0,SEEK_CUR)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Info("Current file pointer ",curr)

	SEEK_SET := 0
	finish, err := syscall.Seek(file,start,SEEK_SET)
	if err != nil {
		glog.Fatal(err)
	}
	if start != finish {
		glog.Fatal("unsuccessful at resetting file pointer", start, finish)
	}
	glog.Info("Final file pointer ",finish)
	return wal{file, mode}
}

func (w wal) writeAhead(bytes []byte) {
	startTime := time.Now()
	n, err := syscall.Write(w.fd, bytes)
	if err != nil {
		glog.Fatal(err)
	}
	if n != len(bytes) {
		glog.Fatal("Short write")
	}
	delim := []byte("\n")
	n2, err := syscall.Write(w.fd, delim)
	if err != nil {
		glog.Fatal(err)
	}
	if n2 != len(delim) {
		glog.Fatal("Short write")
	}
	if w.mode == "fsync" || w.mode == "direct" {
		err = syscall.Fdatasync(w.fd)
		if err != nil {
			glog.Fatal(err)
		}
	}
	if time.Since(startTime) > time.Millisecond {
		glog.Info(n+n2," bytes written & synced to persistent log in ", time.Since(startTime).String())
	}
	glog.V(1).Info(n+n2," bytes written & synced to persistent log in ", time.Since(startTime).String())
}
