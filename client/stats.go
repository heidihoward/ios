package main

import (
	"encoding/csv"
	"github.com/golang/glog"
	"os"
	"strconv"
	"time"
)

type statsFile struct {
	w         *csv.Writer
	startTime time.Time
	requestID int
}

func OpenStatsFile(filename string) *statsFile {
	glog.Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		glog.Fatal(err)
	}
	writer := csv.NewWriter(file)
	return &statsFile{writer, time.Now(), 1}
}

func (sf *statsFile) StartRequest(requestID int) {
	sf.requestID = requestID
	sf.startTime = time.Now()
}

func (sf *statsFile) StopRequest(tries int) {
	latency := strconv.FormatInt(time.Since(sf.startTime).Nanoseconds(), 10)
	err := sf.w.Write([]string{strconv.FormatInt(sf.startTime.UnixNano(), 10), strconv.Itoa(sf.requestID), latency, strconv.Itoa(tries)})
	if err != nil {
		glog.Fatal(err)
	}
}

func (sf *statsFile) CloseStatsFile() {
	sf.w.Flush()
	//TODO:close stats file
}
