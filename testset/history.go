package main

import (
	"encoding/csv"
	"github.com/golang/glog"
	"os"
	"strconv"
	"time"
)

// historyFile handles writing basic request stats such as latency to a csv file
type historyFile struct {
	w         *csv.Writer
	startTime time.Time
	request   string
}

func openHistoryFile(filename string) *historyFile {
	glog.Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		glog.Fatal(err)
	}
	writer := csv.NewWriter(file)
	return &historyFile{writer, time.Now(), ""}
}

func (sf *historyFile) startRequest(request string) {
	sf.request = request
	sf.startTime = time.Now()
}

func (sf *historyFile) stopRequest(response string) {
	err := sf.w.Write([]string{
		strconv.FormatInt(sf.startTime.UnixNano(), 10),
		strconv.FormatInt(time.Now().UnixNano(), 10),
		sf.request,
		response})
	if err != nil {
		glog.Fatal(err)
	}
}

func (sf *historyFile) closeHistoryFile() {
	sf.w.Flush()
	//TODO:close stats file
}
