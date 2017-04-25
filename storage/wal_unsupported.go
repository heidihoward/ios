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

func openWriteAheadFile(filename string, mode string, _ int) (wal, error) {
	var file *os.File
	switch mode {
	case "none", "fsync":
		file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			glog.Fatal(err)
		}
	default:
		glog.Fatal("PersistenceMode not recognised, only fsync and none are avalible on darwin")
	}
	return wal{file, mode}, nil
}

func (w wal) writeAhead(bytes []byte) error {
	// write bytes
	startTime := time.Now()
	n, err := w.file.Write(bytes)
	if err != nil {
		glog.Fatal(err)
	}
	n2, err := w.file.Write([]byte("\n"))
	if err != nil {
		glog.Fatal(err)
	}
	glog.V(1).Info(n+n2, " bytes written to persistent log in ", time.Since(startTime).String())
	// sync if needed
	if w.mode == "fsync" {
		err = w.file.Sync()
		if err != nil {
			glog.Fatal(err)
		}
		glog.V(1).Info(n+n2, " bytes synced to persistent log in ", time.Since(startTime).String())
	}
	slowDiskWarning(startTime, n+n2)
	return nil
}
