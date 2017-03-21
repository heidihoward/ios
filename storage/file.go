package storage

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
  "os"
  "strconv"
)

type FileStorage struct {
  viewFile *os.File
  logFile  WAL
  snapFile string
}

func MakeFileStorage(viewFile *os.File, logFile WAL, snapFile string) *FileStorage {
	s := FileStorage{viewFile, logFile, snapFile}
	return &s
}

func (fs *FileStorage) PersistView(view int) {
  glog.Info("Updating view to ", view, " in persistent storage")
  _, err := fs.viewFile.Write([]byte(strconv.Itoa(view)))
  _, err = fs.viewFile.Write([]byte("\n"))
  if err != nil {
    glog.Fatal(err)
  }
  fs.viewFile.Sync()
}

func (fs *FileStorage) PersistLogUpdate(log msgs.LogUpdate) {
  glog.V(1).Info("Updating log with ", log)
  b, err := msgs.Marshal(log)
  if err != nil {
    glog.Fatal(err)
  }
  // write to persistent storage
  fs.logFile.writeAhead(b)
}

func (fs *FileStorage) PersistSnapshot(snap msgs.Snapshot) {
  glog.V(1).Info("Saving request cache and state machine snapshot upto index", snap.Index,
    " of size ", len(snap.Bytes))
  file, err := os.OpenFile(fs.snapFile, os.O_RDWR|os.O_CREATE, 0777)
  if err != nil {
    glog.Fatal(err)
  }

  _, err = file.Write([]byte(strconv.Itoa(snap.Index)))
  _, err = file.Write([]byte("\n"))
  _, err = file.Write([]byte(snap.Bytes))
  _, err = file.Write([]byte("\n"))
  if err != nil {
    glog.Fatal(err)
  }
}
