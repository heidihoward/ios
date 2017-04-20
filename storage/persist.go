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

// MakeFileStorage opens the files for persistent storage
func MakeFileStorage(diskPath string, persistenceMode string) (*FileStorage, error) {
	// create disk path if needs be
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		err = os.MkdirAll(diskPath, 0777)
		if err != nil {
			return nil, err
		}
	}

	dataFilename := diskPath + "/view.temp"
	viewFile, err := openWriteAheadFile(dataFilename, persistenceMode, 64)
	if err != nil {
		return nil, err
	}
	logFilename := diskPath + "/log.temp"
	logFile, err := openWriteAheadFile(logFilename, persistenceMode, 64*1000*1000)
	if err != nil {
		return nil, err
	}
	snapFilename := diskPath + "/snapshot.temp"
	snapFile, err := openWriteAheadFile(snapFilename, persistenceMode, 64*1000*1000)
	if err != nil {
		return nil, err
	}

	s := FileStorage{viewFile, logFile, snapFile}
	return &s, nil
}

// PersistView writes view update to persistent storage
// Function returns when write+sync is completed
// Error is returned if it was not possible to write
func (fs *FileStorage) PersistView(view int) error {
	glog.Info("Updating view to ", view, " in persistent storage")
	v := []byte(strconv.Itoa(view))
	return fs.viewFile.writeAhead(v)
}

// PersistLogUpdate writes the specified log update to persistent storage
// Function returns when write+sync is completed
// Error is returned if it was not possible to write
func (fs *FileStorage) PersistLogUpdate(log msgs.LogUpdate) error {
	glog.V(1).Info("Updating log with ", log, " in persistent storage")
	b, err := msgs.Marshal(log)
	if err != nil {
		return err
	}
	// write to persistent storage
	if err := fs.logFile.writeAhead(b); err != nil {
		return err
	}
	return nil
}

// PersistSnapshot writes the provided snapshot to persistent storage
// index should be the index of the last request applied to be state machine before snapshoting
func (fs *FileStorage) PersistSnapshot(index int, bytes []byte) error {
	glog.Info("Saving request cache and state machine snapshot upto index", index,
		" of size ", len(bytes))
	if err := fs.snapFile.writeAhead([]byte(strconv.Itoa(index))); err != nil {
		return err
	}
	if err := fs.snapFile.writeAhead(bytes); err != nil {
		return err
	}

	return nil
	// TODO: garbage collection - remove log entries/snapshots from before index
}
