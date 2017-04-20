// +build darwin

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

// openWriteAheadFile will open the specified file for the purpose of write ahead logging
// error returned if file cannot be opened or the persistence mode is not none or fsync
func openWriteAheadFile(filename string, mode string, _ int) (wal, error) {
	var WAL wal
	var err error
	WAL.mode = mode

	switch mode {
	case "none", "fsync":
		WAL.fd, err = syscall.Open(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			glog.Warning(err)
			return WAL, err
		}
		glog.Info("Opened file: ", filename)
		return WAL, nil
	default:
		return WAL, errors.New("PersistenceMode not recognised, only fsync and none are avalible on darwin")
	}
}

// writeAhead writes the specified bytes to the write ahead file
// function returns once bytes are written using specified persistence mode
// error is returned if bytes could not be written
func (w wal) writeAhead(bytes []byte) error {
	startTime := time.Now()
	n, err := syscall.Write(w.fd, bytes)
	if err != nil || n != len(bytes) {
		glog.Warning(err)
		return err
	}
	n2, err := syscall.Write(w.fd, []byte("\n"))
	if err != nil {
		glog.Warning(err)
		return err
	}
	glog.V(1).Info(n+n2, " bytes written to persistent log in ", time.Since(startTime).String())

	if w.mode == "fsync" || w.mode == "direct" {
		err = syscall.Fsync(w.fd)
		if err != nil {
			glog.Warning(err)
			return err
		}
		glog.V(1).Info(n+n2, " bytes synced to persistent log in ", time.Since(startTime).String())
	}
	slowDiskWarning(startTime, n+n2)
	return nil
}
