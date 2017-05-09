// +build linux

package storage

import (
	"errors"
	"os"
	"syscall"
	"time"

	"github.com/golang/glog"
)

type wal struct {
	fd   int
	mode string
}

func openWriteAheadFile(filename string, mode string, size int) (wal, error) {
	var WAL wal
	WAL.mode = mode
	var err error

	switch mode {
	case "osync":
		WAL.fd, err = syscall.Open(filename, syscall.O_SYNC|os.O_WRONLY|os.O_CREATE, 0666)
	case "dsync":
		WAL.fd, err = syscall.Open(filename, syscall.O_DSYNC|os.O_WRONLY|os.O_CREATE, 0666)
	case "direct":
		WAL.fd, err = syscall.Open(filename, syscall.O_DIRECT|os.O_WRONLY|os.O_CREATE, 0666)
	case "none", "fsync":
		WAL.fd, err = syscall.Open(filename, os.O_WRONLY|os.O_CREATE, 0666)
	default:
		return WAL, errors.New("PersistenceMode not reconised")
	}
	if err != nil {
		return WAL, err
	}
	glog.Info("Opened file: ", filename)
	start, err := syscall.Seek(WAL.fd, 0, 2)
	if err != nil {
		return WAL, err
	}

	glog.Info("Starting file pointer ", start)
	err = syscall.Fallocate(WAL.fd, 0, 0, int64(size))
	if err != nil {
		return WAL, err
	}
	return WAL, nil
}

func (w wal) writeAhead(bytes []byte) error {
	startTime := time.Now()
	// TODO: remove hack
	n, err := syscall.Write(w.fd, bytes)
	if err != nil {
		return err
	}
	if n != len(bytes) {
		return errors.New("Short write")
	}
	delim := []byte("\n")
	n2, err := syscall.Write(w.fd, delim)
	if err != nil {
		return err
	}
	if n2 != len(delim) {
		return errors.New("Short write")
	}
	glog.V(1).Info(n+n2, " bytes written to persistent log in ", time.Since(startTime).String())
	if w.mode == "fsync" || w.mode == "direct" {
		err = syscall.Fdatasync(w.fd)
		if err != nil {
			return err
		}
		glog.V(1).Info(n+n2, " bytes synced to persistent log in ", time.Since(startTime).String())
	}
	slowDiskWarning(startTime, n+n2)
	return nil
}
