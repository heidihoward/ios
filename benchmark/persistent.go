// +build ignore

package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"syscall"
	"time"
)

func benchmarkDisk(filename string, size int, count int) {
	startTime := time.Now()
	bytes := make([]byte, size)
	rand.Read(bytes)
	fd, err := syscall.Open(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	err = syscall.Fallocate(fd, 0, 0, int64(size*count))
	if err != nil {
		panic(err)
	}
	for i := 0; i < count; i++ {
		n, err := syscall.Write(fd, bytes)
		if n != size {
			panic(errors.New("Short write"))
		}
		if err != nil {
			panic(err)
		}
		err = syscall.Fdatasync(fd)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("%s\n", time.Since(startTime).String())
	syscall.Close(fd)
}

func main() {
	size := flag.Int("size", 1, "number of bytes to append to file")
	count := flag.Int("count", 1000, "number of appends")
	file := flag.String("file", "bench.dat", "file to write to")
	flag.Parse()
	benchmarkDisk(*file, *size, *count)

}
