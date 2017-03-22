package storage

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
  "strconv"
  "strings"
)


func restoreLog(logFilename string, MaxLength int, snapshotIndex int) (bool, *consensus.Log) {
  exists, logFile := openReader(logFilename)

	if !exists {
		return false, consensus.NewLog(MaxLength)
	}

	found := false
	log := consensus.RestoreLog(MaxLength, snapshotIndex)

	for {
		b, err := logFile.read()
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
  logFile.closeReader()
	return found, log
}

func restoreView(viewFilename string) (bool, int) {
  exists, viewFile := openReader(viewFilename)
	found := false
	view := 0

	if !exists {
		return found, view
	}

	for {
		b, err := viewFile.read()
		if err != nil {
			glog.V(1).Info("No more view updates in persistent storage")
      viewFile.closeReader()
			return found, view
		}
		found = true
		view, _ = strconv.Atoi(string(b))
	}
}

func restoreSnapshot(snapFilename string, appConfig string) (bool, int, *app.StateMachine) {
  exists, snapFile := openReader(snapFilename)
	if !exists {
		return false, -1, app.New(appConfig)
	}

	// fetching index from snapshot file
	b, err := snapFile.read()
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
	snapshot, err := snapFile.read()
	if err != nil {
		glog.Warning("Snapshot corrupted, ignoring snapshot", err)
		return false, -1, app.New(appConfig)
	}

  snapFile.closeReader()
	return true, index, app.RestoreSnapshot(snapshot, appConfig)
}

func RestoreStorage(logFile string, dataFile string, snapFile string, MaxLength int, appConfig string) (bool, int, *consensus.Log, int, *app.StateMachine) {
	// check persistent storage for view
	foundView, view := restoreView(dataFile)
	// check persistent storage for snapshots
	foundSnapshot, index, state := restoreSnapshot(snapFile, appConfig)
	// check persistent storage for commands
	foundLog, log := restoreLog(logFile, MaxLength, index)

	if foundLog && !foundView {
		glog.Fatal("Log is present but view is not, this should not occur")
	}
	if foundSnapshot && !foundView && !foundLog {
		glog.Fatal("Snapshot is present but view/log is not, this should not occur")
	}

	return foundView, view, log, index, state
}
