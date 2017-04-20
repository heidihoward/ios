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

func openHistoryFile(filename string) (*historyFile, error) {
	glog.Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	writer := csv.NewWriter(file)
	return &historyFile{writer, time.Now(), ""}, nil
}

func (sf *historyFile) startRequest(request string) {
	sf.request = request
	sf.startTime = time.Now()
}

func (sf *historyFile) stopRequest(response string) error {
	return sf.w.Write([]string{
		strconv.FormatInt(sf.startTime.UnixNano(), 10),
		strconv.FormatInt(time.Now().UnixNano(), 10),
		sf.request,
		response})
}

func (sf *historyFile) closeHistoryFile() {
	sf.w.Flush()
	//TODO:close stats file
}
