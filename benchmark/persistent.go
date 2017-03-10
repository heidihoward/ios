package main

import (
	"math/rand"
	"os"
	"fmt"
	"time"
)

func benchmarkDisk(filename string, size int, count int) {
	startTime := time.Now()
	bytes := make([]byte, size)
  file, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
  for i := 0; i < count; i++ {
		rand.Read(bytes)
    file.Write(bytes)
    file.Sync()
  }
	fmt.Printf("%s\n",time.Since(startTime).String())
}


func main() {
	benchmarkDisk("testing.log",10,1000)

}
