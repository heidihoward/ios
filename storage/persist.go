package storage

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"os"
	"strconv"
)

type FileStorage struct {
	viewFile wal
	logFile  wal
	snapFile wal
}

func MakeFileStorage(diskPath string, persistenceMode string) *FileStorage {
	// create disk path if needs be
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		err = os.MkdirAll(diskPath, 0777)
		if err != nil {
			glog.Fatal(err)
		}
	}

	logFilename := diskPath + "/log.temp"
	dataFilename := diskPath + "/view.temp"
	snapFilename := diskPath + "/snapshot.temp"

	viewFile := openWriteAheadFile(dataFilename, persistenceMode, 64)
	logFile := openWriteAheadFile(logFilename, persistenceMode, 64*1000*1000)
	snapFile := openWriteAheadFile(snapFilename, persistenceMode, 64*1000*1000)
	s := FileStorage{viewFile, logFile, snapFile}
	return &s
}

func (fs *FileStorage) PersistView(view int) {
	glog.Info("Updating view to ", view, " in persistent storage")
	fs.viewFile.writeAhead([]byte(strconv.Itoa(view)))
}

func (fs *FileStorage) PersistLogUpdate(log msgs.LogUpdate) {
	glog.V(1).Info("Updating log with ", log, " in persistent storage")
	b, err := msgs.Marshal(log)
	if err != nil {
		glog.Fatal(err)
	}
	// write to persistent storage
	fs.logFile.writeAhead(b)
}

// PersistSnapshot writes the provided snapshot to persistent storage
// index should be the index of the last request applied to be state machine before snapshoting
func (fs *FileStorage) PersistSnapshot(index int, bytes []byte) {
	glog.Info("Saving request cache and state machine snapshot upto index", index,
		" of size ", len(bytes))
	fs.snapFile.writeAhead([]byte(strconv.Itoa(index)))
	fs.snapFile.writeAhead(bytes)
	// TODO: garbage collection - remove log entries/snapshots from before index
}
