package storage

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
	"io"
	"os"
	"strconv"
	"strings"
)

func restoreLog(logFilename string, maxLength int, snapshotIndex int) (bool, *consensus.Log) {
	exists, logFile := openReader(logFilename)

	if !exists {
		return false, consensus.NewLog(maxLength)
	}

	found := false
	log := consensus.RestoreLog(maxLength, snapshotIndex)

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
	if !exists {
		glog.Info("No view update file found")
		return false, 0
	}

	found := 0
	view := 0

	for {
		b, err := viewFile.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			glog.Fatal("View storage corrupted")
		}
		view, err = strconv.Atoi(strings.Trim(string(b), "\n"))
		if err != nil {
			glog.Fatal("View storage corrupted ", string(b))
		}
		found++
	}

	if found > 0 {
		glog.Info("No more view updates in persistent storage, most recent view is ", view)
	} else {
		glog.Info("No view updates found in persistent storage")
	}
	viewFile.closeReader()
	return found > 0, view
}

func restoreSnapshot(snapFilename string, appConfig string) (bool, int, *app.StateMachine) {
	exists, snapFile := openReader(snapFilename)
	if !exists {
		return false, -1, app.New(appConfig)
	}

	found := false
	currIndex := -1
	stateMachine := app.New(appConfig)

	for {
		// fetching index from snapshot file
		b, err := snapFile.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			glog.Warning("Snapshot corrupted, ignoring snapshot", err)
			break
		}
		index, err := strconv.Atoi(strings.Trim(string(b), "\n"))
		if err != nil {
			glog.Warning("Snapshot corrupted, ignoring snapshot", err)
			break
		}
		// fetch state machine snapshot from shapshot file
		snapshot, err := snapFile.read()
		if err != nil {
			glog.Warning("Snapshot corrupted, ignoring snapshot", err)
			break
		}
		// update with latest snapshot, now that it is completed
		found = true
		currIndex = index
		stateMachine = app.RestoreSnapshot(snapshot, appConfig)

	}

	snapFile.closeReader()
	return found, currIndex, stateMachine
}

func RestoreStorage(diskPath string, maxLength int, appConfig string) (bool, int, *consensus.Log, int, *app.StateMachine) {
	// check if disk directory exists
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return false, 0, consensus.NewLog(maxLength), -1, app.New(appConfig)
	}

	logFile := diskPath + "/log.temp"
	dataFile := diskPath + "/view.temp"
	snapFile := diskPath + "/snapshot.temp"

	// check persistent storage for view
	foundView, view := restoreView(dataFile)
	// check persistent storage for snapshots
	foundSnapshot, index, state := restoreSnapshot(snapFile, appConfig)
	// check persistent storage for commands
	foundLog, log := restoreLog(logFile, maxLength, index)

	if foundLog && !foundView {
		glog.Fatal("Log is present but view is not, this should not occur")
	}
	if foundSnapshot && !foundView && !foundLog {
		glog.Fatal("Snapshot is present but view/log is not, this should not occur")
	}

	return foundView, view, log, index, state
}
