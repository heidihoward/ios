package unix

import (
	"bufio"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
	"os"
	"strconv"
	"strings"
)

type fileHandler struct {
	Filename string
	IsNew    bool
	W        *bufio.Writer
	R        *bufio.Reader
	Fd       *os.File
}

func openFile(filename string) fileHandler {
	// check if file exists
	var isNew bool
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		glog.V(1).Info("Creating and opening file: ", filename)
		isNew = true
	} else {
		glog.V(1).Info("Opening file: ", filename)
		isNew = false
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
	return fileHandler{filename, isNew, w, r, file}
}

func restoreLog(logFile fileHandler, MaxLength int, snapshotIndex int) (bool, *consensus.Log) {

	if logFile.IsNew {
		return false, consensus.NewLog(MaxLength)
	}

	found := false
	log := consensus.RestoreLog(MaxLength, snapshotIndex)

	for {
		b, err := logFile.R.ReadBytes(byte('\n'))
		if err != nil {
			glog.V(1).Info("No more commands in persistent storage, ", log.LastIndex, " log entries were recovered")
			break
		}
		found = true
		var update msgs.LogUpdate
		err = msgs.Unmarshal(b, &update)
		if err != nil {
			glog.Fatal("Cannot parse log update", err)
		}
		// add enties to the log (in-memory)
		log.AddEntries(update.StartIndex, update.EndIndex, update.Entries)
		glog.V(1).Info("Adding from persistent storage :", update)
	}

	return found, log
}

func restoreView(viewFile fileHandler) (bool, int) {
	found := false
	view := 0

	if viewFile.IsNew {
		return found, view
	}

	for {
		b, err := viewFile.R.ReadBytes(byte('\n'))
		if err != nil {
			glog.V(1).Info("No more view updates in persistent storage")
			return found, view
		}
		found = true
		view, _ = strconv.Atoi(string(b))
	}
}

func restoreSnapshot(snapFile fileHandler, appConfig string) (bool, int, *app.StateMachine) {
	if snapFile.IsNew {
		return false, -1, app.New(appConfig)
	}

	// fetching index from snapshot file
	b, err := snapFile.R.ReadBytes(byte('\n'))
	if err != nil {
		glog.Warning("Snapshot corrupted, ignoring snapshot", err)
		return false, -1, app.New(appConfig)
	}
	index, err := strconv.Atoi(strings.Trim(string(b), "\n"))
	if err != nil {
		glog.Warning("Snapshot corrupted, ignoring snapshot", err)
		return false, -1, app.New(appConfig)
	}

	// fetch state machine snapshot from shapshot file
	snapshot, err := snapFile.R.ReadBytes(byte('\n'))
	if err != nil {
		glog.Warning("Snapshot corrupted, ignoring snapshot", err)
		return false, -1, app.New(appConfig)
	}
	return true, index, app.RestoreSnapshot(snapshot,appConfig)
}

func setupDummyStorage(io *msgs.Io, MaxLength int, appConfig string) (bool, int, *consensus.Log, int, *app.StateMachine) {
	glog.Warning("UNSAFE configuration - Do not use in production")
	go io.DumpPersistentStorage()
	return false, 0, consensus.NewLog(MaxLength) ,-1, app.New(appConfig)
}

func setupPersistentStorage(logFile string, dataFile string, snapFile string, io *msgs.Io, MaxLength int,  persistenceMode string, appConfig string) (bool, int, *consensus.Log, int, *app.StateMachine) {
	// setting up persistent log
	logStorage := openFile(logFile)
	dataStorage := openFile(dataFile)
	snapStorage := openFile(snapFile)

	// check persistent storage for view
	foundView, view := restoreView(dataStorage)
	// check persistent storage for snapshots
	foundSnapshot, index, state := restoreSnapshot(snapStorage, appConfig)
	// check persistent storage for commands
	foundLog, log := restoreLog(logStorage, MaxLength, index)

	if foundLog && !foundView {
		glog.Fatal("Log is present but view is not, this should not occur")
	}
	if foundSnapshot && !foundView && !foundLog {
		glog.Fatal("Snapshot is present but view/log is not, this should not occur")
	}

	// write view updates to persistent storage
	go func() {
		for {
			view := <-io.ViewPersist
			glog.V(1).Info("Updating view to ", view)
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
	logStorage.Fd.Close()
	wal := openWriteAheadFile(logFile,persistenceMode)
	go func() {
		for {
			log := <-io.LogPersist
			glog.V(1).Info("Updating log with ", log)
			b, err := msgs.Marshal(log)
			if err != nil {
				glog.Fatal(err)
			}
			// write to persistent storage
			wal.writeAhead(b)
			//n2, err := logStorage.Fd.Write([]byte("\n"))
			io.LogPersistFsync <- log
		}
	}()
	// write state machine snapshots to persistent storage
	go func() {
		for {
			snap := <-io.SnapshotPersist
			glog.V(1).Info("Saving request cache and state machine snapshot upto index", snap.Index,
				" of size ", len(snap.Bytes))
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

	return foundView, view, log, index, state
}

func SetupStorage(logFile string, dataFile string, snapFile string, io *msgs.Io, maxLength int, dummyStorage bool, persistenceMode string, appConfig string) (bool, int, *consensus.Log, int, *app.StateMachine) {
	if dummyStorage {
		return setupDummyStorage(io,maxLength, appConfig)
	}
 return setupPersistentStorage(logFile, dataFile, snapFile, io, maxLength, persistenceMode, appConfig)
}
