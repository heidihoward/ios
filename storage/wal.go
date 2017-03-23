package storage

import (
	"github.com/golang/glog"
	"time"
)

func slowDiskWarning(startTime time.Time, n int) {
	delay := time.Since(startTime)
	if delay > time.Millisecond {
		glog.Info("Slow disk warning - ", n, " bytes written & synced to persistent log in ", delay.String())
	}
}
