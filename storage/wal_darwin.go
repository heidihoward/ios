// +build darwin

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

func openWriteAheadFile(filename string, mode string, _ int) wal {
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
	glog.Info("Opened file: ", filename)
	return wal{file, mode}
}

func (w wal) writeAhead(bytes []byte) {
	startTime := time.Now()
	n, err := syscall.Write(w.fd, bytes)
	if err != nil || n != len(bytes) {
		glog.Fatal(err)
	}
	n2, err := syscall.Write(w.fd, []byte("\n"))
	if err != nil {
		glog.Fatal(err)
	}
	glog.V(1).Info(n+n2, " bytes written to persistent log in ", time.Since(startTime).String())
	if w.mode == "fsync" || w.mode == "direct" {
		err = syscall.Fsync(w.fd)
		if err != nil {
			glog.Fatal(err)
		}
		glog.V(1).Info(n+n2, " bytes synced to persistent log in ", time.Since(startTime).String())
	}
	slowDiskWarning(startTime, n+n2)
}
