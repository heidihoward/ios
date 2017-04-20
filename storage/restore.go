package storage

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"

	"github.com/golang/glog"
)

// restoreLog looks for an existing log file and if found, recoveres an Ios log from it.
func restoreLog(logFilename string, maxLength int, snapshotIndex int) (bool, *consensus.Log, error) {
	exists, logFile, err := openReader(logFilename)
	if err != nil {
		return false, nil, err
	}

	if !exists {
		return false, consensus.NewLog(maxLength), nil
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
			return true, nil, errors.New("Log file corrupted")
		}
		// add enties to the log (in-memory)
		log.AddEntries(update.StartIndex, update.EndIndex, update.Entries)
		glog.V(1).Info("Adding from persistent storage :", update)
	}

	return found, log, logFile.closeReader()
}

// restoreView looks for an existing metadata file and if found, recoveres an Ios view from it.
func restoreView(viewFilename string) (bool, int, error) {
	exists, viewFile, err := openReader(viewFilename)
	if err != nil {
		return false, 0, nil
	}

	if !exists {
		glog.Info("No view update file found")
		return false, 0, nil
	}

	found := 0
	view := 0

	for {
		b, err := viewFile.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return true, 0, errors.New("View storage corrupted")
		}
		view, err = strconv.Atoi(strings.Trim(string(b), "\n"))
		if err != nil {
			return true, 0, errors.New("View storage corrupted")
		}
		found++
	}

	if found > 0 {
		glog.Info("No more view updates in persistent storage, most recent view is ", view)
	} else {
		glog.Info("No view updates found in persistent storage")
	}

	return found > 0, view, viewFile.closeReader()
}

func restoreSnapshot(snapFilename string, appConfig string) (bool, int, *app.StateMachine, error) {
	exists, snapFile, err := openReader(snapFilename)
	if err != nil {
		return false, -1, nil, err
	}

	if !exists {
		return false, -1, app.New(appConfig), nil
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
		stateMachine, err = app.RestoreSnapshot(snapshot, appConfig)
		if err != nil {
			glog.Warning("Snapshot corrupted, ignoring snapshot", err)
			break
		}

	}

	return found, currIndex, stateMachine, snapFile.closeReader()
}

func RestoreStorage(diskPath string, maxLength int, appConfig string) (bool, int, *consensus.Log, int, *app.StateMachine) {
	// check if disk directory exists
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return false, 0, consensus.NewLog(maxLength), -1, app.New(appConfig)
	}

	// check persistent storage for view
	dataFile := diskPath + "/view.temp"
	foundView, view, err := restoreView(dataFile)
	if err != nil {
		glog.Fatal(err)
	}

	// check persistent storage for snapshots
	snapFile := diskPath + "/snapshot.temp"
	foundSnapshot, index, state, err := restoreSnapshot(snapFile, appConfig)
	if err != nil {
		glog.Fatal(err)
	}

	// check persistent storage for commands
	logFile := diskPath + "/log.temp"
	foundLog, log, err := restoreLog(logFile, maxLength, index)
	if err != nil {
		glog.Fatal(err)
	}

	// check results are consistent
	if foundLog && !foundView {
		glog.Fatal("Log is present but view is not, this should not occur")
	}
	if foundSnapshot && !foundView && !foundLog {
		glog.Fatal("Snapshot is present but view/log is not, this should not occur")
	}

	return foundView, view, log, index, state
}
