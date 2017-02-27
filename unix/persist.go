package unix

import (
	"bufio"
	"github.com/golang/glog"
  "github.com/heidi-ann/ios/msgs"
  "os"
  "strconv"
  "github.com/heidi-ann/ios/store"
  "strings"
)

type FileHandler struct {
  Filename string
  IsNew bool
  W *bufio.Writer
  R *bufio.Reader
  Fd *os.File
}

func openFile(filename string) FileHandler {
	// check if file exists
	var is_new bool
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		glog.Info("Creating and opening file: ", filename)
		is_new = true
	} else {
		glog.Info("Opening file: ", filename)
		is_new = false
	}

	// open file
  // TODO: consider using O_SYNC for write ahead logging
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		glog.Fatal(err)
	}

	// create writer and reader
	w := bufio.NewWriter(file)
	r := bufio.NewReader(file)
	return FileHandler{filename, is_new, w, r, file}
}

func restoreLog(logFile FileHandler, MaxLength int) (bool, []msgs.Entry) {

  if logFile.IsNew {
    return false, []msgs.Entry{}
  }

  found := false
  log := make([]msgs.Entry, MaxLength)
  logLength := 0

  for {
    b, err := logFile.R.ReadBytes(byte('\n'))
    if err != nil {
      glog.Info("No more commands in persistent storage, ",logLength," log entries were recovered")
      break
    }
    found = true
    var update msgs.LogUpdate
    err = msgs.Unmarshal(b, &update)
    if err != nil {
      glog.Fatal("Cannot parse log update", err)
    }
    // add enties to the log (in-memory)
    for i := 0; i < update.EndIndex - update.StartIndex; i++ {
      log[update.StartIndex + i] = update.Entries[i]
    }
    if logLength < update.EndIndex {
      logLength = update.EndIndex
    }
    glog.Info("Adding from persistent storage :", update)
  }

  return found, log[:logLength]
}

func restoreView(viewFile FileHandler) (bool, int) {
  found := false
  view := 0

  if viewFile.IsNew {
    return found, view
  }

  for {
    b, err := viewFile.R.ReadBytes(byte('\n'))
    if err != nil {
      glog.Info("No more view updates in persistent storage")
      return found, view
    }
    found = true
    view, _ = strconv.Atoi(string(b))
  }
}

func restoreSnapshot(snapFile FileHandler) (bool, int, *store.Store) {
  if snapFile.IsNew {
    return false, -1, store.New()
  }

  // fetching index from snapshot file
  b, err := snapFile.R.ReadBytes(byte('\n'))
  if err != nil {
    glog.Warning("Snapshot corrupted, ignoring snapshot", err)
    return false, -1, store.New()
  }
  index, err := strconv.Atoi(strings.Trim(string(b),"\n"))
  if err != nil {
    glog.Warning("Snapshot corrupted, ignoring snapshot", err)
    return false, -1, store.New()
  }

  // fetch state machine snapshot form shapshot file
  snap, err := snapFile.R.ReadBytes(byte('\n'))
  if err != nil {
    glog.Warning("Snapshot corrupted, ignoring snapshot",err)
    return false, -1, store.New()
  }
  return true, index, store.RestoreSnapshot(snap)
}

func SetupPersistentStorage(logFile string, dataFile string, snapFile string, io *msgs.Io, MaxLength int) (bool, int, []msgs.Entry, int, *store.Store) {
	// setting up persistent log
	logStorage := openFile(logFile)
	dataStorage := openFile(dataFile)
  snapStorage := openFile(snapFile)

	// check persistent storage for commands & view
	foundLog, log := restoreLog(logStorage,MaxLength)
	foundView, view := restoreView(dataStorage)
  if foundLog && !foundView {
    glog.Fatal("Log is present but view is not, this should not occur")
  }
  // check persistent storage for snapshots
  foundSnapshot, index, store := restoreSnapshot(snapStorage)
  if foundSnapshot && !foundView && !foundLog{
    glog.Fatal("Snapshot is present but view/log is not, this should not occur")
  }

	// write view updates to persistent storage
	go func() {
		for {
			view := <-io.ViewPersist
			glog.Info("Updating view to ", view)
			_, err := dataStorage.Fd.Write([]byte(strconv.Itoa(view)))
			_, err = dataStorage.Fd.Write([]byte("\n"))
			if err != nil {
				glog.Fatal(err)
			}
			dataStorage.Fd.Sync()
			io.ViewPersistFsync <- view
		}
	}()
	// write log updates to persistent storage
	go func() {
		for {
			log := <-io.LogPersist
			glog.Info("Updating log with ", log)
			b, err := msgs.Marshal(log)
			if err != nil {
				glog.Fatal(err)
			}
			// write to persistent storage
			n1, err := logStorage.Fd.Write(b)
			n2, err := logStorage.Fd.Write([]byte("\n"))
			if err != nil {
				glog.Fatal(err)
			}
			glog.Info(n1+n2, " bytes written to persistent log")
      if log.Sync {
        logStorage.Fd.Sync()
        io.LogPersistFsync <- log
      }

		}
	}()
  // write state machine snapshots to persistent storage
  go func() {
    for {
      snap := <-io.SnapshotPersist
      glog.Info("Saving state machine snapshot upto index", snap.Index,
      " of size ",len(snap.Bytes))
      file, err := os.OpenFile(snapFile, os.O_RDWR|os.O_CREATE, 0777)
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
  }()

  return foundView, view, log, index, store
}
