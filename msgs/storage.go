package msgs

import (
	"github.com/golang/glog"
)

type Storage interface {
  PersistView(view int)
  PersistLogUpdate(logUpdate LogUpdate)
  PersistSnapshot(snap Snapshot)
}

type ExternalStorage struct {
  ViewPersist            chan int
  ViewPersistFsync       chan int
  LogPersist             chan LogUpdate
  LogPersistFsync        chan LogUpdate
  SnapshotPersist        chan Snapshot
}

func MakeExternalStorage() *ExternalStorage {
  buf := 10
	s := ExternalStorage{
		ViewPersist:            make(chan int, buf),
		ViewPersistFsync:       make(chan int, buf),
		LogPersist:             make(chan LogUpdate, buf),
		LogPersistFsync:        make(chan LogUpdate, buf),
		SnapshotPersist:        make(chan Snapshot, buf)}
	return &s
}

func (s *ExternalStorage) PersistView(view int) {
  s.ViewPersist <- view
  <-s.ViewPersistFsync
}

func (s *ExternalStorage) PersistLogUpdate(logUpdate LogUpdate) {
  s.LogPersist <- logUpdate
  <-s.LogPersistFsync
}

func (s *ExternalStorage) PersistSnapshot(snap Snapshot) {
  s.SnapshotPersist <- snap
}

type DummyStorage struct{}

func MakeDummyStorage() *DummyStorage {
  return &DummyStorage{}
}

func (_ *DummyStorage) PersistView(view int) {
  glog.V(1).Info("Updating view to ", view)
}

func (_ *DummyStorage) PersistLogUpdate(log LogUpdate) {
  glog.V(1).Info("Updating log with ", log)
}

func (_ *DummyStorage) PersistSnapshot(snap Snapshot) {
  glog.V(1).Info("Updating snap with ", snap)
}
