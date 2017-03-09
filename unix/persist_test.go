package unix


import (
	"testing"
	"math/rand"
)

func benchmarkDisk(n int, b *testing.B) {
	bytes := make([]byte, n)
  file := openFile("testing.bench")
  for i := 0; i < b.N; i++ {
		rand.Read(bytes)
    file.Fd.Write(bytes)
    file.Fd.Sync()
  }
}

func BenchmarkDisk10Bytes(b *testing.B) { benchmarkDisk(10,b) }
func BenchmarkDisk100Bytes(b *testing.B) { benchmarkDisk(100,b) }
func BenchmarkDisk1000Bytes(b *testing.B) { benchmarkDisk(1000,b) }
