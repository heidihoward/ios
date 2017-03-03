package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
)

// Log holds the replication log used of Ios.
// Only indexes between AbsoluteIndex and (AbsoluteIndex + MaxLength -1)  are accessible
type Log struct {
	LogEntries    []msgs.Entry // contents of log, indexed from 0 to MaxLength - 1
	LastIndex     int          // greatest absolute index in log with entry, -1 means that the log has no entries
	AbsoluteIndex int          // absolute index of first index in log
	MaxLength     int          // maximum length of LogEntries
}

// check protocol invariant
func (l *Log) checkInvariants(index int, nxtEntry msgs.Entry) {
	prevEntry := l.LogEntries[index]
	// if no entry, then no problem
	if reflect.DeepEqual(prevEntry, msgs.Entry{}) {
		return
	}
	// if committed, request never changes
	if prevEntry.Committed && !reflect.DeepEqual(prevEntry.Requests, nxtEntry.Requests) {
		glog.Fatal("Committed entry is being overwritten at ", prevEntry, nxtEntry, index)
	}
	// each index is allocated once per view
	if prevEntry.View == nxtEntry.View && !reflect.DeepEqual(prevEntry.Requests, nxtEntry.Requests) {
		glog.Fatal("Index has been reallocated at ", prevEntry, nxtEntry, index)
	}
	// entries should only be overwritten by same or higher view
	if prevEntry.View > nxtEntry.View {
		glog.Fatal("Attempting to overwrite entry with lower view", prevEntry, nxtEntry, index)
	}
}

func NewLog(maxLength int) *Log {
	return &Log{make([]msgs.Entry, maxLength), -1, 0, maxLength}
}

func RestoreLog(maxLength int, startIndex int) *Log {
	return &Log{make([]msgs.Entry, maxLength), startIndex, startIndex, maxLength}
}

func (l *Log) AddEntries(startIndex int, endIndex int, entries []msgs.Entry) {
	// check correct number of entries has been given
	if len(entries) != endIndex-startIndex {
		glog.Fatal("Wrong number of log entries provided")
	}
	// check indexes are accessible
	if startIndex < l.AbsoluteIndex {
		return
	}
	// check indexes are accessible
	if endIndex > l.AbsoluteIndex+l.MaxLength-1 {
		glog.Fatal("Log index is too large, please snapshot log")
	}
	// add entries and check invariants
	for i := 0; i < len(entries); i++ {
		l.checkInvariants(startIndex+i, entries[i])
		l.LogEntries[startIndex+i-l.AbsoluteIndex] = entries[i]
	}
	// update LastIndex
	if endIndex-1 > l.LastIndex {
		l.LastIndex = endIndex - 1
	}
}

func (l *Log) AddEntry(index int, entry msgs.Entry) {
	l.AddEntries(index, index+1, []msgs.Entry{entry})
}

func (l *Log) GetEntries(startIndex int, endIndex int) []msgs.Entry {
	// check indexes are accessible
	if startIndex < l.AbsoluteIndex || endIndex > l.AbsoluteIndex+l.MaxLength-1 {
		glog.Fatal("Trying to access log out of bounds")
	}
	return l.LogEntries[startIndex-l.AbsoluteIndex : endIndex-l.AbsoluteIndex]
}

func (l *Log) GetEntry(index int) msgs.Entry {
	// check indexes are accessible
	if index < l.AbsoluteIndex || index > l.AbsoluteIndex+l.MaxLength-1 {
		glog.Fatal("Trying to access log out of bounds")
	}
	return l.LogEntries[index-l.AbsoluteIndex]
}
