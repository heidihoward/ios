package msgs

import (
	"github.com/golang/glog"
)

type Storage interface {
	PersistView(view int) error
	PersistLogUpdate(logUpdate LogUpdate) error
	PersistSnapshot(index int, snap []byte) error
}

type ExternalStorage struct {
	ViewPersist      chan int
	ViewPersistFsync chan int
	LogPersist       chan LogUpdate
	LogPersistFsync  chan LogUpdate
}

func MakeExternalStorage() *ExternalStorage {
	buf := 10
	s := ExternalStorage{
		ViewPersist:      make(chan int, buf),
		ViewPersistFsync: make(chan int, buf),
		LogPersist:       make(chan LogUpdate, buf),
		LogPersistFsync:  make(chan LogUpdate, buf)}
	return &s
}

func (s *ExternalStorage) PersistView(view int) error {
	s.ViewPersist <- view
	<-s.ViewPersistFsync
	return nil
}

func (s *ExternalStorage) PersistLogUpdate(logUpdate LogUpdate) error {
	s.LogPersist <- logUpdate
	<-s.LogPersistFsync
	return nil
}

func (s *ExternalStorage) PersistSnapshot(index int, snap []byte) error {
	// TODO: complete stub
	return nil
}

type DummyStorage struct{}

func MakeDummyStorage() *DummyStorage {
	return &DummyStorage{}
}

func (_ *DummyStorage) PersistView(view int) error {
	glog.V(1).Info("Updating view to ", view)
	return nil
}

func (_ *DummyStorage) PersistLogUpdate(log LogUpdate) error {
	glog.V(1).Info("Updating log with ", log)
	return nil
}

func (_ *DummyStorage) PersistSnapshot(index int, snap []byte) error {
	glog.V(1).Info("Updating snap with ", index, snap)
	return nil
}
