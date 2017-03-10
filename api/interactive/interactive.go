// Package interactive provides a simple user facing terminal client for Ios.
package interactive

import (
	"bufio"
	"fmt"
	"github.com/golang/glog"
	"os"
	"strings"
)

type Interative bufio.Reader

func Create() *Interative {
	fmt.Print(`Starting Ios client in interactive mode.

The following commands are available:
	get [key]: to return the value of a given key
	exists [key]: to test if a given key is present
	update [key] [value]: to set the value of a given key, if key already exists then overwrite
	delete [key]: to remove a key value pair if present
	count: to return the number of keys
	print: to return all key value pairs
`)
	return (*Interative)(bufio.NewReader(os.Stdin))

}

func (i *Interative) Next() (string, bool, bool) {
	b := (*bufio.Reader)(i)
	fmt.Print("Enter command: ")
	text, err := b.ReadString('\n')
	if err != nil {
		glog.Fatal(err)
	}
	text = strings.Trim(text, "\n")
	glog.V(1).Info("User entered", text)
	return text, true, true
}

func (_ *Interative) Return(str string) {
	// , time time.Duration  "request took ", time
	fmt.Print(str + "\n")
}
