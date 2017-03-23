package storage

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/msgs"
	"github.com/stretchr/testify/assert"
)

func restoreStorageEmpty(t *testing.T, dir string) {
	assert := assert.New(t)
	found, view, log, index, state := RestoreStorage(dir, 1000, "kv-store")
	assert.False(found, "Unexpected persistent storage found")
	assert.Equal(0, view, "Unexpected view")
	assert.Equal(make([]msgs.Entry, 1000), log.LogEntries, "Unexpected log entries")
	assert.Equal(-1, index, "Unexpected index")
	assert.Equal(app.New("kv-store"), state, "Unexpected kv store")
}

func TestPersistentStorage(t *testing.T) {
	flag.Parse()
	defer glog.Flush()
	assert := assert.New(t)

	//Create temp directory
	dir, err := ioutil.TempDir("", "IosPersistentStorageTests")
	if err != nil {
		glog.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	// check recovery when no files exist
	restoreStorageEmpty(t, dir)

	//check file creation
	fs := MakeFileStorage(dir, "fsync")
	restoreStorageEmpty(t, dir)

	//verify files were created in directory
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		glog.Fatal(err)
	}

	filenames := make([]string, len(files))
	for i, file := range files {
		filenames[i] = file.Name()
	}
	assert.EqualValues([]string{"log.temp", "snapshot.temp", "view.temp"}, filenames)

	//verfiy that view storage works
	viewFile := dir + "/view.temp"
	found, view := restoreView(viewFile)
	assert.False(found, "Unexpected view found")
	for v := 0; v < 5; v++ {
		fs.PersistView(v)
		found, view = restoreView(viewFile)
		assert.True(found, "Missing view in ", viewFile)
		assert.Equal(v, view, "Incorrect view")
	}

	//verfiy that log storage works
	logFile := dir + "/log.temp"
	found, log := restoreLog(logFile, 100, -1)
	assert.False(found, "Unexpected log found")

	req1 := msgs.ClientRequest{
		ClientID:        1,
		RequestID:       1,
		ForceViewChange: false,
		Request:         "update A 1"}

	req2 := msgs.ClientRequest{
		ClientID:        1,
		RequestID:       2,
		ForceViewChange: false,
		Request:         "update B 2"}

	entry1 := msgs.Entry{
		View:      0,
		Committed: false,
		Requests:  []msgs.ClientRequest{req1}}

	up1 := msgs.LogUpdate{
		StartIndex: 0,
		EndIndex:   1,
		Entries:    []msgs.Entry{entry1}}

	fs.PersistLogUpdate(up1)
	found, log = restoreLog(logFile, 100, -1)
	assert.True(found, "Log expected but is missing")
	assert.Equal(entry1, log.GetEntry(0), "Log entry not as expected")

	//verify that snapshot storage works
	snapFile := dir + "/snapshot.temp"
	found, index, actualSm := restoreSnapshot(snapFile, "kv-store")
	assert.False(found, "Unexpected log found")
	assert.Equal(-1, index, "Unexpected index")
	assert.Equal(app.New("kv-store"), actualSm, "Unexpected kv store")

	sm := app.New("kv-store")
	sm.Apply(req1)
	fs.PersistSnapshot(0, sm.MakeSnapshot())
	found, index, actualSm = restoreSnapshot(snapFile, "kv-store")
	assert.True(found, "Unexpected log missing")
	assert.Equal(0, index, "Unexpected index")
	assert.Equal(sm, actualSm, "Missing kv store")

	// try 2nd snapshot
	sm.Apply(req2)
	fs.PersistSnapshot(1, sm.MakeSnapshot())
	found, index, actualSm = restoreSnapshot(snapFile, "kv-store")
	assert.True(found, "Unexpected log missing")
	assert.Equal(1, index, "Unexpected index")
	assert.Equal(sm, actualSm, "Missing kv store")

}
