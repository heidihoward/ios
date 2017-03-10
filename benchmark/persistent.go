package main

import (
	"math/rand"
	"os"
	"fmt"
	"flag"
	"time"
)

func benchmarkDisk(filename string, size int, count int) {
	startTime := time.Now()
	bytes := make([]byte, size)
	rand.Read(bytes)
  file, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
  for i := 0; i < count; i++ {
    file.Write(bytes)
    file.Sync()
  }
	fmt.Printf("%s\n",time.Since(startTime).String())
}


func main() {
	size := flag.Int("size", 1, "number of bytes to append to file")
	count := flag.Int("count", 1000, "number of appends")
	file := flag.String("file", "bench.dat", "file to write to")
	flag.Parse()
	benchmarkDisk(*file,*size,*count)

}
